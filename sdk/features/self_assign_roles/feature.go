package self_assign_roles

import (
	"github.com/bil0u/galaxy-os/sdk"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

func init() {
	sdk.RegisterFeature[SelfAssignRolesFeature]("self_assign_roles")
}

type SelfAssignRolesFeature struct {
	Enabled         bool           `toml:"enabled"`
	Roles           []snowflake.ID `toml:"roles"`
	ClearRolesFirst bool           `toml:"clear_roles_first"`
}

func (f SelfAssignRolesFeature) Name() sdk.LocalizedString {
	return sdk.LocalizedString{
		discord.LocaleEnglishUS: "Self assign roles",
		discord.LocaleFrench:    "Auto-attribution de rôles",
	}
}

func (f SelfAssignRolesFeature) Description() sdk.LocalizedString {
	return sdk.LocalizedString{
		discord.LocaleEnglishUS: "The bot will assign roles to itself automatically",
		discord.LocaleFrench:    "Le bot s'attribuera des rôles automatiquement",
	}
}

func (f SelfAssignRolesFeature) IsEnabled() bool {
	return f.Enabled
}

func (f SelfAssignRolesFeature) Setup(bot *sdk.Bot) error {
	return nil
}
