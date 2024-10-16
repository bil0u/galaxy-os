package suspicious_interview

import (
	"github.com/bil0u/galaxy-os/sdk"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

func init() {
	sdk.RegisterFeature[SuspiciousInterwiewFeature]("suspicious_interview")
}

type SuspiciousInterwiewFeature struct {
	Enabled             bool           `toml:"enabled"`
	DetectRoles         []snowflake.ID `toml:"detect_roles"`
	IfSuccess           snowflake.ID   `toml:"if_success"`
	IfFailure           snowflake.ID   `toml:"if_failure"`
	ClearAfterInterview bool           `toml:"clear_after_interview"`
}

func (f SuspiciousInterwiewFeature) Name() sdk.LocalizedString {
	return sdk.LocalizedString{
		discord.LocaleEnglishUS: "Suspicious Role Interview",
		discord.LocaleFrench:    "Entretien des rôles suspects",
	}
}

func (f SuspiciousInterwiewFeature) Description() sdk.LocalizedString {
	return sdk.LocalizedString{
		discord.LocaleEnglishUS: "Interview users with suspicious roles with a set of questions, and assign them a role based on their answers",
		discord.LocaleFrench:    "Interviewer les utilisateurs avec des rôles suspects, et leur attribuer un rôle en fonction de leurs réponses",
	}
}

func (f SuspiciousInterwiewFeature) IsEnabled() bool {
	return f.Enabled
}

func (f SuspiciousInterwiewFeature) Setup(bot *sdk.Bot) error {
	return nil
}

func (f SuspiciousInterwiewFeature) CommandsCreate() []discord.ApplicationCommandCreate {
	return nil
}
