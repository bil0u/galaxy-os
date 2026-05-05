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
	syncCommands bool
)

var botCmd = &cobra.Command{
	Use:   "bot",
	Short: "Main command for the bot",
}

func init() {
	botCmd.AddCommand(startCmd)

	startCmd.Flags().BoolVarP(&enableCron, "cron", "c", false, "Whether to enable cron jobs")
	startCmd.Flags().BoolVarP(&enableOAuth2, "oauth2", "o", false, "Whether to enable oauth2 server")
	startCmd.Flags().BoolVarP(&syncCommands, "sync", "s", false, "Sync slash commands to Discord before starting")
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the discord bot",
	RunE:  startBot,
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

func startBot(cmd *cobra.Command, _ []string) error {
	featureSet, ok := botFeatures[bot]
	if !ok {
		return fmt.Errorf("bot %q not found", bot)
	}
	if featureSet == nil {
		return fmt.Errorf("no features found for bot %q", bot)
	}

	if err := config.Init(bot); err != nil {
		return fmt.Errorf("initializing config: %w", err)
	}

	config.Global.Development = development == "true"
	config.Global.Version = version
	config.Global.Commit = commit

	if _, err := services.InitLogger(config.Log.Level, config.Log.Format, config.Log.AddSource); err != nil {
		return fmt.Errorf("initializing logger: %w", err)
	}

	if _, err := services.InitDiscordClient(config.Bot.Token, []cache.Flags{cache.FlagsAll}, []gateway.Intents{gateway.IntentsAll}); err != nil {
		return fmt.Errorf("initializing discord client: %w", err)
	}

	if err := config.InitGuilds(services.GetRestClient()); err != nil {
		return fmt.Errorf("initializing guilds config: %w", err)
	}

	if config.Guilds.Count() > 0 {
		if _, err := services.InitDiscordShardedClient(config.Bot.Token, config.Guilds.Count(), []cache.Flags{cache.FlagsAll}, []gateway.Intents{gateway.IntentsAll}); err != nil {
			return fmt.Errorf("initializing sharded discord client: %w", err)
		}
	}

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

	if err := features.Init(*featureSet); err != nil {
		return fmt.Errorf("initializing features: %w", err)
	}

	if syncCommands {
		client := services.GetClient()
		if err := features.Manager.SyncCommands(*client, config.Guilds.IDs(development == "true")); err != nil {
			return fmt.Errorf("syncing commands: %w", err)
		}
	}

	ctx := context.Background()

	if (*client).HasShardManager() {
		if err := (*client).OpenShardManager(ctx); err != nil {
			return fmt.Errorf("opening gateway through shard manager: %w", err)
		}
	} else {
		if err := (*client).OpenGateway(ctx); err != nil {
			return fmt.Errorf("opening gateway: %w", err)
		}
	}

	if enableCron {
		services.StartCron()
	}

	if enableOAuth2 {
		services.StartOAuth()
	}

	defer func() {
		if enableCron {
			services.StopCron()
		}
		withTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		(*client).Close(withTimeout)
	}()

	slog.Info("----- Bot is running 🚀 Press CTRL-C to exit -----")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
	slog.Info("Shutting down bot...")
	return nil
}
