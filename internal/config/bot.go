package config

import (
	"fmt"
	"strings"

	"github.com/bil0u/galaxy-os/internal/utils"
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

// `validate` validates the bot configuration
func (config BotConfig) validate() error {
	errs := utils.ManyErrors{}

	if config.Token == "" {
		errs.Add(fmt.Errorf("bot.token must be provided"))
	}

	if config.ApplicationID == 0 {
		errs.Add(fmt.Errorf("bot.application_id must be provided"))
	}

	if config.BaseURL == "" || config.ClientSecret == "" {
		errs.Add(fmt.Errorf("bot.base_url and bot.client_secret must be provided"))
	}

	return errs.ToError()
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
		err := cfg.Sub("bot." + botName).Unmarshal(&newCfg)

		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal bot configuration: %w", err)
		}
		newCfg.FeaturesDefs = cfg.Sub("features." + botName)
	} else {
		newCfg.FeaturesDefs = viper.New()
	}

	return &newCfg, newCfg.validate()
}
