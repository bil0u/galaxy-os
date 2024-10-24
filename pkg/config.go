package pkg

import (
	"fmt"
	"log/slog"

	"github.com/bil0u/galaxy-os/pkg/utils"
	"github.com/disgoorg/snowflake/v2"
)

var (
	botConfigFormat      = "config.%s.toml"
	guildConfigFormat    = "guild.%s.toml"
	defaultBotConfigFile = fmt.Sprintf(botConfigFormat, "default")
)

// -- GLOBAL CONFIGURATION --

// Top level configuration
type Config struct {
	Development bool
	Log         LogConfig     `toml:"log"`
	Bot         BotConfig     `toml:"bot"`
	Guilds      []GuildConfig `toml:"guilds"`
}

// ValidateConfig validates a Config object
func (cfg Config) Validate() (errs []error) {

	errs = append(errs, cfg.Log.Validate()...)
	errs = append(errs, cfg.Bot.Validate()...)
	for _, guildCfg := range cfg.Guilds {
		errs = append(errs, guildCfg.Validate()...)
	}

	return
}

// NewConfig creates a new configuration object from a bot name and configuration file
func NewConfig(botName, configFile, configDir string) (Config, error) {
	config := Config{}

	defaultConfigPath := fmt.Sprintf("%s/%s", configDir, defaultBotConfigFile)

	// Load generic config file
	err := utils.LoadFromFile(defaultConfigPath, &config)
	if err != nil {
		return config, fmt.Errorf("encountered error while loading default config '%s'", defaultBotConfigFile)
	}

	// If no config file is provided, use the default one that matches the bot name
	if configFile == "" {
		configFile = fmt.Sprintf(botConfigFormat, botName)
	}

	botConfigPath := fmt.Sprintf("%s/%s", configDir, configFile)

	// Load bot specific config using the same logic
	if botConfigPath != defaultConfigPath {
		err = utils.LoadFromFile(botConfigPath, &config)
		if err != nil {
			return config, fmt.Errorf("encountered error while loading bot config '%s'", configFile)
		}
	}

	return config, nil
}

// GetGuildsIDs returns a list of supported guild IDs, optionally filtering by dev guilds
func (c Config) GetGuildsIDs(devOnly bool) []snowflake.ID {
	var guilds []snowflake.ID
	for _, guildCfg := range c.Guilds {
		switch {
		case devOnly && guildCfg.DevGuild:
			guilds = append(guilds, guildCfg.ID)
		case !devOnly:
			guilds = append(guilds, guildCfg.ID)
		}
	}
	return guilds
}

// GetGuildConfig returns the configuration for a specific guild, if it exists
func (c Config) GetGuildConfig(guildID snowflake.ID) (GuildConfig, error) {
	for i, guild := range c.Guilds {
		if guild.ID == guildID {
			return c.Guilds[i], nil
		}
	}
	return GuildConfig{}, fmt.Errorf("guild '%s' not found", guildID)
}

// -- LOG CONFIGURATION --

// LogConfig holds the configuration for the logger
type LogConfig struct {
	Level     slog.Level `toml:"level"`
	Format    string     `toml:"format"`
	AddSource bool       `toml:"add_source"`
}

// Validate validates a LogConfig object
func (cfg LogConfig) Validate() (errs []error) {

	if cfg.Format == "" {
		errs = append(errs, fmt.Errorf("log.format must be provided"))
	}

	return
}

// -- BOT CONFIGURATION --

// BotConfig holds the configuration for the bot
type BotConfig struct {
	Token         string       `toml:"token"`
	ApplicationID snowflake.ID `toml:"application_id"`
	ClientSecret  string       `toml:"client_secret"`
}

func (cfg BotConfig) Validate() (errs []error) {

	if cfg.Token == "" {
		errs = append(errs, fmt.Errorf("bot.token must be provided"))
	}

	if cfg.ApplicationID == 0 {
		errs = append(errs, fmt.Errorf("bot.application_id must be provided"))
	}

	return
}

// -- GUILD CONFIGURATION --

// GuildConfig holds the configuration for a specific guild.
type GuildConfig struct {
	ID       snowflake.ID    `toml:"id"`
	DevGuild bool            `toml:"dev_guild"`
	Timezone string          `toml:"timezone"`
	Features GuildFeatureSet `toml:"features"`
}

func (cfg GuildConfig) Validate() (errs []error) {

	if cfg.ID == 0 {
		errs = append(errs, fmt.Errorf("guild.id must be provided"))
	}

	if cfg.Timezone == "" {
		errs = append(errs, fmt.Errorf("guild.timezone must be provided"))
	}

	if cfg.Features == nil {
		errs = append(errs, fmt.Errorf("guild.features must be defined"))
	}

	return
}

// NewGuildConfig creates a new guild configuration object
// If a guild config directory exists, it will try to load the guild config from it
func NewGuildConfig(guildID snowflake.ID, botName string, allowedFeatures BotFeatureSet) GuildConfig {
	// Default guild config
	guildCfg := GuildConfig{
		ID:       guildID,
		DevGuild: false,
		Timezone: "",
		Features: GuildFeatureSet{},
	}

	guildConfigFile := fmt.Sprintf(guildConfigFormat, guildID)

	utils.LoadFromFile(guildConfigFile, &guildCfg)

	// From the same file, load the features definition
	featuresDefinition := TomlFeaturesDefinition{}
	utils.LoadFromFile(guildConfigFile, &featuresDefinition)

	// Convert the features definition to actual features, filtering by bot name and allowed features
	guildCfg.Features = featuresDefinition.ToFeatures(botName, allowedFeatures)

	return guildCfg
}
