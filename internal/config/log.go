package config

import (
	"fmt"
	"log/slog"

	"github.com/spf13/viper"
)

// LogConfig holds the configuration for the logging system
type LogConfig struct {
	Level     slog.Level `mapstructure:"level"`
	Format    string     `mapstructure:"format"`
	AddSource bool       `mapstructure:"add_source"`
}

// validate validates the log configuration
func (config LogConfig) validate() error {
	if config.Format == "" {
		return fmt.Errorf("log.format must be provided")
	}
	return nil
}

// NewLogConfig creates a new LogConfig object from a file
func NewLogConfig(raw *viper.Viper) (*LogConfig, error) {
	newCfg := LogConfig{
		Level:     slog.LevelInfo,
		Format:    "json",
		AddSource: false,
	}

	err := raw.Sub("log").Unmarshal(&newCfg)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal log configuration: %w", err)
	}

	return &newCfg, newCfg.validate()
}
