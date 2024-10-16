package bot_presence

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bil0u/galaxy-os/sdk"
	disbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
)

func SetPresenceWhenReady(bot *sdk.Bot) disbot.EventListener {
	return disbot.NewListenerFunc(func(_ *events.Ready) {
		slog.Info(fmt.Sprintf("Bot '%s' is ready", bot.Name))
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := bot.Client.SetPresence(ctx, gateway.WithCustomActivity("Loading Kernel..."), gateway.WithOnlineStatus(discord.OnlineStatusOnline)); err != nil {
			slog.Error("Failed to set presence", slog.Any("err", err))
		}
	})
}
