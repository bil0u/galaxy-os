package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/bil0u/galaxy-os/cmd/generators"
	"github.com/bil0u/galaxy-os/pkg"
	"github.com/bil0u/galaxy-os/pkg/features"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
)

var (
	development  string = "false"
	version      string = "dev"
	commit       string = "unknown"
	botsFeatures map[string]pkg.BotFeatureSet
)

// Storing flags in a struct
type CliFlags struct {
	botName         string
	configFile      string
	configDirectory string
	syncCommands    bool
	runAsGenerator  bool
	enableCron      bool
	enableOAuth2    bool
}

func init() {

	pkg.RegisterGenerator(generators.RolesGenerator)
	pkg.RegisterGenerator(generators.ChannelsGenerator)

	botsFeatures = map[string]pkg.BotFeatureSet{
		"hue": {
			// Default Features
			features.BotPresenceFeature{},
			features.BotInfosFeature{},
			features.LogPermissionsFeature{},
			features.TestCommandFeature{},
			features.SelfAssignRolesFeature{},
			// Hue specific Features
			features.DailyMessageFeature{},
			features.SuspiciousInterwiewFeature{},
		},
		"kevin": {
			// Default Features
			features.BotPresenceFeature{},
			features.BotInfosFeature{},
			features.LogPermissionsFeature{},
			features.TestCommandFeature{},
			features.SelfAssignRolesFeature{},
		},
	}
}

func main() {

	// Parse flags
	var flags CliFlags

	flag.StringVar(&flags.botName, "bot", "default", "Name of the bot to run")
	flag.StringVar(&flags.configFile, "use-config", "", "Path to toml configuration file")
	flag.StringVar(&flags.configDirectory, "config-dir", ".", "Path to the directory in which to find the config file")
	flag.BoolVar(&flags.syncCommands, "sync-commands", false, "Whether to sync commands with discord API")
	flag.BoolVar(&flags.runAsGenerator, "generator", false, "Whether to run the bot in generator mode")
	flag.BoolVar(&flags.enableCron, "enable-cron", false, "Whether to enable cron jobs")
	flag.BoolVar(&flags.enableOAuth2, "enable-oauth2", false, "Whether to enable oauth2 server")
	flag.Parse()

	// Get configuration from file
	config, err := pkg.NewConfig(flags.botName, flags.configFile, flags.configDirectory)
	if err != nil {
		slog.Error("Failed to create config", slog.Any("err", err))
		os.Exit(-1)
	}

	config.Development = (development == "true")

	// Run bot in generate mode if needed
	if flags.runAsGenerator {
		config.Log.AddSource = true
		pkg.SetupLogger(config.Log)
		err := startGenerator(config)
		if err != nil {
			slog.Error("Failed to start bot generators", slog.Any("err", err))
			os.Exit(-1)
		}
		os.Exit(0)
	}

	// Run bot in normal mode
	pkg.SetupLogger(config.Log)
	if err := startBot(flags.botName, config, flags.syncCommands, flags.enableCron, flags.enableOAuth2); err != nil {
		slog.Error("Failed to start bot", slog.Any("err", err))
		os.Exit(-1)
	}
}

// Regular bot mode

func startBot(botName string, config pkg.Config, syncCommands, enableCron, enableOAuth2 bool) error {

	dummyClient, err := pkg.NewBotClient(config.Bot.Token, []gateway.Intents{
		gateway.IntentGuilds,
	}, nil)
	if err != nil {
		return fmt.Errorf("error while building disgo dummy client: %w", err)
	}

	// Fetching each guild the bot is in
	botGuilds, err := dummyClient.Rest().GetCurrentUserGuilds("", 0, 0, 0, true)
	if err != nil {
		return fmt.Errorf("failed to fetch bot guilds: %w", err)
	}

	if len(botGuilds) == 0 {
		return fmt.Errorf("bot is not in any guild")
	}

	// Retrieve guild IDs
	guildIDs := make([]snowflake.ID, len(botGuilds))
	for i, guild := range botGuilds {
		guildIDs[i] = guild.ID
	}

	// Creating client using token
	botIntents := []gateway.Intents{gateway.IntentsAll}
	botCaches := []cache.Flags{cache.FlagsAll}
	botClient, err := pkg.NewBotShardedClient(len(guildIDs), config.Bot.Token, botIntents, botCaches)
	if err != nil {
		return err
	}

	// Create bot
	bot := pkg.NewBot(botClient, config, botName, version, commit)

	// Get bot features
	features, ok := botsFeatures[botName]
	if !ok {
		return fmt.Errorf("features not found for bot '%s'", botName)
	}

	// Setup bot
	if err = bot.Setup(guildIDs, features); err != nil {
		return err
	}

	// Start bot
	return bot.Start(syncCommands, enableCron, enableOAuth2)
}

// Generator mode

func startGenerator(config pkg.Config) error {

	slog.Info("Running bot in generator mode...")

	// Creating client to interact with discord
	client, err := pkg.NewBotClient(config.Bot.Token, nil, nil)
	if err != nil {
		return err
	}

	// Run generators
	pkg.RunAllGenerators(client, config)
	slog.Info("Complete!")
	return nil
}
