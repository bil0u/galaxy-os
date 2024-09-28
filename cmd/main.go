package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/bil0u/galaxy-os/cmd/bots"
	"github.com/bil0u/galaxy-os/sdk"
	"github.com/disgoorg/disgo/bot"
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

	// Create bot
	b := sdk.NewBot(*cfg, botName, version, commit)
	var (
		botCommands  []discord.ApplicationCommandCreate
		botListeners []bot.EventListener
	)

	// Checking if the provided bot name is valid
	switch botName {
	case "hue":
		botCommands = bots.HueCommands
		botListeners = bots.HueEventListeners(b)
	case "kevin":
		botCommands = bots.KevinCommands
		botListeners = bots.KevinEventListeners(b)
	default:
		slog.Error("Unknown bot", slog.String("bot", botName))
		os.Exit(-1)
	}

	// Make sure we don't sync commands if we don't want to
	if !f.syncCommands {
		botCommands = nil
	}

	b.SetupBot(botListeners)
	b.Start(botCommands)

}
