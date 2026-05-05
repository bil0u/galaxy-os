package config

import (
	"fmt"
	"log/slog"

	"github.com/spf13/viper"
)

// Log holds the configuration for the logging system.
type Log struct {
	Level     slog.Level `mapstructure:"level"`
	Format    string     `mapstructure:"format"`
	AddSource bool       `mapstructure:"add_source"`
}

// validate validates the log configuration
func (config Log) validate() error {
	if config.Format == "" {
		return fmt.Errorf("log.format must be provided")
	}
	return nil
}

// NewLog creates a new Log from a viper object.
func NewLog(raw *viper.Viper) (*Log, error) {
	newCfg := Log{
		Level:     slog.LevelInfo,
		Format:    "json",
		AddSource: false,
	}

	logSub := raw.Sub("log")
	if logSub == nil {
		return &newCfg, newCfg.validate()
	}
	if err := logSub.Unmarshal(&newCfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal log configuration: %w", err)
	}

	return &newCfg, newCfg.validate()
}
