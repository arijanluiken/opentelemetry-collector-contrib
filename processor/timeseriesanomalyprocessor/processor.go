// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package timeseriesanomalyprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/timeseriesanomalyprocessor"

import (
	"context"
	"math"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
)

// processor implements metrics-only anomaly detection using z-scores per series.
// It maintains a sliding window of values for each distinct series key and computes
// mean/std with Welford's algorithm. Optionally filters non-anomalies.

type processorImpl struct {
	cfg *Config

	mu     sync.Mutex
	series map[string]*seriesState

	stopCh chan struct{}
}

type seriesState struct {
	values []float64
	mean   float64
	m2     float64 // sum of squares of differences from the current mean
	count  int
	last   time.Time
}

func newProcessor(cfg *Config, set processor.Settings, next consumer.Metrics) processor.Metrics {
	p := &processorImpl{
		cfg:    cfg,
		series: make(map[string]*seriesState),
		stopCh: make(chan struct{}),
	}
	return &metricsProc{impl: p, next: next}
}

// metricsProc wires to collector interfaces.

type metricsProc struct {
	impl *processorImpl
	next consumer.Metrics
}

func (mp *metricsProc) Start(ctx context.Context, _ component.Host) error {
	// Start a GC goroutine
	ticker := time.NewTicker(mp.impl.cfg.GCInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				mp.impl.gc()
			case <-mp.impl.stopCh:
				ticker.Stop()
				return
			}
		}
	}()
	return nil
}

func (mp *metricsProc) Shutdown(context.Context) error {
	close(mp.impl.stopCh)
	return nil
}

func (*metricsProc) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func (mp *metricsProc) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Iterate over metrics and process number data points (gauge/sum)
	rmSlice := md.ResourceMetrics()
	for i := 0; i < rmSlice.Len(); i++ {
		rm := rmSlice.At(i)
		resAttrs := rm.Resource().Attributes()
		smSlice := rm.ScopeMetrics()
		for j := 0; j < smSlice.Len(); j++ {
			sm := smSlice.At(j)
			mSlice := sm.Metrics()
			for k := 0; k < mSlice.Len(); k++ {
				m := mSlice.At(k)
				switch m.Type() {
				case pmetric.MetricTypeGauge:
					g := m.Gauge()
					newDPs := pmetric.NewNumberDataPointSlice()
					for idx := 0; idx < g.DataPoints().Len(); idx++ {
						dp := g.DataPoints().At(idx)
						keep := mp.impl.processPoint(m, dp, resAttrs)
						if keep {
							ndp := newDPs.AppendEmpty()
							dp.CopyTo(ndp)
						}
					}
					g.DataPoints().RemoveIf(func(pmetric.NumberDataPoint) bool { return true })
					for idx := 0; idx < newDPs.Len(); idx++ {
						ndp := g.DataPoints().AppendEmpty()
						newDPs.At(idx).CopyTo(ndp)
					}
				case pmetric.MetricTypeSum:
					s := m.Sum()
					newDPs := pmetric.NewNumberDataPointSlice()
					for idx := 0; idx < s.DataPoints().Len(); idx++ {
						dp := s.DataPoints().At(idx)
						keep := mp.impl.processPoint(m, dp, resAttrs)
						if keep {
							ndp := newDPs.AppendEmpty()
							dp.CopyTo(ndp)
						}
					}
					s.DataPoints().RemoveIf(func(pmetric.NumberDataPoint) bool { return true })
					for idx := 0; idx < newDPs.Len(); idx++ {
						ndp := s.DataPoints().AppendEmpty()
						newDPs.At(idx).CopyTo(ndp)
					}
				}
			}
		}
	}
	return mp.next.ConsumeMetrics(ctx, md)
}

func (p *processorImpl) processPoint(m pmetric.Metric, dp pmetric.NumberDataPoint, resAttrs pcommon.Map) bool {
	// Derive a series key
	key := buildSeriesKey(m, dp, resAttrs)

	// Only process double values; convert ints to double
	var value float64
	switch dp.ValueType() {
	case pmetric.NumberDataPointValueTypeInt:
		value = float64(dp.IntValue())
	case pmetric.NumberDataPointValueTypeDouble:
		value = dp.DoubleValue()
	default:
		return true // passthrough
	}

	z, isAnom := p.updateAndScore(key, value)

	// Enrich
	dp.Attributes().PutDouble(p.cfg.ScoreAttribute, z)
	dp.Attributes().PutBool(p.cfg.FlagAttribute, isAnom)

	// Filter mode handling
	switch p.cfg.Mode {
	case "filter":
		return isAnom
	case "both":
		// Keep anomalies only but also enrich
		return isAnom
	default: // enrich only
		return true
	}
}

func (p *processorImpl) updateAndScore(key string, x float64) (float64, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	st, ok := p.series[key]
	if !ok {
		if p.cfg.MaxSeries > 0 && len(p.series) >= p.cfg.MaxSeries {
			// refuse to allocate new series; treat as non-anomaly passthrough
			return 0, false
		}
		st = &seriesState{}
		p.series[key] = st
	}

	// Compute z-score based on existing window (exclude current x)
	mean := 0.0
	m2 := 0.0
	count := 0
	for _, v := range st.values {
		count++
		delta := v - mean
		mean += delta / float64(count)
		m2 += delta * (v - mean)
	}
	variance := 0.0
	if count > 1 {
		variance = m2 / float64(count-1)
	}
	stddev := math.Sqrt(variance)
	z := 0.0
	if stddev > 0 {
		z = math.Abs((x - mean) / stddev)
	}
	isAnom := z >= p.cfg.ZScoreThreshold

	// Now append value and update series state for future points
	st.values = append(st.values, x)
	st.last = time.Now()
	if len(st.values) > p.cfg.WindowSize {
		st.values = st.values[len(st.values)-p.cfg.WindowSize:]
	}
	// Update cached stats for debugging (not used functionally elsewhere)
	st.mean, st.m2, st.count = 0, 0, 0
	for _, v := range st.values {
		st.count++
		delta := v - st.mean
		st.mean += delta / float64(st.count)
		st.m2 += delta * (v - st.mean)
	}
	return z, isAnom
}

func (p *processorImpl) gc() {
	p.mu.Lock()
	defer p.mu.Unlock()
	deadline := time.Now().Add(-p.cfg.IdleTTL)
	for k, st := range p.series {
		if st.last.Before(deadline) {
			delete(p.series, k)
		}
	}
}

func buildSeriesKey(m pmetric.Metric, dp pmetric.NumberDataPoint, res pcommon.Map) string {
	b := make([]byte, 0, 256)
	b = append(b, m.Name()...)
	b = append(b, '\x1f')
	// resource attrs
	res.Range(func(k string, v pcommon.Value) bool {
		b = append(b, k...)
		b = append(b, '=')
		b = append(b, v.AsString()...)
		b = append(b, '\x1e')
		return true
	})
	b = append(b, '\x1f')
	// datapoint attrs
	dp.Attributes().Range(func(k string, v pcommon.Value) bool {
		b = append(b, k...)
		b = append(b, '=')
		b = append(b, v.AsString()...)
		b = append(b, '\x1e')
		return true
	})
	return string(b)
}
