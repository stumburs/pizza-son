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
		Handler: func(ctx bot.CommandContext) bool {
			// Exclude Discord
			if ctx.Message.Platform == models.PlatformDiscord {
				return false
			}
			if strings.HasPrefix(ctx.Message.Text, config.Get().Bot.Prefix) {
				return false
			}
			services.LoggerServiceInstance.Log(ctx.Message)
			return false
		},
	})
}
