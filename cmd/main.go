package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bil0u/galaxy-os/cmd/bots"
	"github.com/bil0u/galaxy-os/sdk"
	"github.com/bil0u/galaxy-os/sdk/handlers"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/handler"
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

	var routerInit sdk.RouterInitializer

	// Checking if the provided bot name is valid
	switch botName {
	case "hue":
		routerInit = bots.HueRouter
	case "kevin":
		routerInit = bots.KevinRouter
	default:
		slog.Error("Unknown bot", slog.String("bot", botName))
		os.Exit(-1)
	}

	slog.Info(fmt.Sprintf("Starting bot '%s'", botName), slog.String("version", version), slog.String("commit", commit))

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

	// Setup handler
	h := handler.New()

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

	// Execute register handler and get available commands
	availableCommands := routerInit(b, h)

	// Sync commands if needed
	if f.syncCommands {
		slog.Info("Syncing commands", slog.Any("guild_ids", cfg.Bot.DevGuilds))
		if err := handler.SyncCommands(b.Client, availableCommands, cfg.Bot.DevGuilds); err != nil {
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
