package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/config"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "log",
		Description: "Logs all messages",
		Handler: func(ctx bot.CommandContext) bool {
			if strings.HasPrefix(ctx.Message.Message, config.Get().Bot.Prefix) {
				return false
			}
			services.LoggerServiceInstance.Log(ctx.Message)
			return false
		},
	})
}
