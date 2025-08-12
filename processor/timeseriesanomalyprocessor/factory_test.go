// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package timeseriesanomalyprocessor

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/processor/processortest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/timeseriesanomalyprocessor/internal/metadata"
)

func TestFactory_Basics(t *testing.T) {
	f := NewFactory()
	require.Equal(t, metadata.Type, f.Type())
	cfg := f.CreateDefaultConfig()
	require.NotNil(t, cfg)
	_, ok := cfg.(*Config)
	require.True(t, ok)
}

func TestFactory_CreateAndLifecycle(t *testing.T) {
	f := NewFactory()
	cfg := f.CreateDefaultConfig()
	settings := processortest.NewNopSettings(metadata.Type)
	next := consumertest.NewNop()
	p, err := f.CreateMetrics(context.Background(), settings, cfg, next)
	require.NoError(t, err)
	require.NotNil(t, p)

	require.NoError(t, p.Start(context.Background(), componenttest.NewNopHost()))
	require.NoError(t, p.Shutdown(context.Background()))
}
