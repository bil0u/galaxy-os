package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bil0u/galaxy-os/internal/config"
	"github.com/bil0u/galaxy-os/internal/features"
	bot_features "github.com/bil0u/galaxy-os/internal/features/bot"
	guild_features "github.com/bil0u/galaxy-os/internal/features/guild"
	"github.com/bil0u/galaxy-os/internal/services"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/gateway"
	"github.com/spf13/cobra"
)

var (
	enableCron   bool
	enableOAuth2 bool
)

var botCmd = &cobra.Command{
	Use:   "bot",
	Short: "Main command for the bot",
}

func init() {
	botCmd.AddCommand(startCmd)

	startCmd.Flags().BoolVarP(&enableCron, "cron", "c", false, "Whether to enable cron jobs")
	startCmd.Flags().BoolVarP(&enableOAuth2, "oauth2", "o", false, "Whether to enable oauth2 server")
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the discord bot",
	Run:   startBot,
}

var botFeatures = map[string]*features.FeatureSet{
	"hue": features.NewFeatureSet(
		// Bot related features
		bot_features.BotInfosFeature,
		bot_features.LogPermissionsFeature,
		bot_features.TestFeature,
		// Guild related features
		guild_features.BotPresenceFeature,
		guild_features.SelfAssignRolesFeature,
		guild_features.DailyMessageFeature,
		guild_features.SuspiciousInterviewFeature,
	),
	"kevin": features.NewFeatureSet(
		// Bot related features
		bot_features.BotInfosFeature,
		bot_features.LogPermissionsFeature,
		bot_features.TestFeature,
		// Guild related features
		guild_features.BotPresenceFeature,
		guild_features.SelfAssignRolesFeature,
		guild_features.DailyMessageFeature,
	),
}

func startBot(cmd *cobra.Command, _ []string) {

	featureSet, ok := botFeatures[bot]
	if !ok {
		panic(fmt.Errorf("bot %s not found", bot))
	}
	if featureSet == nil {
		panic(fmt.Errorf("no features found for bot %s", bot))
	}

	config.Init(bot)

	config.Global.Development = development == "true"
	config.Global.Version = version
	config.Global.Commit = commit

	// // Init the services
	services.InitLogger(config.Log.Level, config.Log.Format, config.Log.AddSource)

	services.InitDiscordClient(config.Bot.Token, []cache.Flags{cache.FlagsAll}, []gateway.Intents{gateway.IntentsAll})
	config.InitGuilds(services.GetRestClient())

	if config.Guilds.Count() > 0 {
		services.InitDiscordShardedClient(config.Bot.Token, config.Guilds.Count(), []cache.Flags{cache.FlagsAll}, []gateway.Intents{gateway.IntentsAll})
	}

	// Logging all configs in debug mode
	slog.Debug(fmt.Sprintf("Global configuration: %+v", config.Global))
	slog.Debug(fmt.Sprintf("Bot configuration: %++v", config.Bot))
	slog.Debug(fmt.Sprintf("Log configuration: %+v", config.Log))
	slog.Debug(fmt.Sprintf("Guilds configuration: %+v", config.Guilds))

	client := services.GetClient()

	services.InitRouter()
	services.InitPaginator(client)

	if enableCron {
		services.InitCron()
	}

	if enableOAuth2 {
		services.InitOAuth(config.Bot.ApplicationID, config.Bot.ClientSecret, config.Bot.BaseURL)
	}

	// Init the features
	features.Init(*featureSet)

	ctx := context.Background()

	if (*client).HasShardManager() {
		if err := (*client).OpenShardManager(ctx); err != nil {
			panic(fmt.Errorf("failed to open gateway through shard manager: %w", err))
		}
	} else {
		if err := (*client).OpenGateway(ctx); err != nil {
			panic(fmt.Errorf("failed to open gateway: %w", err))
		}
	}

	if enableCron {
		services.StartCron(ctx)
	}

	if enableOAuth2 {
		services.StartOAuth(ctx)
	}

	defer func() {
		withTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		(*client).Close(withTimeout)
		ctx.Done()
	}()

	// Wait for signal to shutdown
	slog.Info("----- Bot is running 🚀 Press CTRL-C to exit -----")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
	slog.Info("Shutting down bot...")
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync the discord bot commands",
	Run:   syncCommands,
}

func syncCommands(cmd *cobra.Command, _ []string) {
	featureSet, ok := botFeatures[bot]
	if !ok {
		panic(fmt.Errorf("bot %s not found", bot))
	}

	config.Init(bot)

	services.InitLogger(config.Log.Level, config.Log.Format, config.Log.AddSource)
	client := services.InitDiscordClient(config.Bot.Token, []cache.Flags{cache.FlagsAll}, []gateway.Intents{gateway.IntentsAll})
	config.InitGuilds((*client).Rest())

	manager, _ := features.NewManager(
		features.WithFeatureSet(*featureSet),
	)

	manager.SyncCommands(*client, config.Guilds.IDs(development == "true"))
}
