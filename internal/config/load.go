package config

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/rest"
	"github.com/spf13/viper"
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

func mergeSecrets(cfg *viper.Viper, secretsName, path string) {
	secrets, err := readLocalConfig(secretsName, path)
	if err != nil {
		slog.Debug("no secrets file found, continuing without", slog.String("name", secretsName))
		return
	}
	if err := cfg.MergeConfigMap(secrets.AllSettings()); err != nil {
		slog.Error("merging secrets", slog.String("name", secretsName), slog.Any("error", err))
	}
}

// Init initializes the configuration and returns all config values.
func Init(botName string) (*Global, *Log, *Bot, error) {
	cfg, err := readLocalConfig("config", "conf")
	if err != nil {
		return nil, nil, nil, err
	}
	mergeSecrets(cfg, "secrets", "conf")

	globalCfg, err := NewGlobal(cfg)
	if err != nil {
		return nil, nil, nil, err
	}

	logCfg, err := NewLog(cfg)
	if err != nil {
		return nil, nil, nil, err
	}

	botCfg, err := NewBot(cfg, botName)
	if err != nil {
		return nil, nil, nil, err
	}

	return globalCfg, logCfg, botCfg, nil
}

// InitGuilds queries the Discord API for joined guilds and reads per-guild config files.
func InitGuilds(ctx context.Context, client rest.Rest, botName string) (*GuildMap, error) {
	botGuilds, err := client.GetCurrentUserGuilds("", 0, 0, 0, true)
	if err != nil {
		return nil, fmt.Errorf("fetching bot guilds: %w", err)
	}

	guilds := &GuildMap{}

	if len(botGuilds) == 0 {
		return guilds, nil
	}

	for _, guild := range botGuilds {
		cfg, err := readLocalConfig(("config." + guild.ID.String()), "conf")
		if err != nil {
			cfg = nil
			slog.Error(err.Error())
		}
		if cfg != nil {
			mergeSecrets(cfg, "secrets."+guild.ID.String(), "conf")
		}
		config, _ := NewGuild(cfg, botName)

		err = guilds.Set(guild.ID, config)
		if err != nil {
			slog.Error(err.Error())
		}
	}

	return guilds, nil
}

type config interface {
	validate() error
}
