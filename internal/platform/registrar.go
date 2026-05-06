package platform

import "github.com/disgoorg/disgo/handler"

// Registrar provides typed command and interaction registration.
// Features register through this, never directly on the disgo router.
type Registrar interface {
	SlashCommand(path string, h SlashCommandHandler)
	ButtonComponent(customID string, h ButtonComponentHandler)
	Autocomplete(path string, h AutocompleteHandler)
}

// SlashCommandHandler handles a slash command interaction.
type SlashCommandHandler func(e *handler.CommandEvent) error

// ButtonComponentHandler handles a button component interaction.
type ButtonComponentHandler func(e *handler.ComponentEvent) error

// AutocompleteHandler handles an autocomplete interaction.
type AutocompleteHandler func(e *handler.AutocompleteEvent) error
