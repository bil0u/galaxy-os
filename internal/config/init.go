package config

import (
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/rest"
	"github.com/spf13/viper"
)

var (
	Global *GlobalConfig
	Log    *LogConfig
	Bot    *BotConfig
	Guilds *GuildsConfigs
)

func readLocalConfig(filename, path string) (*viper.Viper, error) {
	raw := viper.New()
	raw.SetConfigName(filename)
	raw.AddConfigPath(path)
	if err := raw.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("config file not found: %w", err)
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	return raw, nil
}

// Init initializes the configuration. It should be called once at the start of the program.
func Init(botName string) error {
	cfg, err := readLocalConfig("config", ".")
	if err != nil {
		return err
	}

	Global, err = NewGlobalConfig(cfg)
	if err != nil {
		return err
	}

	Log, err = NewLogConfig(cfg)
	if err != nil {
		return err
	}

	Bot, err = NewBotConfig(cfg, botName)
	if err != nil {
		return err
	}

	Guilds = &GuildsConfigs{}
	return nil
}

// InitGuilds initializes the guilds configuration. It should be called after Init.
func InitGuilds(client rest.Rest) error {
	botGuilds, err := client.GetCurrentUserGuilds("", 0, 0, 0, true)
	if err != nil {
		return fmt.Errorf("fetching bot guilds: %w", err)
	}

	if len(botGuilds) == 0 {
		return nil
	}

	for _, guild := range botGuilds {
		cfg, err := readLocalConfig(("config." + guild.ID.String()), ".")
		if err != nil {
			cfg = nil
			slog.Error(err.Error())
		}
		config, _ := NewGuildConfig(cfg, Bot.Name)

		err = Guilds.Set(guild.ID, config)
		if err != nil {
			slog.Error(err.Error())
		}
	}

	return nil
}

type config interface {
	validate() error
}
