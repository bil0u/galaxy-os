package sdk

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/disgoorg/disgo/bot"
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
	Name      string
	Version   string
	Commit    string
	Cfg       Config
	Client    bot.Client
	Paginator *paginator.Manager
}

// SetupBot sets up the bot with the provided parts
func (b *Bot) SetupBot(parts BotParts) error {

	// Add default listeners
	b.Client.AddEventListeners(b.Paginator)
	b.Client.AddEventListeners(bot.NewListenerFunc(b.OnReady))

	// Create router and register it as an event listener
	router := parts.CreateRouter(b)
	b.Client.AddEventListeners(router)

	// Create bot listeners
	listeners := parts.CreateListeners(b)
	b.Client.AddEventListeners(listeners...)

	return nil
}

func (b *Bot) Start(syncCommands []discord.ApplicationCommandCreate, syncRoles bool) {

	slog.Info(fmt.Sprintf("Starting bot '%s' ...", b.Name))

	// Deferring client close
	slog.Info("Deferring client close")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b.Client.Close(ctx)
	}()

	// Sync roles if needed
	if syncRoles {
		for _, guildID := range b.Cfg.Bot.GetGuildsToSync() {
			guildRoles := b.Cfg.Bot.GetGuildRoles(guildID)
			err := AssignRolesToBot(b.Client, guildID, guildRoles)
			slog.Info(fmt.Sprintf("Syncing roles for guild '%s'", guildID), slog.Any("roles", guildRoles))
			if err != nil {
				slog.Error("Error assigning role to bot:", slog.Any("err", err))
			}
		}
	}

	// Sync commands if needed
	if syncCommands != nil {
		guilds := b.Cfg.Bot.GetGuildsToSync()
		slog.Info("Syncing commands", slog.Any("guilds", guilds))
		if err := handler.SyncCommands(b.Client, syncCommands, guilds); err != nil {
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

func (b *Bot) OnReady(_ *events.Ready) {
	slog.Info(fmt.Sprintf("Bot '%s' is ready", b.Name))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := b.Client.SetPresence(ctx, gateway.WithListeningActivity("you"), gateway.WithOnlineStatus(discord.OnlineStatusOnline)); err != nil {
		slog.Error("Failed to set presence", slog.Any("err", err))
	}
}
