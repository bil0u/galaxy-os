package features

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/bil0u/galaxy-os/internal/utils"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/robfig/cron/v3"
)

type Config interface {
	Validate() error
}

type featureType string

const (
	BotFeature   featureType = "bot"
	GuildFeature featureType = "guild"
)

type SetupDeps struct {
	Client bot.Client
	Router *handler.Mux
	Cron   *cron.Cron
}

type Feature struct {
	Type        featureType
	Key         string
	Name        utils.LocalizedString
	Description utils.LocalizedString
	Setup       func(deps SetupDeps) error
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

func New[T Config](setup func(deps SetupDeps) error, opts ...featureOption) Feature {
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
	pattern := regexp.MustCompile(`^.*?\.(.*?)(?:Feature|Config)*$`)
	matches := pattern.FindStringSubmatch(configName)
	if matches != nil {
		feature.Key = utils.ToSnakeCase(matches[1])
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

func WithDescription(description utils.LocalizedString) featureOption {
	return func(f *Feature) {
		f.Description = description
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
