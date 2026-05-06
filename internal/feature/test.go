package feature

import (
	"github.com/bil0u/galaxy-os/internal/locale"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

type TestConfig struct {
	Foo string
}

func (f TestConfig) Validate() error {
	return nil
}

var TestFeature = New[TestConfig](
	TestSetup,
	WithType(BotFeature),
	WithLocalizedName(discord.LocaleFrench, "Test"),
	WithDescription(locale.Text{
		discord.LocaleEnglishUS: "Testing feature, do not use",
		discord.LocaleFrench:    "Test, ne pas utiliser",
	}),
	WithCommandsToSync(testCommand),
)

func TestSetup(deps SetupDeps) error {
	deps.Bot.Router.SlashCommand("/test", TestHandler)
	deps.Bot.Router.Autocomplete("/test", TestAutocompleteHandler)
	deps.Bot.Router.ButtonComponent("/test-button", TestComponent)
	return nil
}

func TestComponent(_ discord.ButtonInteractionData, e *handler.ComponentEvent) error {
	return e.UpdateMessage(discord.NewMessageUpdate().WithContent(locale.Text{
		discord.LocaleEnglishUS: "The text has been updated",
		discord.LocaleFrench:    "Le texte a été mis à jour",
	}.Using(e.Locale())))
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

func TestHandler(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	return e.CreateMessage(discord.NewMessageCreate().
		WithContentf(locale.Text{
			discord.LocaleEnglishUS: "Test command. Choice: %s",
			discord.LocaleFrench:    "Commande de test. Choix: %s",
		}.Using(e.Locale()), data.String("choice")).
		AddActionRow(discord.NewPrimaryButton("test", "/test-button")),
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
