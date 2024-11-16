package config

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
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
		} else {
			return nil, fmt.Errorf("config file found but another error was produced: %w", err)
		}
	}
	return raw, nil
}

func LocalDefault() *viper.Viper {
	cfg, err := readLocalConfig("config", ".")
	if err != nil {
		slog.Error(err.Error())
		return nil
	}
	return cfg
}

func LocalGuild(guildID snowflake.ID) *viper.Viper {
	cfg, err := readLocalConfig(("config." + guildID.String()), ".")
	if err != nil {
		slog.Error(err.Error())
		return nil
	}
	return cfg
}

// `Init` initializes the configuration. It should be called once at the start of the program.
func Init(botName string) {
	sync.OnceFunc(func() {
		var err error

		// Fetching the global configuration using the default local file
		cfg, err := readLocalConfig("config", ".")
		if err != nil {
			slog.Error(err.Error())
			os.Exit(-1)
		}

		// Instantiating the global, log and bot configurations
		Global, err = NewGlobalConfig(cfg)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(-1)
		}

		Log, err = NewLogConfig(cfg)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(-1)
		}

		Bot, err = NewBotConfig(cfg, botName)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(-1)
		}

		// Empty guilds configuration for future use
		Guilds = &GuildsConfigs{}

	})()
}

// `InitGuilds` initializes the guilds configuration. It should be called after `Init`.
func InitGuilds(client rest.Rest) {
	sync.OnceFunc(func() {

		// Fetching each guild the bot is in
		botGuilds, err := client.GetCurrentUserGuilds("", 0, 0, 0, true)
		if err != nil {
			slog.Error("Failed to fetch bot guilds", slog.Any("err", err))
			os.Exit(-1)
		}

		if len(botGuilds) == 0 {
			return
		}

		for _, guild := range botGuilds {
			// Fetching the guild configuration using the default local file if any, ignoring errors
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

	})()
}

type config interface {
	validate() error
}
