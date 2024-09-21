package bots

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bil0u/galaxy-os/sdk"
	"github.com/bil0u/galaxy-os/sdk/handlers"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
)

type routerInitializer func(*sdk.Bot, *handler.Mux) []discord.ApplicationCommandCreate

func Run(botName string, version string, commit string, initRouter routerInitializer) {
	// Parse flags
	shouldSyncCommands := flag.Bool("sync-commands", false, "Whether to sync commands to discord")
	configPath := flag.String("config", fmt.Sprintf("config.%s.toml", botName), "path to config")
	flag.Parse()

	// Load config file
	cfg, err := sdk.LoadConfig(*configPath)
	if err != nil {
		slog.Error("Failed to read config", slog.Any("err", err))
		os.Exit(-1)
	}

	// Setup logger
	sdk.SetupLogger(cfg.Log)
	slog.Info(fmt.Sprintf("Starting bot '%s'", botName), slog.String("version", version), slog.String("commit", commit))
	slog.Info("Syncing commands", slog.Bool("sync", *shouldSyncCommands))

	// Create bot
	b := sdk.NewBot(*cfg, botName, version, commit)

	// Setup handlers
	h := handler.New()

	// Execute register handler and get available commands
	availableCommands := initRouter(b, h)

	// Setup bot
	if err = b.SetupBot(h, bot.NewListenerFunc(b.OnReady), handlers.MessageHandler(b)); err != nil {
		slog.Error("Failed to setup bot", slog.Any("err", err))
		os.Exit(-1)
	}

	// Defer close
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b.Client.Close(ctx)
	}()

	// Sync commands if needed
	if *shouldSyncCommands {
		slog.Info("Syncing commands", slog.Any("guild_ids", cfg.Bot.DevGuilds))
		if err = handler.SyncCommands(b.Client, availableCommands, cfg.Bot.DevGuilds); err != nil {
			slog.Error("Failed to sync commands", slog.Any("err", err))
		}
	}

	// Open gateway
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = b.Client.OpenGateway(ctx); err != nil {
		slog.Error("Failed to open gateway", slog.Any("err", err))
		os.Exit(-1)
	}

	// Wait for signal to shutdown
	slog.Info("Bot is running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
	slog.Info("Shutting down bot...")
}
