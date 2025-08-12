// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package timeseriesanomalyprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor/processortest"
)

func TestProcessor_FlagsAnomaly(t *testing.T) {
	cfg := &Config{
		WindowSize:      5,
		ZScoreThreshold: 2.0,
		Mode:            "both",
		ScoreAttribute:  "anomaly.score",
		FlagAttribute:   "anomaly.flag",
		MaxSeries:       100,
		GCInterval:      60_000_000_000,  // 60s
		IdleTTL:         300_000_000_000, // 5m
	}
	require.NoError(t, cfg.Validate())

	p := newProcessor(cfg, processortest.NewNopSettings(component.MustNewType("timeseriesanomaly")), consumertest.NewNop()).(*metricsProc)

	// Build a metric with mostly stable points then an outlier
	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()
	m := sm.Metrics().AppendEmpty()
	m.SetName("load")

	g := m.SetEmptyGauge()
	for _, v := range []float64{10, 10, 11, 9, 10, 30} { // 30 should be an outlier
		dp := g.DataPoints().AppendEmpty()
		dp.SetDoubleValue(v)
	}

	err := p.ConsumeMetrics(context.Background(), md)
	require.NoError(t, err)

	// In both mode, only anomalies should remain, expect at least one
	g2 := sm.Metrics().At(0).Gauge()
	require.GreaterOrEqual(t, g2.DataPoints().Len(), 1)
	// Ensure the last point (30) is present and flagged
	var found bool
	for i := 0; i < g2.DataPoints().Len(); i++ {
		dp := g2.DataPoints().At(i)
		if dp.ValueType() == pmetric.NumberDataPointValueTypeDouble && dp.DoubleValue() == 30 {
			if _, ok := dp.Attributes().Get("anomaly.flag"); ok {
				found = true
				break
			}
		}
	}
	require.True(t, found)
}
