package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "log-listener",
		Description: "Logs all messages",
		Handler: func(ctx bot.CommandContext) bool {
			services.LoggerServiceInstance.Log(ctx.Message)
			return false
		},
	})
}
