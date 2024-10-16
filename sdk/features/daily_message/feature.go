package daily_message

import (
	"fmt"
	"regexp"

	"github.com/bil0u/galaxy-os/sdk"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

func init() {
	sdk.RegisterFeature[DailyMessageFeature]("daily_message")
}

type DailyMessageFeature struct {
	Enabled bool         `toml:"enabled"`
	Channel snowflake.ID `toml:"channel"`
	Time    string       `toml:"schedule"`
}

func (f DailyMessageFeature) Name() sdk.LocalizedString {
	return sdk.LocalizedString{
		discord.LocaleEnglishUS: "Daily message",
		discord.LocaleFrench:    "Message du jour",
	}
}

func (f DailyMessageFeature) Description() sdk.LocalizedString {
	return sdk.LocalizedString{
		discord.LocaleEnglishUS: "Send a message every day at a specific time",
		discord.LocaleFrench:    "Envoie un message tous les jours à une heure spécifique",
	}
}

func (f DailyMessageFeature) IsEnabled() bool {
	return f.Enabled
}

func (f DailyMessageFeature) Setup(bot *sdk.Bot) error {
	return nil
}

func (f DailyMessageFeature) CommandsCreate() []discord.ApplicationCommandCreate {
	return nil
}

func (cfg DailyMessageFeature) getCronSchdule() (string, error) {
	pattern := regexp.MustCompile(`^[0-9]{1,2}:[0-9]{1,2}$`)
	match := pattern.FindStringSubmatch(cfg.Time)
	if match == nil {
		return "", fmt.Errorf("invalid time format")
	}
	return fmt.Sprintf("%s %s * * *", match[2], match[1]), nil
}
