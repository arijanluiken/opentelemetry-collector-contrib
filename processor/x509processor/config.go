// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package x509processor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/x509processor"

import (
	"errors"

	"go.opentelemetry.io/collector/component"
)

type Operation int

const (
	Sign Operation = iota
	Encrypt
)

// Config defines configuration for Resource processor.
type Config struct {
	KeyFile          string   `mapstructure:"keyfile"`
	SignAttribute    []string `mapstructure:"sign"`
	EncryptAttribute []string `mapstructure:"encrypt"`
}

var _ component.Config = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	if len(cfg.KeyFile) == 0 {
		return errors.New("missing required field \"keyfile\"")
	}
	return nil
}
