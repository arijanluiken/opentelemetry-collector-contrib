// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package timeseriesanomalyprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/timeseriesanomalyprocessor"

import (
	"errors"
	"time"

	"go.opentelemetry.io/collector/component"
)

// Config holds configuration for timeseries anomaly processor.
// It uses online mean/variance with a sliding window to compute z-scores per metric series.
// A series key is derived from resource + scope + metric name + datapoint attributes.
// If z-score exceeds threshold, datapoint is flagged and enriched with attributes.
// Mode "filter" drops non-anomalies, "enrich" only adds attributes, "both" does both per signal semantics.
type Config struct {
	// WindowSize is number of recent points to keep per series for mean/std.
	WindowSize int `mapstructure:"window_size"`
	// ZScoreThreshold triggers anomaly when abs(z) >= threshold.
	ZScoreThreshold float64 `mapstructure:"zscore_threshold"`
	// Mode: enrich|filter|both
	Mode string `mapstructure:"mode"`
	// ScoreAttribute name for numeric score attribute.
	ScoreAttribute string `mapstructure:"score_attribute"`
	// FlagAttribute name for boolean flag attribute.
	FlagAttribute string `mapstructure:"flag_attribute"`
	// MaxSeries limits number of concurrent time series tracked.
	MaxSeries int `mapstructure:"max_series"`
	// GCInterval periodically prunes idle series.
	GCInterval time.Duration `mapstructure:"gc_interval"`
	// IdleTTL time after which a series is considered stale.
	IdleTTL time.Duration `mapstructure:"idle_ttl"`
}

func createDefaultConfig() component.Config {
	return &Config{
		WindowSize:      50,
		ZScoreThreshold: 3.0,
		Mode:            "enrich",
		ScoreAttribute:  "anomaly.score",
		FlagAttribute:   "anomaly.flag",
		MaxSeries:       10000,
		GCInterval:      5 * time.Minute,
		IdleTTL:         15 * time.Minute,
	}
}

func (c *Config) Validate() error {
	if c.WindowSize <= 1 {
		return errors.New("window_size must be > 1")
	}
	if c.ZScoreThreshold <= 0 {
		return errors.New("zscore_threshold must be > 0")
	}
	switch c.Mode {
	case "enrich", "filter", "both":
	default:
		return errors.New("mode must be 'enrich', 'filter', or 'both'")
	}
	if c.ScoreAttribute == "" || c.FlagAttribute == "" {
		return errors.New("score_attribute and flag_attribute must be set")
	}
	if c.MaxSeries < 0 {
		return errors.New("max_series must be >= 0")
	}
	if c.GCInterval <= 0 || c.IdleTTL <= 0 {
		return errors.New("gc_interval and idle_ttl must be positive")
	}
	return nil
}
