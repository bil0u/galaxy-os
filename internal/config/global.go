package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// `GlobalConfig` holds the configuration for the global settings
type GlobalConfig struct {
	Development bool   `mapstructure:"development"`
	Version     string `mapstructure:"-"`
	Commit      string `mapstructure:"-"`
}

// `validate` validates a global configuration
func (config GlobalConfig) validate() error {
	return nil
}

// `NewGlobalConfig` creates a new global configuration object from a viper object
func NewGlobalConfig(raw *viper.Viper) (*GlobalConfig, error) {
	newCfg := &GlobalConfig{}

	err := raw.Unmarshal(newCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal global configuration: %w", err)
	}

	if newCfg.Development && newCfg.Version == "" {
		newCfg.Version = "development"
	}

	return newCfg, newCfg.validate()
}
