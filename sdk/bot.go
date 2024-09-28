package sdk

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/paginator"
)

func NewBot(cfg Config, name string, version string, commit string) *Bot {
	return &Bot{
		Cfg:       cfg,
		Name:      name,
		Version:   version,
		Commit:    commit,
		Paginator: paginator.New(),
		Client:    nil,
	}
}

type Bot struct {
	Cfg       Config
	Name      string
	Version   string
	Commit    string
	Client    bot.Client
	Paginator *paginator.Manager
}

func (b *Bot) SetupBot(listeners []bot.EventListener) error {
	if (listeners == nil) || (len(listeners) == 0) {
		return fmt.Errorf("no command listener provided")
	}
	slog.Info(fmt.Sprintf("Setting up bot '%s' ...", b.Name))
	client, err := disgo.New(b.Cfg.Bot.Token,
		bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuilds, gateway.IntentGuildMessages, gateway.IntentMessageContent)),
		bot.WithCacheConfigOpts(cache.WithCaches(cache.FlagGuilds)),
		bot.WithEventListeners(b.Paginator),
		bot.WithEventListeners(bot.NewListenerFunc(b.onReady)),
		bot.WithEventListeners(listeners...),
	)
	if err != nil {
		return err
	}
	b.Client = client
	return nil
}

func (b *Bot) Start(syncCommands []discord.ApplicationCommandCreate) {

	slog.Info(fmt.Sprintf("Starting bot '%s' ...", b.Name))

	// Deferring client close
	slog.Info("Deferring client close")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b.Client.Close(ctx)
	}()

	// Sync commands if needed
	if syncCommands != nil {
		slog.Info("Syncing commands", slog.Any("guild_ids", b.Cfg.Bot.DevGuilds))
		if err := handler.SyncCommands(b.Client, syncCommands, b.Cfg.Bot.DevGuilds); err != nil {
			slog.Error("Failed to sync commands", slog.Any("err", err))
		}
	}

	// Open gateway
	slog.Info("Opening gateway")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := b.Client.OpenGateway(ctx); err != nil {
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

func (b *Bot) onReady(_ *events.Ready) {
	slog.Info(fmt.Sprintf("Bot '%s' is ready", b.Name))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := b.Client.SetPresence(ctx, gateway.WithListeningActivity("you"), gateway.WithOnlineStatus(discord.OnlineStatusOnline)); err != nil {
		slog.Error("Failed to set presence", slog.Any("err", err))
	}
}
