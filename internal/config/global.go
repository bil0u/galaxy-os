package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Global holds the configuration for the global settings.
type Global struct {
	Development bool   `mapstructure:"development"`
	Version     string `mapstructure:"-"`
	Commit      string `mapstructure:"-"`
}

// `validate` validates a global configuration
func (config Global) validate() error {
	return nil
}

// NewGlobal creates a new Global from a viper object.
func NewGlobal(raw *viper.Viper) (*Global, error) {
	newCfg := &Global{}

	err := raw.Unmarshal(newCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal global configuration: %w", err)
	}

	if newCfg.Development && newCfg.Version == "" {
		newCfg.Version = "development"
	}

	return newCfg, newCfg.validate()
}
