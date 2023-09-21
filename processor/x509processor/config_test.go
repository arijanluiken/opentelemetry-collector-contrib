// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package x509processor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/x509processor"

import (
	"path/filepath"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/x509processor/internal/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id       component.ID
		expected component.Config
		valid    bool
	}{
		{
			id: component.NewIDWithName(metadata.Type, ""),
			expected: &Config{
				KeyFile: "key.pem",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
			require.NoError(t, err)

			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, component.UnmarshalConfig(sub, cfg))

			if tt.valid {
				assert.NoError(t, component.ValidateConfig(cfg))
			} else {
				assert.Error(t, component.ValidateConfig(cfg))
			}
		})
	}
}
