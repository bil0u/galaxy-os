package features

import (
	"fmt"
	"reflect"

	"github.com/bil0u/galaxy-os/internal/config"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/snowflake/v2"
)

var mgr *Manager

// Init initializes the feature module.
// Should be called after config.Init(), config.InitGuilds(), and service initialization.
func Init(fs Set, deps SetupDeps) error {
	var err error

	mgr, err = NewManager(
		WithSet(fs),
		WithBotConfig(config.BotCfg),
		WithGuildsConfigs(config.GuildsCfg),
	)
	if err != nil {
		return fmt.Errorf("creating feature manager: %w", err)
	}

	if err := mgr.SetupFeatures(deps); err != nil {
		return fmt.Errorf("setting up features: %w", err)
	}

	return nil
}

// SyncCommands syncs the registered commands to the Discord API.
func SyncCommands(client bot.Client, guildIDs []snowflake.ID) error {
	if mgr == nil {
		return fmt.Errorf("feature manager not initialized, call features.Init() first")
	}
	return mgr.SyncCommands(client, guildIDs)
}

// GetConfig returns a feature configuration for a specific guild.
func GetConfig[T Config](guildID snowflake.ID) (T, error) {
	var zero T
	if mgr == nil {
		return zero, fmt.Errorf("feature manager not initialized, call features.Init() first")
	}
	cfg, err := mgr.FeatureConfigFor(guildID, reflect.TypeOf(zero))
	if err != nil {
		return zero, err
	}

	return cfg.(T), nil
}

// GetBotConfig returns a feature configuration for the bot.
func GetBotConfig[T Config]() (T, error) {
	return GetConfig[T](0)
}
