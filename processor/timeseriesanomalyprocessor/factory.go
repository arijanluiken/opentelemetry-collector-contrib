// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:generate mdatagen metadata.yaml

package timeseriesanomalyprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/timeseriesanomalyprocessor"

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/timeseriesanomalyprocessor/internal/metadata"
)

// NewFactory creates a factory for the timeseries anomaly processor (metrics only).
func NewFactory() processor.Factory {
	return processor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		processor.WithMetrics(createMetricsProcessor, metadata.MetricsStability),
	)
}

func createMetricsProcessor(_ context.Context, set processor.Settings, cfg component.Config, next consumer.Metrics) (processor.Metrics, error) {
	c, ok := cfg.(*Config)
	if !ok {
		return nil, errors.New("invalid config type")
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return newProcessor(c, set, next), nil
}
