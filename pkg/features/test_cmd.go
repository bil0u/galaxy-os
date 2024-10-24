package features

import (
	"github.com/bil0u/galaxy-os/pkg"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"
)

func init() {
	pkg.RegisterFeature[TestCommandFeature]("test")
}

type TestCommandFeature struct{}

func (f TestCommandFeature) Name() pkg.LocalizedString {
	return pkg.LocalizedString{
		discord.LocaleEnglishUS: "Test",
		discord.LocaleFrench:    "Test",
	}
}

func (f TestCommandFeature) Description() pkg.LocalizedString {
	return pkg.LocalizedString{
		discord.LocaleEnglishUS: "Testing feature, do not use",
		discord.LocaleFrench:    "Test, ne pas utiliser",
	}
}

func (f TestCommandFeature) IsEnabled() bool {
	return true
}

func (f TestCommandFeature) IsProperlyConfigured() error {
	return nil
}

func (f TestCommandFeature) Setup(bot *pkg.Bot) error {
	bot.Router.Command("/test", TestHandler)
	bot.Router.Autocomplete("/test", TestAutocompleteHandler)
	bot.Router.Component("/test-button", TestComponent)

	bot.AddCommandsToSync(testCommand)
	return nil
}

func TestComponent(e *handler.ComponentEvent) error {
	return e.UpdateMessage(discord.MessageUpdate{
		Content: json.Ptr(pkg.LocalizedString{
			discord.LocaleEnglishUS: "The text has been updated",
			discord.LocaleFrench:    "Le texte a été mis à jour",
		}.String(e.Locale())),
	})
}

var testCommand = discord.SlashCommandCreate{
	Name: "test",
	NameLocalizations: pkg.LocalizedString{
		discord.LocaleEnglishUS: "test",
		discord.LocaleFrench:    "test",
	},
	Description: "Test command",
	DescriptionLocalizations: pkg.LocalizedString{
		discord.LocaleEnglishUS: "Test command",
		discord.LocaleFrench:    "Commande de test",
	},
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name: "choice",
			NameLocalizations: pkg.LocalizedString{
				discord.LocaleEnglishUS: "choice",
				discord.LocaleFrench:    "choix",
			},
			Description: "Select a number",
			DescriptionLocalizations: pkg.LocalizedString{
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
		SetContentf(pkg.LocalizedString{
			discord.LocaleEnglishUS: "Test command. Choice: %s",
			discord.LocaleFrench:    "Commande de test. Choix: %s",
		}[e.Locale()], data.String("choice")).
		// SetContentf, data.String("choice")).
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
