package cmd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bil0u/galaxy-os/internal/config"
	discordadapter "github.com/bil0u/galaxy-os/internal/discord"
	"github.com/bil0u/galaxy-os/internal/feature"
	"github.com/bil0u/galaxy-os/internal/platform"
	"github.com/bil0u/galaxy-os/internal/service"
	"github.com/disgoorg/disgo"
	disbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/spf13/cobra"
)

var syncCommands bool

var botCmd = &cobra.Command{
	Use:   "bot",
	Short: "Main command for the bot",
}

func init() {
	botCmd.AddCommand(startCmd)

	startCmd.Flags().BoolVarP(&syncCommands, "sync", "s", false, "Sync slash commands to Discord before starting")
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the discord bot",
	RunE:  startBot,
}

type botDef struct {
	features   []platform.Feature
	cacheFlags cache.Flags
	intents    gateway.Intents
}

var bots = map[string]botDef{
	"hue": {
		features: []platform.Feature{
			feature.Info,
			feature.Permissions,
			feature.Test,
			feature.Presence,
			feature.SelfAssign,
			feature.DailyMessage,
			feature.SuspiciousInterview,
		},
		cacheFlags: cache.FlagGuilds | cache.FlagMembers | cache.FlagRoles,
		intents:    gateway.IntentGuilds | gateway.IntentGuildMembers,
	},
	"kevin": {
		features: []platform.Feature{
			feature.Info,
			feature.Permissions,
			feature.Test,
			feature.Presence,
			feature.SelfAssign,
			feature.DailyMessage,
		},
		cacheFlags: cache.FlagGuilds | cache.FlagRoles,
		intents:    gateway.IntentGuilds,
	},
}

// muxRegistrar adapts handler.Mux to platform.Registrar.
// disgo v0.19.3 handler types include a typed data parameter;
// the platform.Registrar signatures omit it for simplicity.
type muxRegistrar struct {
	mux *handler.Mux
}

func (r *muxRegistrar) SlashCommand(path string, h platform.SlashCommandHandler) {
	r.mux.SlashCommand(path, func(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
		return h(e)
	})
}

func (r *muxRegistrar) ButtonComponent(customID string, h platform.ButtonComponentHandler) {
	r.mux.ButtonComponent(customID, func(_ discord.ButtonInteractionData, e *handler.ComponentEvent) error {
		return h(e)
	})
}

func (r *muxRegistrar) Autocomplete(path string, h platform.AutocompleteHandler) {
	r.mux.Autocomplete(path, func(e *handler.AutocompleteEvent) error {
		return h(e)
	})
}

