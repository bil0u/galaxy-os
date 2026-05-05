package features

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/bil0u/galaxy-os/internal/config"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/spf13/viper"
)

// `FeatureManager` manages a feature set.
// It allows to get feature configurations for a specific guild, and to sync commands to the Discord API
// It also allows to setup the features
// It should be created with a feature set, a bot configuration and a guilds configuration using the `NewManager` function
type FeatureManager struct {
	features      FeatureSet
	botConfig     *config.BotConfig
	guildsConfigs *config.GuildsConfigs
}

// `getFeaturesDefs` returns the corresponding features definitions for a specific guild.
// If the guild ID is 0, it returns the bot features
func (fm *FeatureManager) getFeaturesDefs(guildID snowflake.ID) []*Feature {
	if guildID == 0 {
		return fm.features.Bot()
	}
	return fm.features.Guild()
}

// `guildFeaturesRaw` returns the raw features configurations for a specific guild
func (manager *FeatureManager) guildFeaturesRaw(guildID snowflake.ID) *viper.Viper {
	if guildID == 0 {
		return manager.botConfig.FeaturesDefs
	}

	config := manager.guildsConfigs.Get(guildID)
	if config == nil {
		return nil
	}

	return config.FeaturesDefs
}

// `ConfigRaw` returns the raw feature configuration for a specific guild and feature key
func (manager *FeatureManager) ConfigRaw(guildID snowflake.ID, featureKey string) (*viper.Viper, error) {
	slog.Debug(fmt.Sprintf("Getting feature configuration for guild '%d' and feature '%s'", guildID, featureKey))
	guildFeatures := manager.guildFeaturesRaw(guildID)
	if guildFeatures == nil {
		return nil, fmt.Errorf("no feature config exists for guild '%d'", guildID)
	}
	if !guildFeatures.IsSet(featureKey) {
		return nil, fmt.Errorf("no config defined for feature with key '%s' in guild '%d'", featureKey, guildID)
	}

	return guildFeatures.Sub(featureKey), nil
}

// `Config` returns a new feature configuration for a specific guild and feature, loading data from the configuration
func (fm *FeatureManager) Config(guildID snowflake.ID, featureKey string) (FeatureConfig, error) {
	feature := fm.features.WithKey(featureKey)
	if feature == nil {
		return nil, fmt.Errorf("feature not found in the feature set")
	}
	cfg, err := fm.ConfigRaw(guildID, featureKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get feature configuration: %w", err)
	}
	newCfg := reflect.Zero(feature.cfgType).Interface().(FeatureConfig)
	if err := cfg.Unmarshal(&newCfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal feature configuration: %w", err)
	}
	featureConfig := newCfg
	return featureConfig, featureConfig.Validate()
}

// `ConfigFor` tries to get a feature configuration for a specific guild and type
// type should be reflected from a FeatureConfig implementation
func (fm *FeatureManager) ConfigFor(guildID snowflake.ID, cfgType reflect.Type) (FeatureConfig, error) {
	for _, f := range fm.features {
		if f.cfgType == cfgType {
			return fm.Config(guildID, f.Key)
		}
	}
	return nil, fmt.Errorf("feature not found in the feature set")
}

// `SetupFeatures` sets up the features, checking for validity first
func (fm *FeatureManager) SetupFeatures(deps SetupDeps) error {
	var errs []error
	for _, feature := range fm.features {
		if err := feature.IsValid(); err != nil {
			errs = append(errs, err)
			continue
		}
		if err := feature.Setup(deps); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// `SyncCommands` syncs the commands to the Discord API
func (fm *FeatureManager) SyncCommands(client bot.Client, guildsIDs []snowflake.ID) error {

	commandsToSync := []discord.ApplicationCommandCreate{}

	for _, f := range fm.features {
		commandsToSync = append(commandsToSync, f.CommandsToSync()...)
	}

	if len(commandsToSync) == 0 {
		slog.Info("No commands to sync")
		return nil
	}

	slog.Info("Syncing commands to Discord API...")

	if err := handler.SyncCommands(client, commandsToSync, guildsIDs); err != nil {
		return fmt.Errorf("failed to sync commands: %w", err)
	}

	slog.Info("Commands successfully synced")
	return nil
}

// -- Manager factory with options --

// `NewManager` creates a new feature manager
func NewManager(opts ...FeatureManagerOpts) (*FeatureManager, error) {
	fm := &FeatureManager{}

	for _, opt := range opts {
		opt(fm)
	}

	return fm, nil
}

// `FeatureManagerOpts` is a type that allows to pass options to the FeatureManager constructor
type FeatureManagerOpts func(*FeatureManager)

// `WithFeatureSet` sets the feature set for the FeatureManager
func WithFeatureSet(fs FeatureSet) FeatureManagerOpts {
	return func(fm *FeatureManager) {
		fm.features = fs
	}
}

// `WithBotConfig` sets the bot configuration for the FeatureManager
func WithBotConfig(cfg *config.BotConfig) FeatureManagerOpts {
	return func(fm *FeatureManager) {
		fm.botConfig = cfg
	}
}

// `WithGuildsConfigs` sets the guilds configurations for the FeatureManager
func WithGuildsConfigs(cfgs *config.GuildsConfigs) FeatureManagerOpts {
	return func(fm *FeatureManager) {
		fm.guildsConfigs = cfgs
	}
}
