package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/config"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "ollama",
		Description: "Feeds chat messages to LLM context",
		Handler: func(ctx bot.CommandContext) bool {
			if strings.HasPrefix(ctx.Message.Text, config.Get().Bot.Prefix) {
				return false
			}
			go services.OllamaServiceInstance.OnPrivateMessage(ctx.Message)
			return true
		},
	})
}
