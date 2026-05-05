package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/disgoorg/snowflake/v2"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// `BotConfig` holds the configuration for the bot
type BotConfig struct {
	Name          string       `mapstructure:"-"`
	Token         string       `mapstructure:"token"`
	ApplicationID snowflake.ID `mapstructure:"application_id"`
	ClientSecret  string       `mapstructure:"client_secret"`
	BaseURL       string       `mapstructure:"base_url"`
	FeaturesDefs  *viper.Viper `mapstructure:"-"`
}

// `validate` validates the required bot configuration
func (config BotConfig) validate() error {
	var errs []error

	if config.Token == "" {
		errs = append(errs, fmt.Errorf("bot.token must be provided"))
	}

	if config.ApplicationID == 0 {
		errs = append(errs, fmt.Errorf("bot.application_id must be provided"))
	}

	return errors.Join(errs...)
}

// ValidateOAuth checks that OAuth2-specific fields are present.
// Should only be called when --oauth2 is enabled.
func (config BotConfig) ValidateOAuth() error {
	var errs []error

	if config.ClientSecret == "" {
		errs = append(errs, fmt.Errorf("bot.client_secret is required for oauth2"))
	}

	if config.BaseURL == "" {
		errs = append(errs, fmt.Errorf("bot.base_url is required for oauth2"))
	}

	return errors.Join(errs...)
}

// `DisplayName` returns the bot name in a human readable format
func (config BotConfig) DisplayName() string {
	titleCaser := cases.Title(language.English)
	return titleCaser.String(strings.ReplaceAll(config.Name, "_", " "))
}

// `NewBotConfig` creates a new BotConfig object from a viper configuration
func NewBotConfig(cfg *viper.Viper, botName string) (*BotConfig, error) {

	newCfg := BotConfig{
		Name: botName,
	}

	if cfg != nil {
		botSub := cfg.Sub("bot." + botName)
		if botSub == nil {
			return nil, fmt.Errorf("no configuration found for bot %q", botName)
		}
		if err := botSub.Unmarshal(&newCfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal bot configuration: %w", err)
		}
		newCfg.FeaturesDefs = cfg.Sub("features." + botName)
	} else {
		newCfg.FeaturesDefs = viper.New()
	}

	return &newCfg, newCfg.validate()
}
