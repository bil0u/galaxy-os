package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/bil0u/galaxy-os/sdk"
)

var (
	development string = "false"
	version     string = "dev"
	commit      string = "unknown"
)

// Storing flags in a struct
type CliFlags struct {
	botName         string
	configFile      string
	configDirectory string
	syncCommands    bool
	runAsGenerator  bool
}

func main() {

	// Parse flags
	var flags CliFlags

	flag.StringVar(&flags.botName, "bot", "default", "Name of the bot to run")
	flag.StringVar(&flags.configFile, "use-config", "", "Path to toml configuration file")
	flag.StringVar(&flags.configDirectory, "config-dir", ".", "Path to the directory in which to find the config file")
	flag.BoolVar(&flags.syncCommands, "sync-commands", false, "Whether to sync commands with discord API")
	flag.BoolVar(&flags.runAsGenerator, "generator", false, "Whether to run the bot in generator mode")
	flag.Parse()

	slog.Info("Starting bot", slog.Any("flags", flags))

	// Get configuration from file
	config, err := sdk.NewConfig(flags.botName, flags.configFile, flags.configDirectory)
	if err != nil {
		slog.Error("Failed to create config", slog.Any("err", err))
		os.Exit(-1)
	}

	config.Development = (development == "true")

	// Run bot in generate mode if needed
	if flags.runAsGenerator {
		config.Log.AddSource = true
		sdk.SetupLogger(config.Log)
		err := startGenerator(config)
		if err != nil {
			slog.Error("Failed to start bot generators", slog.Any("err", err))
			os.Exit(-1)
		}
		os.Exit(0)
	}

	// Run bot in normal mode
	sdk.SetupLogger(config.Log)
	if err := startBot(flags.botName, config, flags.syncCommands); err != nil {
		slog.Error("Failed to start bot", slog.Any("err", err))
		os.Exit(-1)
	}
}

// Regular bot mode

func startBot(botName string, config sdk.Config, syncCommands bool) error {

	// Creating client using token
	botClient, err := sdk.NewBotClient(config.Bot.Token, nil, nil)
	if err != nil {
		return err
	}

	// Create bot
	bot := sdk.NewBot(botClient, config, botName, version, commit)

	// Get bot features
	features, ok := BotsFeatures[botName]
	if !ok {
		return fmt.Errorf("features not found for bot '%s'", botName)
	}

	// Setup bot
	if err = bot.Setup(features); err != nil {
		return err
	}

	// Sync commands if needed
	if syncCommands {
		err = bot.SyncCommands()
		if err != nil {
			return err
		}
	}

	slog.Info("Bot configation: ", slog.Any("config", bot.Config))

	// Start bot
	return bot.Start()
}

// Generator mode

func startGenerator(config sdk.Config) error {

	slog.Info("Running bot in generator mode...")

	// Creating client to interact with discord
	client, err := sdk.NewBotClient(config.Bot.Token, nil, nil)
	if err != nil {
		return err
	}

	// Run generators
	sdk.RunAllGenerators(client, config)
	slog.Info("Complete!")
	return nil
}
