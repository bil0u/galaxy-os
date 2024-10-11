package sdk

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/bil0u/galaxy-os/sdk/enums"
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
	if len(cfg.Guilds) == 0 {
		return fmt.Errorf("at least one guild must be defined")
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
	Log    LogConfig                    `toml:"log"`
	Bot    BotConfig                    `toml:"bot"`
	Guilds map[snowflake.ID]GuildConfig `toml:"guilds"`
}

type LogConfig struct {
	Level     slog.Level `toml:"level"`
	Format    string     `toml:"format"`
	AddSource bool       `toml:"add_source"`
}

type BotConfig struct {
	Token         string       `toml:"token"`
	ApplicationID snowflake.ID `toml:"application_id"`
}

type GuildConfig struct {
	DevGuild bool           `toml:"dev_guild"`
	Timezone string         `toml:"timezone"`
	BotRoles []snowflake.ID `toml:"bot_roles"`
}

// GetGuildRoles returns the roles for a specific guild
func (c Config) GetGuildRoles(guildID snowflake.ID) []enums.RoleEnum {
	if guild, ok := c.Guilds[guildID]; ok {
		var roles []enums.RoleEnum
		for _, roleID := range guild.BotRoles {
			// Getting role from RoleMap using roleID
			if role := enums.GetRoleEnum(roleID); role.IsValid() {
				roles = append(roles, role)
			}

		}
		return roles
	}
	return nil
}

// GetGuilds returns the guilds that the bot is in
func (c Config) GetGuildsIDs() []snowflake.ID {
	var guilds []snowflake.ID
	for guildID := range c.Guilds {
		guilds = append(guilds, guildID)
	}
	return guilds
}

// GetDevGuilds returns the dev guilds that the bot is in
func (c Config) GetDevGuildsIDs() []snowflake.ID {
	var devGuilds []snowflake.ID
	for guildID, guild := range c.Guilds {
		if guild.DevGuild {
			devGuilds = append(devGuilds, guildID)
		}
	}
	return devGuilds
}

// GetGuildTimezone returns the timezone for a specific guild
func (c Config) GetGuildConfig(guildID snowflake.ID) *GuildConfig {
	if guild, ok := c.Guilds[guildID]; ok {
		return &guild
	}
	return nil
}
