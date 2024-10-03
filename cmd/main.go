package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/bil0u/galaxy-os/sdk"
	"github.com/bil0u/galaxy-os/sdk/utils"
	"github.com/disgoorg/disgo/discord"
)

const defaultConfig = "config.default.toml"

var (
	version string = "dev"
	commit  string = "unknown"
)

// Storing flags in a struct
type CliFlags struct {
	configDirectory string
	configFile      string
	botName         string
	runAsGenerator  bool
	syncCommands    bool
	syncRoles       bool
	logPermissions  bool
}

func main() {

	// Parse flags
	var flags CliFlags

	flag.StringVar(&flags.botName, "bot", "default", "Name of the bot to run")
	flag.StringVar(&flags.configFile, "use-config", "", "Path to toml configuration file")
	flag.StringVar(&flags.configDirectory, "config-dir", ".", "Path to the directory in which to find the config file")
	flag.BoolVar(&flags.runAsGenerator, "generator", false, "Whether to run the bot only in generate mode")
	flag.BoolVar(&flags.syncCommands, "sync-commands", false, "Whether to sync commands to discord")
	flag.BoolVar(&flags.syncRoles, "sync-roles", false, "Whether to sync bot roles to guilds")
	flag.BoolVar(&flags.logPermissions, "log-permissions", false, "If true, log bot application permissions")
	flag.Parse()

	// Run bot in generator mode if bot name is "generator"
	if flags.botName == "generator" {
		flags.runAsGenerator = true
	}

	// Run bot in generate mode if needed
	if flags.runAsGenerator {
		if err := startGenerator(flags); err != nil {
			slog.Error("Failed to start bot in generate mode", slog.Any("err", err))
			os.Exit(-1)
		}
		os.Exit(0)
	}

	// Run bot in normal mode
	if err := startBot(flags); err != nil {
		slog.Error("Failed to start bot", slog.Any("err", err))
		os.Exit(-1)
	}
}

func getConfig(flags CliFlags) (*sdk.Config, error) {
	config := new(sdk.Config)

	// Load generic config file
	defaultConfigPath := fmt.Sprintf("%s/%s", flags.configDirectory, defaultConfig)
	config, err := sdk.LoadConfig(defaultConfigPath, config)
	if err != nil {
		return nil, fmt.Errorf("encountered error while loading default config '%s'", defaultConfig)
	}

	// Retrieve config file passed as argument, or default to bot specific config
	botConfigFile := fmt.Sprintf("config.%s.toml", flags.botName)
	if flags.configFile != "" {
		botConfigFile = flags.configFile
	}

	// Load bot specific config using the same logic
	botConfigPath := fmt.Sprintf("%s/%s", flags.configDirectory, botConfigFile)
	if botConfigPath != defaultConfigPath {
		config, err = sdk.LoadConfig(botConfigPath, config)
		if err != nil {
			return nil, fmt.Errorf("encountered error while loading bot config '%s'", botConfigFile)
		}
	}

	// Validate config
	if err := sdk.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	return config, nil
}

func startBot(flags CliFlags) error {

	config, err := getConfig(flags)
	if err != nil {
		return fmt.Errorf("failed to create config: %v", err)
	}

	// Setup logger
	sdk.SetupLogger(config.Log)
	// Create bot
	b := sdk.NewBot(*config, flags.botName, version, commit)

	// Get bot components
	botParts, err := sdk.GetBotParts(flags.botName)
	if err != nil {
		return err
	}

	// Creating client using token
	botClient, err := sdk.NewBotClient(b.Cfg.Bot.Token, botParts)
	if err != nil {
		return err
	}
	b.Client = *botClient

	// Setup bot
	if err = b.SetupBot(botParts); err != nil {
		return err
	}

	// Make sure we don't sync commands if we don't want to
	var botCommands []discord.ApplicationCommandCreate
	if flags.syncCommands {
		botCommands = botParts.Commands
	}
	// Log permissions if needed
	if flags.logPermissions {
		utils.LogPermissions(b.Client, b.Cfg.Bot.DevGuilds)
	}

	// Start bot
	b.Start(botCommands, flags.syncRoles)
	return nil
}

// Generator mode

func startGenerator(flags CliFlags) error {

	// Defaulting to hue config for generators
	flags.configFile = "config.hue.toml"
	config, err := getConfig(flags)
	if err != nil {
		return fmt.Errorf("failed to create config: %v", err)
	}

	config.Log.AddSource = true

	// Setup logger
	sdk.SetupLogger(config.Log)

	slog.Info("Running bot in generator mode...")

	// Get bot components
	botParts, err := sdk.GetBotParts("generator")
	if err != nil {
		return err
	}

	// Creating client to interact with discord
	client, err := sdk.NewBotClient(config.Bot.Token, botParts)
	if err != nil {
		return err
	}

	// Run generators
	utils.RunAllGenerators(*client, config.Bot)
	slog.Info("Complete!")
	return nil
}
