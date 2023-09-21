// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package x509processor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/x509processor"

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component/componenttest"
)

func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	assert.NoError(t, componenttest.CheckConfigStruct(cfg))
	assert.NotNil(t, cfg)
}

func TestCreateProcessor(t *testing.T) {
}
