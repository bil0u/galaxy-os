package features

import (
	"fmt"
	"reflect"

	"github.com/bil0u/galaxy-os/internal/config"
	"github.com/disgoorg/snowflake/v2"
)

var Manager *FeatureManager

// Init initializes the feature module.
// Should be called after config.Init(), config.InitGuilds(), and service initialization.
func Init(fs FeatureSet) error {
	var err error

	Manager, err = NewManager(
		WithFeatureSet(fs),
		WithBotConfig(config.Bot),
		WithGuildsConfigs(config.Guilds),
	)
	if err != nil {
		return fmt.Errorf("creating feature manager: %w", err)
	}

	if err := Manager.SetupFeatures(); err != nil {
		return fmt.Errorf("setting up features: %w", err)
	}

	return nil
}

// GetConfig returns a feature configuration for a specific guild.
func GetConfig[T FeatureConfig](guildID snowflake.ID) (T, error) {
	var zero T
	if Manager == nil {
		return zero, fmt.Errorf("feature manager not initialized, call features.Init() first")
	}
	config, err := Manager.ConfigFor(guildID, reflect.TypeOf(zero))
	if err != nil {
		return zero, err
	}

	return config.(T), nil
}

// GetBotConfig returns a feature configuration for the bot.
func GetBotConfig[T FeatureConfig]() (T, error) {
	return GetConfig[T](0)
}
