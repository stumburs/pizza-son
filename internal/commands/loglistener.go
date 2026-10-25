package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/config"
	"pizza-son/internal/models"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "log",
		Description: "Logs all messages",
		Passive:     true,
		Handler: func(ctx bot.CommandContext) bool {
			// Exclude Discord
			if ctx.Message.Platform == models.PlatformDiscord {
				return false
			}
			if strings.HasPrefix(ctx.Message.Text, config.Get().Bot.Prefix) {
				return false
			}
			services.LoggerServiceInstance.Log(ctx.Message)
			// The activity service already records every message, so the
			// logging listener is not reported as a listener activation.
			return false
		},
	})
}
