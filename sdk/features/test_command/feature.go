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
		discord.LocaleEnglishUS: "Test",
		discord.LocaleFrench:    "Test",
	}
}

func (f TestCommandFeature) Description() sdk.LocalizedString {
	return sdk.LocalizedString{
		discord.LocaleEnglishUS: "Testing feature, do not use",
		discord.LocaleFrench:    "Test, ne pas utiliser",
	}
}

func (f TestCommandFeature) IsEnabled() bool {
	return true
}

func (f TestCommandFeature) Setup(bot *sdk.Bot) error {
	bot.AddCommandsToSync(testCommand)
	return nil
}
