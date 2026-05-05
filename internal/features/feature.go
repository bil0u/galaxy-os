package features

import (
	"bytes"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/bil0u/galaxy-os/internal/config"
	"github.com/bil0u/galaxy-os/internal/locale"
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

// BotServices holds per-bot Discord resources.
type BotServices struct {
	Client *bot.Client
	Router *handler.Mux
	Logger *slog.Logger
}

// SharedServices holds services shared across bots.
type SharedServices struct {
	Cron *cron.Cron
}

// Configs holds all configuration needed by features.
type Configs struct {
	Bot      *config.Bot
	Guilds   *config.GuildMap
	Global   *config.Global
	Features *FeatureRegistry
}

// SetupDeps groups all dependencies passed to feature Setup functions.
type SetupDeps struct {
	Bot     BotServices
	Shared  SharedServices
	Configs Configs
}

type Feature struct {
	Type        featureType
	Key         string
	Name        locale.Text
	Description locale.Text
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

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func addSpaces(s string) string {
	buf := &bytes.Buffer{}
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			buf.WriteRune(' ')
		}
		buf.WriteRune(r)
	}
	return buf.String()
}

func defaultFeature(configName string) Feature {
	feature := Feature{}
	pattern := regexp.MustCompile(`^.*?\.(.*?)(?:Feature|Config)*$`)
	matches := pattern.FindStringSubmatch(configName)
	if matches != nil {
		feature.Key = toSnakeCase(matches[1])
		feature.Name = locale.Text{
			discord.LocaleEnglishUS: addSpaces(matches[1]),
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

func WithName(name locale.Text) featureOption {
	return func(f *Feature) {
		f.Name = name
	}
}

func WithDescription(description locale.Text) featureOption {
	return func(f *Feature) {
		f.Description = description
	}
}

func WithLocalizedName(l discord.Locale, name string) featureOption {
	return func(f *Feature) {
		if f.Name == nil {
			f.Name = locale.Text{}
		}
		f.Name[l] = name
	}
}

func WithLocalizedDescription(l discord.Locale, description string) featureOption {
	return func(f *Feature) {
		if f.Description == nil {
			f.Description = locale.Text{}
		}
		f.Description[l] = description
	}
}

func WithCommandsToSync(cmd ...discord.ApplicationCommandCreate) featureOption {
	return func(f *Feature) {
		f.cmdCreates = append(f.cmdCreates, cmd...)
	}
}
