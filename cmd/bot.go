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
	"github.com/bil0u/galaxy-os/internal/feature"
	"github.com/bil0u/galaxy-os/internal/service"
	"github.com/disgoorg/disgo"
	disbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
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
	features   *feature.Set
	cacheFlags cache.Flags
	intents    gateway.Intents
}

var bots = map[string]botDef{
	"hue": {
		features: feature.NewSet(
			feature.BotInfosFeature,
			feature.LogPermissionsFeature,
			feature.TestFeature,
			feature.BotPresenceFeature,
			feature.SelfAssignRolesFeature,
			feature.DailyMessageFeature,
			feature.SuspiciousInterviewFeature,
		),
		cacheFlags: cache.FlagGuilds | cache.FlagMembers | cache.FlagRoles,
		intents:    gateway.IntentGuilds | gateway.IntentGuildMembers,
	},
	"kevin": {
		features: feature.NewSet(
			feature.BotInfosFeature,
			feature.LogPermissionsFeature,
			feature.TestFeature,
			feature.BotPresenceFeature,
			feature.SelfAssignRolesFeature,
			feature.DailyMessageFeature,
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

	globalCfg, logCfg, botCfg, err := config.Init(bot)
	if err != nil {
		return fmt.Errorf("initializing config: %w", err)
	}

	globalCfg.Development = development == "true"
	globalCfg.Version = version
	globalCfg.Commit = commit

	if _, err := service.InitLogger(logCfg.Level, logCfg.Format, logCfg.AddSource); err != nil {
		return fmt.Errorf("initializing logger: %w", err)
	}

	botLogger := slog.Default().With("bot", bot)

	client, err := disgo.New(botCfg.Token,
		disbot.WithCacheConfigOpts(cache.WithCaches(def.cacheFlags)),
		disbot.WithGatewayConfigOpts(
			gateway.WithIntents(def.intents),
		),
	)
	if err != nil {
		return fmt.Errorf("building discord client: %w", err)
	}

	ctx := context.Background()

	guilds, err := config.InitGuilds(ctx, client.Rest, bot)
	if err != nil {
		return fmt.Errorf("initializing guilds config: %w", err)
	}

	slog.Debug(fmt.Sprintf("Global configuration: %+v", globalCfg))
	slog.Debug(fmt.Sprintf("Bot configuration: %++v", botCfg))
	slog.Debug(fmt.Sprintf("Log configuration: %+v", logCfg))
	slog.Debug(fmt.Sprintf("Guilds configuration: %+v", guilds))

	router := handler.New()
	client.AddEventListeners(router)

	client.AddEventListeners(disbot.NewListenerFunc(func(_ *events.Resumed) {
		botLogger.Info("gateway reconnected")
	}))

	if enableCron {
		service.InitCron()
	}

	if enableOAuth2 {
		if err := botCfg.ValidateOAuth(); err != nil {
			return fmt.Errorf("oauth2 config: %w", err)
		}
		service.InitOAuth(botCfg.ApplicationID, botCfg.ClientSecret, botCfg.BaseURL)
	}

	registry := feature.NewRegistry(*def.features, botCfg, guilds)

	deps := feature.SetupDeps{
		Bot: feature.BotServices{
			Client: client,
			Router: router,
			Logger: botLogger,
		},
		Shared: feature.SharedServices{
			Cron: service.Scheduler(),
		},
		Configs: feature.Configs{
			Bot:      botCfg,
			Guilds:   guilds,
			Global:   globalCfg,
			Features: registry,
		},
	}

	if err := feature.SetupFeatures(*def.features, deps); err != nil {
		return fmt.Errorf("setting up features: %w", err)
	}

	if syncCommands {
		if err := feature.SyncCommands(*def.features, client, guilds.IDs(development == "true")); err != nil {
			return fmt.Errorf("syncing commands: %w", err)
		}
	}

	if err := client.OpenGateway(ctx); err != nil {
		return fmt.Errorf("opening gateway: %w", err)
	}

	if enableCron {
		service.StartCron()
	}

	if enableOAuth2 {
		service.StartOAuth()
	}

	defer func() {
		if enableCron {
			service.StopCron()
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
