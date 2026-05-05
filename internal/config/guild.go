package config

import (
	"fmt"

	"github.com/bil0u/galaxy-os/internal/utils"
	"github.com/disgoorg/snowflake/v2"
	"github.com/spf13/viper"
)

// `GuildConfig` holds the configuration for a specific guild.
type GuildConfig struct {
	DevGuild     bool         `mapstructure:"dev_guild"`
	Timezone     string       `mapstructure:"timezone"`
	FeaturesDefs *viper.Viper `mapstructure:"-"`
}

// `validate` validates the guild configuration.
func (config GuildConfig) validate() error {

	errs := utils.ManyErrors{}

	if config.Timezone == "" {
		errs.Add(fmt.Errorf("guild.timezone must be provided"))
	}

	return errs.ToError()
}

// `NewGuildConfig` creates a new GuildConfig object from a viper configuration.
func NewGuildConfig(cfg *viper.Viper, botName string) (*GuildConfig, error) {
	newCfg := &GuildConfig{}

	if cfg != nil {
		if err := cfg.Unmarshal(newCfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal guild configuration: %w", err)
		}
		newCfg.FeaturesDefs = cfg.Sub("features." + botName)
	}

	return newCfg, newCfg.validate()
}

// --------------------------------------------------------------------

// `GuildsConfigs` is a wrapper type for multiple guild configurations.
type GuildsConfigs map[snowflake.ID]*GuildConfig

// `Count` returns the number of guild configurations.
func (gc GuildsConfigs) Count() int {
	return len(gc)
}

// `All` returns all guild configurations.
func (gc GuildsConfigs) All() map[snowflake.ID]*GuildConfig {
	return gc
}

// `Dev` returns all guild configurations for development guilds.
func (gc GuildsConfigs) Dev() map[snowflake.ID]*GuildConfig {
	devGuilds := make(map[snowflake.ID]*GuildConfig)
	for id, guildConfig := range gc {
		if guildConfig.DevGuild {
			devGuilds[id] = guildConfig
		}
	}
	return devGuilds
}

// `GetIDs` returns a list of guild IDs for all guilds in the configuration.
func (gc GuildsConfigs) IDs(devOnly bool) []snowflake.ID {
	guildsIDs := make([]snowflake.ID, 0)
	for id, guildConfig := range gc {
		if !devOnly || guildConfig.DevGuild {
			guildsIDs = append(guildsIDs, id)
		}
	}
	return guildsIDs
}

// `Get` returns a list of guild IDs for all guilds in the configuration.
func (gc GuildsConfigs) Get(guildID snowflake.ID) *GuildConfig {
	return gc[guildID]
}

// `Set` sets a guild configuration for a specific guild.
func (gc *GuildsConfigs) Set(guildID snowflake.ID, config *GuildConfig) error {
	_, ok := (*gc)[guildID]
	if ok {
		return fmt.Errorf("guild configuration already exists")
	}
	(*gc)[guildID] = config
	return nil
}

// `Override` overrides a guild configuration for a specific guild.
func (gc *GuildsConfigs) Override(guildID snowflake.ID, config *GuildConfig) {
	(*gc)[guildID] = config
}
