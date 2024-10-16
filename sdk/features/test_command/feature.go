package test_command

import (
	"github.com/bil0u/galaxy-os/sdk"
	"github.com/disgoorg/disgo/discord"
)

func init() {
	sdk.RegisterFeature[TestCommandFeature]("test")
}

type TestCommandFeature struct{}

func (f TestCommandFeature) Name() sdk.LocalizedString {
	return sdk.LocalizedString{
		discord.LocaleEnglishUS: "Version",
		discord.LocaleFrench:    "Version",
	}
}

func (f TestCommandFeature) Description() sdk.LocalizedString {
	return sdk.LocalizedString{
		discord.LocaleEnglishUS: "Display the bot bot-infos",
		discord.LocaleFrench:    "Affiche la bot-infos du bot",
	}
}

func (f TestCommandFeature) IsEnabled() bool {
	return true
}

func (f TestCommandFeature) Setup(b *sdk.Bot) error {
	return nil
}

func (f TestCommandFeature) CommandsCreate() []discord.ApplicationCommandCreate {
	return []discord.ApplicationCommandCreate{
		testCommand,
	}
}
