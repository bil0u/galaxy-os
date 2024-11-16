package features

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"sync"

	"github.com/bil0u/galaxy-os/internal/config"
	"github.com/disgoorg/snowflake/v2"
)

var (
	Manager *FeatureManager
	Enabled []string
)

// Init initializes the feature module
// features should be a list of feature keys as defined in the guild or bot configuration
//
// Should be called after all of those functions have been executed:
// - config.Init()
// - config.InitGuilds()
// - services.Init()
func Init(fs FeatureSet) {
	sync.OnceFunc(func() {

		var err error

		Manager, err = NewManager(
			WithFeatureSet(fs),
			WithBotConfig(config.Bot),
			WithGuildsConfigs(config.Guilds),
		)
		if err != nil {
			slog.Error(err.Error())
			os.Exit(-1)
		}

		Manager.SetupFeatures()

	})()
}

// GetConfig returns a feature configuration for a specific guild
func GetConfig[T FeatureConfig](guildID snowflake.ID) (T, error) {

	var zero T
	if Manager == nil {
		return zero, fmt.Errorf("feature manager not initialized. Call features.Init() first")
	}
	config, err := Manager.ConfigFor(guildID, reflect.TypeOf(zero))
	if err != nil {
		return zero, err
	}

	return config.(T), nil
}

// GetGuildConfig returns a feature configuration for the bot
func GetBotConfig[T FeatureConfig]() (T, error) {
	return GetConfig[T](0)
}
