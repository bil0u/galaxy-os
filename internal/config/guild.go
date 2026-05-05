package config

import (
	"errors"
	"fmt"

	"github.com/disgoorg/snowflake/v2"
	"github.com/spf13/viper"
)

// Guild holds the configuration for a specific guild.
type Guild struct {
	DevGuild     bool         `mapstructure:"dev_guild"`
	Timezone     string       `mapstructure:"timezone"`
	FeaturesDefs *viper.Viper `mapstructure:"-"`
}

// `validate` validates the guild configuration.
func (config Guild) validate() error {
	var errs []error

	if config.Timezone == "" {
		errs = append(errs, fmt.Errorf("guild.timezone must be provided"))
	}

	return errors.Join(errs...)
}

// NewGuild creates a new Guild from a viper configuration.
func NewGuild(cfg *viper.Viper, botName string) (*Guild, error) {
	newCfg := &Guild{}

	if cfg != nil {
		if err := cfg.Unmarshal(newCfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal guild configuration: %w", err)
		}
		newCfg.FeaturesDefs = cfg.Sub("features." + botName)
	}

	return newCfg, newCfg.validate()
}

// --------------------------------------------------------------------

// GuildMap is a wrapper type for multiple guild configurations.
type GuildMap map[snowflake.ID]*Guild

// `Count` returns the number of guild configurations.
func (gc GuildMap) Count() int {
	return len(gc)
}

// `All` returns all guild configurations.
func (gc GuildMap) All() map[snowflake.ID]*Guild {
	return gc
}

// `Dev` returns all guild configurations for development guilds.
func (gc GuildMap) Dev() map[snowflake.ID]*Guild {
	devGuilds := make(map[snowflake.ID]*Guild)
	for id, guildConfig := range gc {
		if guildConfig.DevGuild {
			devGuilds[id] = guildConfig
		}
	}
	return devGuilds
}

// `GetIDs` returns a list of guild IDs for all guilds in the configuration.
func (gc GuildMap) IDs(devOnly bool) []snowflake.ID {
	guildsIDs := make([]snowflake.ID, 0)
	for id, guildConfig := range gc {
		if !devOnly || guildConfig.DevGuild {
			guildsIDs = append(guildsIDs, id)
		}
	}
	return guildsIDs
}

// `Get` returns a list of guild IDs for all guilds in the configuration.
func (gc GuildMap) Get(guildID snowflake.ID) *Guild {
	return gc[guildID]
}

// `Set` sets a guild configuration for a specific guild.
func (gc *GuildMap) Set(guildID snowflake.ID, config *Guild) error {
	_, ok := (*gc)[guildID]
	if ok {
		return fmt.Errorf("guild configuration already exists")
	}
	(*gc)[guildID] = config
	return nil
}

// `Override` overrides a guild configuration for a specific guild.
func (gc *GuildMap) Override(guildID snowflake.ID, config *Guild) {
	(*gc)[guildID] = config
}