func startBot(_ *cobra.Command, _ []string) error {
	def, ok := bots[bot]
	if !ok {
		return fmt.Errorf("bot %q not found", bot)
	}
	if len(def.features) == 0 {
		return fmt.Errorf("no features defined for bot %q", bot)
	}

	// Capture closed-over state needed by stage closures.
	// These are populated during stage execution and shared across stages.
	var (
		globalCfg *config.Global
		logCfg    *config.Log
		botCfg    *config.Bot
		guilds    *config.GuildMap
		botLogger *slog.Logger
		registry  *service.Registry
		resolver  *config.Resolver
	)

	stages := []platform.Stage{
		{
			Name:     "config",
			Provides: []string{"config"},
			Run: func(_ context.Context, state *platform.BootState) error {
				var err error
				globalCfg, logCfg, botCfg, err = config.Init(bot)
				if err != nil {
					return fmt.Errorf("initializing config: %w", err)
				}
				globalCfg.Development = development == "true"
				globalCfg.Version = version
				globalCfg.Commit = commit

				state.Config.Global = globalCfg
				state.Config.Bot = botCfg
				state.Config.Log = logCfg
				return nil
			},
		},
		{
			Name:     "logger",
			Requires: []string{"config"},
			Provides: []string{"logger"},
			Run: func(_ context.Context, state *platform.BootState) error {
				logger, err := service.InitLogger(logCfg.Level, logCfg.Format, logCfg.AddSource)
				if err != nil {
					return fmt.Errorf("initializing logger: %w", err)
				}
				botLogger = slog.Default().With("bot", bot)
				state.Obs = &platform.Observability{Logger: logger}
				return nil
			},
		},
		{
			Name:     "discord",
			Requires: []string{"config", "logger"},
			Provides: []string{"discord"},
			Run: func(_ context.Context, state *platform.BootState) error {
				client, err := disgo.New(botCfg.Token,
					disbot.WithCacheConfigOpts(cache.WithCaches(def.cacheFlags)),
					disbot.WithGatewayConfigOpts(
						gateway.WithIntents(def.intents),
					),
				)
				if err != nil {
					return fmt.Errorf("building discord client: %w", err)
				}

				router := handler.New()
				client.AddEventListeners(router)
				client.AddEventListeners(disbot.NewListenerFunc(func(_ *events.Resumed) {
					botLogger.Info("gateway reconnected")
				}))

				state.Client = client
				state.Router = router
				return nil
			},
		},
		{
			Name:     "guilds",
			Requires: []string{"discord"},
			Provides: []string{"guilds"},
			Run: func(ctx context.Context, state *platform.BootState) error {
				var err error
				guilds, err = config.InitGuilds(ctx, state.Client.Rest, bot)
				if err != nil {
					return fmt.Errorf("initializing guilds config: %w", err)
				}

				slog.Debug(fmt.Sprintf("Global configuration: %+v", globalCfg))
				slog.Debug(fmt.Sprintf("Bot configuration: %++v", botCfg))
				slog.Debug(fmt.Sprintf("Log configuration: %+v", logCfg))
				slog.Debug(fmt.Sprintf("Guilds configuration: %+v", guilds))

				store := config.NewFileStore(".")
				resolver = config.NewResolver(store, bot)
				guildIDs := guilds.IDs(development == "true")
				resolver.SetGuildIDs(guildIDs)
				state.GuildIDs = guildIDs
				return nil
			},
		},
		{
			Name:     "services",
			Requires: []string{"config", "guilds"},
			Provides: []string{"services"},
			Run: func(ctx context.Context, state *platform.BootState) error {
				registry = service.NewRegistry()
				registry.Register(platform.CronService, service.NewCronService())

				if err := botCfg.ValidateOAuth(); err == nil {
					registry.Register(platform.OAuthService,
						service.NewOAuthService(botCfg.ApplicationID, botCfg.ClientSecret, botCfg.BaseURL))
				}

				for _, f := range def.features {
					registry.Require(f.Needs()...)
				}

				if err := registry.StartAll(ctx); err != nil {
					return fmt.Errorf("starting services: %w", err)
				}

				state.Services = registry
				return nil
			},
		},
		{
			Name:     "features",
			Requires: []string{"discord", "guilds", "services"},
			Provides: []string{"features"},
			Run: func(_ context.Context, state *platform.BootState) error {
				registrar := &muxRegistrar{mux: state.Router}
				restClient := discordadapter.NewRestAdapter(state.Client.Rest)
				localeResolver := discordadapter.NewLocaleResolver(discord.LocaleEnglishUS)

				var cronScheduler platform.CronScheduler
				if svc := registry.Service(platform.CronService); svc != nil {
					cronScheduler = svc.(*service.CronService)
				}
				var oauthProvider platform.OAuthProvider
				if svc := registry.Service(platform.OAuthService); svc != nil {
					oauthProvider = svc.(*service.OAuthService)
				}

				var errs []error
				for _, f := range def.features {
					cfgProvider := config.NewProvider(resolver, f.Name())
					switch f.Scope() {
					case platform.BotScope:
						deps := platform.Deps{
							Logger:   botLogger.With("feature", f.Name()),
							Rest:     restClient,
							Commands: registrar,
							Locale:   localeResolver,
							Configs:  cfgProvider,
							BotName:  bot,
							Cron:     cronScheduler,
							OAuth:    oauthProvider,
						}
						if err := f.Setup(deps); err != nil {
							errs = append(errs, fmt.Errorf("setting up feature %q: %w", f.Name(), err))
						}
					case platform.GuildScope:
						for _, guildID := range state.GuildIDs {
							deps := platform.Deps{
								Logger:   botLogger.With("feature", f.Name(), "guild", guildID),
								Rest:     restClient,
								Commands: registrar,
								Locale:   localeResolver,
								Configs:  cfgProvider,
								BotName:  bot,
								GuildID:  guildID,
								Cron:     cronScheduler,
								OAuth:    oauthProvider,
							}
							if err := f.Setup(deps); err != nil {
								errs = append(errs, fmt.Errorf("setting up feature %q for guild %s: %w", f.Name(), guildID, err))
							}
						}
					case platform.CrossGuildScope:
						deps := platform.Deps{
							Logger:   botLogger.With("feature", f.Name()),
							Rest:     restClient,
							Commands: registrar,
							Locale:   localeResolver,
							Configs:  cfgProvider,
							BotName:  bot,
							Cron:     cronScheduler,
							OAuth:    oauthProvider,
						}
						if err := f.Setup(deps); err != nil {
							errs = append(errs, fmt.Errorf("setting up feature %q: %w", f.Name(), err))
						}
					}
				}
				if err := errors.Join(errs...); err != nil {
					return fmt.Errorf("setting up features: %w", err)
				}

				state.Features = def.features

				if syncCommands {
					slog.Warn("command syncing not yet implemented with new feature framework")
				}
				return nil
			},
		},
		{
			Name:     "gateway",
			Requires: []string{"features"},
			Provides: []string{"gateway"},
			Run: func(ctx context.Context, state *platform.BootState) error {
				if err := state.Client.OpenGateway(ctx); err != nil {
					return fmt.Errorf("opening gateway: %w", err)
				}
				return nil
			},
		},
		{
			Name:     "ready",
			Requires: []string{"gateway"},
			Run: func(ctx context.Context, state *platform.BootState) error {
				slog.Info("----- Bot is running. Press CTRL-C to exit -----")

				sig := make(chan os.Signal, 1)
				signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

				select {
				case <-sig:
				case <-ctx.Done():
				}

				slog.Info("shutting down bot...")

				shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				if err := platform.Shutdown(shutdownCtx, state); err != nil {
					slog.Error("shutdown completed with errors", slog.Any("error", err))
				}
				return nil
			},
		},
	}

	ctx := context.Background()
	_, err := platform.Run(ctx, platform.BotSpec{
		Name:     bot,
		Features: def.features,
		Intents:  def.intents,
	}, stages)

	return err
}
