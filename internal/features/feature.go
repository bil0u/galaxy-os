package features

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/bil0u/galaxy-os/internal/utils"
	"github.com/disgoorg/disgo/discord"
)

// type FeatureConfig interface {
// 	Validate() error
// }

type FeatureConfig interface {
	Validate() error
}

type featureType string

const (
	BotFeature   featureType = "bot"
	GuildFeature featureType = "guild"
)

type Feature struct {
	Type        featureType
	Key         string
	Name        utils.LocalizedString
	Description utils.LocalizedString
	Setup       func() error
	cfgType     reflect.Type
	cmdCreates  []discord.ApplicationCommandCreate
}

func (f Feature) IsValid() error {
	var errs []error
	if f.Type == "" {
		errs = append(errs, fmt.Errorf("type is required"))
	}
	if f.Key == "" {
		errs = append(errs, fmt.Errorf("key is required"))
	}
	if f.Name.String() == "" {
		errs = append(errs, fmt.Errorf("name is required"))
	}
	if f.Setup == nil {
		errs = append(errs, fmt.Errorf("setup function is required"))
	}
	if f.cfgType == nil {
		errs = append(errs, fmt.Errorf("config type is required"))
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid feature: %v", errs)
	}
	return nil
}

func (f Feature) CommandsToSync() []discord.ApplicationCommandCreate {
	if f.cmdCreates == nil {
		return []discord.ApplicationCommandCreate{}
	}
	return f.cmdCreates
}

func New[T FeatureConfig](setup func() error, opts ...featureOption) Feature {
	var zero T

	feature := defaultFeature(fmt.Sprintf("%T", zero))
	feature.Setup = setup
	feature.cfgType = reflect.TypeOf(zero)
	for _, opt := range opts {
		opt(&feature)
	}
	return feature
}

type featureOption func(*Feature)

func defaultFeature(configName string) Feature {
	feature := Feature{}
	// Set the default key, matching the following pattern: "bot_features.(BotExample)FeatureConfig"
	if strings.HasPrefix(configName, "bot") {
		feature.Type = BotFeature
	} else if strings.HasPrefix(configName, "guild") {
		feature.Type = GuildFeature
	}
	// Extract base string from the config name
	pattern := regexp.MustCompile(`^.*?\.(.*?)(?:Feature|Config)*$`)
	matches := pattern.FindStringSubmatch(configName)
	if matches != nil {
		// Convert the key to snake case
		feature.Key = utils.ToSnakeCase(matches[1])
		// Set the default name, adding spaces between camel case words
		feature.Name = utils.LocalizedString{
			discord.LocaleEnglishUS: utils.AddSpaces(matches[1]),
		}
	}
	return feature
}

func WithType(fType featureType) featureOption {
	return func(f *Feature) {
		f.Type = fType
	}
}

func WithKey(key string) featureOption {
	return func(f *Feature) {
		f.Key = key
	}
}

func WithName(name utils.LocalizedString) featureOption {
	return func(f *Feature) {
		f.Name = name
	}
}

func WithDescription(name utils.LocalizedString) featureOption {
	return func(f *Feature) {
		f.Name = name
	}
}

func WithLocalizedName(locale discord.Locale, name string) featureOption {
	return func(f *Feature) {
		if f.Name == nil {
			f.Name = utils.LocalizedString{}
		}
		f.Name[locale] = name
	}
}

func WithLocalizedDescription(locale discord.Locale, description string) featureOption {
	return func(f *Feature) {
		if f.Description == nil {
			f.Description = utils.LocalizedString{}
		}
		f.Description[locale] = description
	}
}

func WithCommandsToSync(cmd ...discord.ApplicationCommandCreate) featureOption {
	return func(f *Feature) {
		f.cmdCreates = append(f.cmdCreates, cmd...)
	}
}
