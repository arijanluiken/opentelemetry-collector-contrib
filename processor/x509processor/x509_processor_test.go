// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package x509processor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/x509processor"

import (
	"context"
	"fmt"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/processor/processortest"
)

func TestX509Processor(t *testing.T) {
	tests := []struct {
		name             string
		config           component.Config
		sourceAttributes map[string]string
	}{
		{
			name: "i3_and_c3_data",
			config: &Config{
				KeyFile:          "key.pem",
				SignAttribute:    []string{"i3_attribute"},
				EncryptAttribute: []string{"c3_attribute"},
			},
			sourceAttributes: map[string]string{
				"i3_attribute": "will be signed",
				"c3_attribute": "will be encrypted",
				"c2_attribute": "i don't need encryption or signing",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewFactory()

			// Test logs consumer
			tln := new(consumertest.LogsSink)
			rlp, err := factory.CreateLogsProcessor(context.Background(), processortest.NewNopCreateSettings(), tt.config, tln)
			require.NoError(t, err)
			assert.True(t, rlp.Capabilities().MutatesData)

			sourceLogData := generateLogData(tt.sourceAttributes)

			fmt.Printf("< %v\n", sourceLogData.ResourceLogs().At(0).Resource().Attributes().AsRaw())

			err = rlp.ConsumeLogs(context.Background(), sourceLogData)
			require.NoError(t, err)
			logs := tln.AllLogs()
			require.Len(t, logs, 1)

			fmt.Printf("> %v\n", logs[0].ResourceLogs().At(0).Resource().Attributes().AsRaw())
		})
	}

}

func generateLogData(attributes map[string]string) plog.Logs {
	ld := testdata.GenerateLogsOneLogRecordNoResource()
	if attributes == nil {
		return ld
	}
	resource := ld.ResourceLogs().At(0).Resource()
	for k, v := range attributes {
		resource.Attributes().PutStr(k, v)
	}
	return ld
}
