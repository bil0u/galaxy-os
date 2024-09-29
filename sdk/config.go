package sdk

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/disgoorg/snowflake/v2"
	"github.com/pelletier/go-toml/v2"
)

func ValidateConfig(cfg *Config) error {

	if cfg.Bot.Token == "" {
		return fmt.Errorf("token must be provided")
	}
	if cfg.Bot.ApplicationID == 0 {
		return fmt.Errorf("ApplicationID must be provided")
	}
	if len(cfg.Bot.Guilds) == 0 {
		if len(cfg.Bot.DevGuilds) == 0 {
			return fmt.Errorf("at least one guild must be provided in either guilds or dev_guilds")
		}
		cfg.Bot.Guilds = cfg.Bot.DevGuilds
	}
	return nil
}

func LoadConfig(path string, cfg *Config) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return cfg, fmt.Errorf("failed to open config: %w", err)
	}
	if err = toml.NewDecoder(file).Decode(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

type Config struct {
	Log LogConfig `toml:"log"`
	Bot BotConfig `toml:"bot"`
}

type BotConfig struct {
	Token         string                          `toml:"token"`
	ApplicationID snowflake.ID                    `toml:"application_id"`
	DevGuilds     []snowflake.ID                  `toml:"dev_guilds"`
	Guilds        []snowflake.ID                  `toml:"guilds"`
	GuildsRoles   map[snowflake.ID][]snowflake.ID `toml:"guilds_roles"`
}

// GetGuildRoles returns the roles for a specific guild
func (c BotConfig) GetGuildRoles(guildID snowflake.ID) []snowflake.ID {
	if roles, ok := c.GuildsRoles[guildID]; ok {
		return roles
	}
	return nil
}

// GetGuildsToSync returns the guilds to sync
func (c BotConfig) GetGuildsToSync() []snowflake.ID {
	if len(c.DevGuilds) > 0 {
		return c.DevGuilds
	}
	return c.Guilds
}

type LogConfig struct {
	Level     slog.Level `toml:"level"`
	Format    string     `toml:"format"`
	AddSource bool       `toml:"add_source"`
}
