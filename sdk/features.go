package sdk

import (
	"fmt"
	"log/slog"
	"reflect"
	"sync"

	"github.com/disgoorg/disgo/discord"
	"github.com/pelletier/go-toml/v2"
)

var (
	knowFeatures = map[string]Feature{}
	mu           sync.Mutex
)

type Feature interface {
	// Name returns the name of the feature
	Name() LocalizedString
	// Description returns the description of the feature
	Description() LocalizedString
	// IsEnabled returns whether the feature is enabled
	IsEnabled() bool
	// Setup sets up the feature using the bot
	Setup(bot *Bot) error
	// CommandCreate returns the command to create the feature
	CommandsCreate() []discord.ApplicationCommandCreate
}

// RegisterFeature registers a feature constructor with the SDK
func RegisterFeature[T Feature](name string) {
	mu.Lock()
	defer mu.Unlock()
	var zero T
	knowFeatures[name] = zero
}

type TomlFeaturesDefinition struct {
	RawData map[string]map[string]any `toml:"features"`
}

// Custom unmarshalling logic for GuildFeatureSet
func (fd TomlFeaturesDefinition) ToFeatures(botName string, superset BotFeatureSet) []Feature {

	features := []Feature{}

	botDefinitions, ok := fd.RawData[botName]
	if !ok {
		slog.Error(fmt.Sprintf("bot '%s' not found in features definition", botName))
		return nil
	}

	for featureKey, data := range botDefinitions {
		feature, ok := knowFeatures[featureKey]
		if !ok {
			slog.Error(fmt.Sprintf("feature '%s' not registered, skipping", featureKey))
			continue
		}

		if !superset.HasFeature(feature) {
			slog.Info(fmt.Sprintf("feature '%s' not enabled, skipping", featureKey))
			continue
		}

		// create a new instance of the feature
		featureInstance := reflect.New(reflect.TypeOf(feature)).Interface().(Feature)

		dataBytes, _ := toml.Marshal(data)
		err := toml.Unmarshal(dataBytes, featureInstance)
		if err != nil {
			slog.Error(fmt.Sprintf("failed to unmarshal feature '%s': %s", featureKey, err))
			continue
		}

		features = append(features, featureInstance)
	}

	return features
}

// BotFeatureSet is a slice of features
type BotFeatureSet []Feature

// HasFeature returns whether the feature BotFeatureSet has the provided feature
func (bfs BotFeatureSet) HasFeature(feature Feature) bool {
	for _, f := range bfs {
		if reflect.TypeOf(f) == reflect.TypeOf(feature) {
			return true
		}
	}
	return false
}

// GuildFeatureSet is a slice of features
type GuildFeatureSet []Feature

// GetFeatures returns all registered features
func (gfs GuildFeatureSet) GetFeatures(onlyEnabled bool) []Feature {
	enabledFeatures := []Feature{}
	for _, feature := range gfs {
		if onlyEnabled && feature.IsEnabled() {
			enabledFeatures = append(enabledFeatures, feature)
		} else if !onlyEnabled {
			enabledFeatures = append(enabledFeatures, feature)
		}
	}
	return enabledFeatures
}

// GetFeature returns a feature by key from a feature slice, casting it to the provided type
func GetFeature[T Feature](features GuildFeatureSet) (T, error) {
	var zero T
	for _, feat := range features {
		if reflect.TypeOf(feat) == reflect.TypeOf(zero) {
			return feat.(T), nil
		}
	}
	return zero, fmt.Errorf("feature '%T' not found", reflect.TypeOf(zero))
}
