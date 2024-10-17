package log_permissions

import (
	"github.com/bil0u/galaxy-os/sdk"
	"github.com/disgoorg/disgo/discord"
)

func init() {
	sdk.RegisterFeature[LogPermissionsFeature]("log_permissions")
}

type LogPermissionsFeature struct{}

func (f LogPermissionsFeature) Name() sdk.LocalizedString {
	return sdk.LocalizedString{
		discord.LocaleEnglishUS: "Log Permissions",
		discord.LocaleFrench:    "Affiche les Permissions",
	}
}

func (f LogPermissionsFeature) Description() sdk.LocalizedString {
	return sdk.LocalizedString{
		discord.LocaleEnglishUS: "Log the permissions of the bot",
		discord.LocaleFrench:    "Affiche les permissions du bot",
	}
}

func (f LogPermissionsFeature) IsEnabled() bool {
	return true
}

func (f LogPermissionsFeature) Setup(bot *sdk.Bot) error {
	return nil
}
