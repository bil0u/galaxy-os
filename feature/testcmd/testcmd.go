package testcmd

import (
	"context"

	"github.com/bil0u/galaxy-os/internal/contracts"
	"github.com/bil0u/galaxy-os/internal/i18n"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

// Test is a testing feature, not intended for production use.
var Feature = &test{}

type test struct{}

type testConfig struct {
	Foo string
}

func (c testConfig) Validate() error { return nil }

func (f *test) Name() string                 { return "test" }
func (f *test) Scope() contracts.Scope       { return contracts.BotScope }
func (f *test) Needs() []contracts.ServiceID { return nil }

func (f *test) Setup(deps contracts.Deps) error {
	deps.Commands.SlashCommand("/test", testHandler)
	deps.Commands.Autocomplete("/test", testAutocompleteHandler)
	deps.Commands.ButtonComponent("/test-button", testComponentHandler)
	return nil
}

func (f *test) Start(ctx context.Context) error { return nil }
func (f *test) Stop(ctx context.Context) error  { return nil }

func testComponentHandler(e *handler.ComponentEvent) error {
	return e.UpdateMessage(discord.NewMessageUpdate().WithContent(i18n.Text{
		discord.LocaleEnglishUS: "The text has been updated",
		discord.LocaleFrench:    "Le texte a été mis à jour",
	}.Using(e.Locale())))
}

var testCommand = discord.SlashCommandCreate{
	Name: "test",
	NameLocalizations: i18n.Text{
		discord.LocaleEnglishUS: "test",
		discord.LocaleFrench:    "test",
	},
	Description: "Test command",
	DescriptionLocalizations: i18n.Text{
		discord.LocaleEnglishUS: "Test command",
		discord.LocaleFrench:    "Commande de test",
	},
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionString{
			Name: "choice",
			NameLocalizations: i18n.Text{
				discord.LocaleEnglishUS: "choice",
				discord.LocaleFrench:    "choix",
			},
			Description: "Select a number",
			DescriptionLocalizations: i18n.Text{
				discord.LocaleEnglishUS: "Select a number",
				discord.LocaleFrench:    "Selectionne un nombre",
			},
			Required:     true,
			Autocomplete: true,
		},
	},
}

func testHandler(e *handler.CommandEvent) error {
	data := e.SlashCommandInteractionData()
	return e.CreateMessage(discord.NewMessageCreate().
		WithContentf(i18n.Text{
			discord.LocaleEnglishUS: "Test command. Choice: %s",
			discord.LocaleFrench:    "Commande de test. Choix: %s",
		}.Using(e.Locale()), data.String("choice")).
		AddActionRow(discord.NewPrimaryButton("test", "/test-button")),
	)
}

func testAutocompleteHandler(e *handler.AutocompleteEvent) error {
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
