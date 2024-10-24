package features

import (
	"github.com/bil0u/galaxy-os/pkg"
	"github.com/disgoorg/disgo/discord"
	"github.com/pelletier/go-toml/v2"
)

func init() {
	// pkg.RegisterFeature[DummyFeature]("dummy_feature")
}

type DummyFeature struct {
	Enabled bool `toml:"enabled"`
}

func (f DummyFeature) Name() pkg.LocalizedString {
	return pkg.LocalizedString{
		discord.LocaleEnglishUS: "Dummy",
		discord.LocaleFrench:    "Dummy",
	}
}

func (f DummyFeature) Description() pkg.LocalizedString {
	return pkg.LocalizedString{
		discord.LocaleEnglishUS: "Dummy",
		discord.LocaleFrench:    "Dummy",
	}
}

func (f DummyFeature) IsEnabled() bool {
	return f.Enabled
}

func (f DummyFeature) IsProperlyConfigured() error {
	return nil
}

func (f DummyFeature) Start(bot *pkg.Bot) error {
	return nil
}

func (f DummyFeature) CommandsCreate() []discord.ApplicationCommandCreate {
	return nil
}

func (f DummyFeature) UnmarshalTOML(data any) error {
	bytes, err := toml.Marshal(data)
	if err != nil {
		return err
	}
	return toml.Unmarshal(bytes, &f)
}
