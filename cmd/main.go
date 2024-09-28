package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/bil0u/galaxy-os/cmd/bots"
	"github.com/bil0u/galaxy-os/sdk"
	"github.com/disgoorg/disgo/discord"
)

var (
	botName = "unknown"
	version = "dev"
	commit  = "unknown"
)

// Storing flags in a struct
type flags struct {
	syncCommands bool
	configPath   string
}

func init() {
	sdk.RegisterBotParts("hue", sdk.BotParts{
		Commands:  bots.HueCommands,
		Listeners: bots.HueEventListeners,
	})

	sdk.RegisterBotParts("kevin", sdk.BotParts{
		Commands:  bots.KevinCommands,
		Listeners: bots.KevinEventListeners,
	})
}

func main() {
	// Parse flags
	var f flags
	flag.BoolVar(&f.syncCommands, "sync-commands", false, "Whether to sync commands to discord")
	flag.StringVar(&f.configPath, "config", fmt.Sprintf("config.%s.toml", botName), "path to config")
	flag.Parse()

	// Load config file
	cfg, err := sdk.LoadConfig(f.configPath)
	if err != nil {
		slog.Error("Failed to read config", slog.Any("err", err))
		os.Exit(-1)
	}

	// Setup logger
	sdk.SetupLogger(cfg.Log)

	// Create and start bot
	if err := startBot(cfg, botName, version, commit, f.syncCommands); err != nil {
		slog.Error("Failed to start bot", slog.Any("err", err))
		os.Exit(-1)
	}
}

func startBot(cfg *sdk.Config, botName, version, commit string, syncCommands bool) error {
	b := sdk.NewBot(*cfg, botName, version, commit)

	// Get bot components
	components, err := sdk.GetBotParts(botName)
	if err != nil {
		return err
	}

	// Setup bot listeners
	botListeners := components.Listeners(b)
	if err := b.SetupBot(botListeners); err != nil {
		return err
	}

	// Make sure we don't sync commands if we don't want to
	var botCommands []discord.ApplicationCommandCreate
	if syncCommands {
		botCommands = components.Commands(b)
	}

	// Start bot
	b.Start(botCommands)
	return nil
}
