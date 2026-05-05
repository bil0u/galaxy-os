package bot_features

import (
	"github.com/bil0u/galaxy-os/internal/features"
	"github.com/bil0u/galaxy-os/internal/locale"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"
)

type TestConfig struct {
	Foo string
}

func (f TestConfig) Validate() error {
	return nil
}

var TestFeature = features.New[TestConfig](
	setupTestFeature,
	features.WithType(features.BotFeature),
	features.WithLocalizedName(discord.LocaleFrench, "Test"),
	features.WithDescription(locale.Text{
		discord.LocaleEnglishUS: "Testing feature, do not use",
		discord.LocaleFrench:    "Test, ne pas utiliser",
	}),
	features.WithCommandsToSync(testCommand),
)

func setupTestFeature(deps features.SetupDeps) error {
	deps.Router.Command("/test", TestHandler)
	deps.Router.Autocomplete("/test", TestAutocompleteHandler)
	deps.Router.Component("/test-button", TestComponent)
	return nil
}

func TestComponent(e *handler.ComponentEvent) error {
	return e.UpdateMessage(discord.MessageUpdate{
		Content: json.Ptr(locale.Text{
			discord.LocaleEnglishUS: "The text has been updated",
			discord.LocaleFrench:    "Le texte a été mis à jour",
		}.Using(e.Locale())),
	})
}

var testCommand = discord.SlashCommandCreate{
	Name: "test",
	NameLocalizations: locale.Text{
		discord.LocaleEnglishUS: "test",
		discord.LocaleFrench:    "test",
	},
	Description: "Test command",
	DescriptionLocalizations: locale.Text{
		discord.LocaleEnglishUS: "Test command",
		discord.LocaleFrench:    "Commande de test",
	},
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name: "choice",
			NameLocalizations: locale.Text{
				discord.LocaleEnglishUS: "choice",
				discord.LocaleFrench:    "choix",
			},
			Description: "Select a number",
			DescriptionLocalizations: locale.Text{
				discord.LocaleEnglishUS: "Select a number",
				discord.LocaleFrench:    "Selectionne un nombre",
			},
			Required:     true,
			Autocomplete: true,
		},
	},
}

func TestHandler(e *handler.CommandEvent) error {
	data := e.SlashCommandInteractionData()
	return e.CreateMessage(discord.NewMessageCreateBuilder().
		SetContentf(locale.Text{
			discord.LocaleEnglishUS: "Test command. Choice: %s",
			discord.LocaleFrench:    "Commande de test. Choix: %s",
		}.Using(e.Locale()), data.String("choice")).
		AddActionRow(discord.NewPrimaryButton("test", "/test-button")).
		Build(),
	)
}

func TestAutocompleteHandler(e *handler.AutocompleteEvent) error {
	return e.AutocompleteResult([]discord.AutocompleteChoice{
		discord.AutocompleteChoiceString{
			Name:  "1",
			Value: "1",
		},
		discord.AutocompleteChoiceString{
			Name:  "2",
			Value: "2",
		},
		discord.AutocompleteChoiceString{
			Name:  "3",
			Value: "3",
		},
	})
}
