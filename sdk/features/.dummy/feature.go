package feature

import (
	"github.com/bil0u/galaxy-os/sdk"
	"github.com/disgoorg/disgo/discord"
	"github.com/pelletier/go-toml/v2"
)

func init() {
	sdk.RegisterFeature[DummyFeature]("dummy_feature")
}

type DummyFeature struct {
	Enabled bool `toml:"enabled"`
}

func (f DummyFeature) Name() sdk.LocalizedString {
	return sdk.LocalizedString{
		discord.LocaleEnglishUS: "Dummy",
		discord.LocaleFrench:    "Dummy",
	}
}

func (f DummyFeature) Description() sdk.LocalizedString {
	return sdk.LocalizedString{
		discord.LocaleEnglishUS: "Dummy",
		discord.LocaleFrench:    "Dummy",
	}
}

func (f DummyFeature) IsEnabled() bool {
	return f.Enabled
}

func (f DummyFeature) Setup(bot *sdk.Bot) error {
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
