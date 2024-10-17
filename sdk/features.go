package sdk

import (
	"fmt"
	"log/slog"
	"reflect"
	"sync"

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
	// Setup sets up the feature using a feature kit
	Setup(b *Bot) error
	// CommandCreate returns the command to create the feature
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

	// Iterating over the superset to create the features
	for _, botFeature := range superset {

		// Getting the type of the feature
		featureType := reflect.TypeOf(botFeature)

		// Checking for the feature key in the known features
		var featureKey string
		for key, candidate := range knowFeatures {
			if reflect.TypeOf(candidate) == featureType {
				featureKey = key
				break
			}
		}

		// If the feature is not registered, we ignore it
		if featureKey == "" {
			slog.Error(fmt.Sprintf("feature '%s' not registered, skipping", featureType))
			continue
		}

		// Create a new feature instance to unmarshal the data into
		featureInstance := reflect.New(featureType).Interface().(Feature)

		// Retrieve the feature data from the configuration
		dataBytes, _ := toml.Marshal(botDefinitions[featureKey])

		// Unmarshal the data into the feature instance
		err := toml.Unmarshal(dataBytes, featureInstance)
		if err != nil {
			slog.Error(fmt.Sprintf("failed to unmarshal feature '%s': %s", featureKey, err))
			continue
		}

		// Save the feature instance
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
