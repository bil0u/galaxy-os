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

type botDef struct {
	features   *features.Set
	cacheFlags cache.Flags
	intents    gateway.Intents
}

var bots = map[string]botDef{
	"hue": {
		features: features.NewSet(
			bot_features.BotInfosFeature,
			bot_features.LogPermissionsFeature,
			bot_features.TestFeature,
			guild_features.BotPresenceFeature,
			guild_features.SelfAssignRolesFeature,
			guild_features.DailyMessageFeature,
			guild_features.SuspiciousInterviewFeature,
		),
		cacheFlags: cache.FlagGuilds | cache.FlagMembers | cache.FlagRoles,
		intents:    gateway.IntentGuilds | gateway.IntentGuildMembers,
	},
	"kevin": {
		features: features.NewSet(
			bot_features.BotInfosFeature,
			bot_features.LogPermissionsFeature,
			bot_features.TestFeature,
			guild_features.BotPresenceFeature,
			guild_features.SelfAssignRolesFeature,
			guild_features.DailyMessageFeature,
		),
		cacheFlags: cache.FlagGuilds | cache.FlagRoles,
		intents:    gateway.IntentGuilds,
	},
}

func startBot(cmd *cobra.Command, _ []string) error {
	def, ok := bots[bot]
	if !ok {
		return fmt.Errorf("bot %q not found", bot)
	}
	if def.features == nil {
		return fmt.Errorf("no features defined for bot %q", bot)
	}

	if err := config.Init(bot); err != nil {
		return fmt.Errorf("initializing config: %w", err)
	}

	config.GlobalCfg.Development = development == "true"
	config.GlobalCfg.Version = version
	config.GlobalCfg.Commit = commit

	if _, err := services.InitLogger(config.LogCfg.Level, config.LogCfg.Format, config.LogCfg.AddSource); err != nil {
		return fmt.Errorf("initializing logger: %w", err)
	}

	if _, err := services.InitDiscordClient(config.BotCfg.Token, def.cacheFlags, def.intents); err != nil {
		return fmt.Errorf("initializing discord client: %w", err)
	}

	ctx := context.Background()

	if err := config.InitGuilds(ctx, services.RestClient()); err != nil {
		return fmt.Errorf("initializing guilds config: %w", err)
	}

	slog.Debug(fmt.Sprintf("Global configuration: %+v", config.GlobalCfg))
	slog.Debug(fmt.Sprintf("Bot configuration: %++v", config.BotCfg))
	slog.Debug(fmt.Sprintf("Log configuration: %+v", config.LogCfg))
	slog.Debug(fmt.Sprintf("Guilds configuration: %+v", config.GuildsCfg))

	client := services.Client()

	services.InitRouter()
	services.InitPaginator(client)

	if enableCron {
		services.InitCron()
	}

	if enableOAuth2 {
		if err := config.BotCfg.ValidateOAuth(); err != nil {
			return fmt.Errorf("oauth2 config: %w", err)
		}
		services.InitOAuth(config.BotCfg.ApplicationID, config.BotCfg.ClientSecret, config.BotCfg.BaseURL)
	}

	deps := features.SetupDeps{
		Client: client,
		Router: services.Router(),
		Cron:   services.Cron(),
	}

	if err := features.Init(*def.features, deps); err != nil {
		return fmt.Errorf("initializing features: %w", err)
	}

	if syncCommands {
		if err := features.SyncCommands(client, config.GuildsCfg.IDs(development == "true")); err != nil {
			return fmt.Errorf("syncing commands: %w", err)
		}
	}

	if err := client.OpenGateway(ctx); err != nil {
		return fmt.Errorf("opening gateway: %w", err)
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
		client.Close(withTimeout)
	}()

	slog.Info("----- Bot is running 🚀 Press CTRL-C to exit -----")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
	slog.Info("Shutting down bot...")
	return nil
}
