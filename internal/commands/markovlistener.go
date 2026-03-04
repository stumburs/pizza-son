package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/config"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "markov-learn",
		Description: "Feeds chat messages into Markov generator",
		Handler: func(ctx bot.CommandContext) bool {
			if strings.HasPrefix(ctx.Message.Message, config.Get().Bot.Prefix) {
				return false
			}
			go services.MarkovServiceInstance.Learn(ctx.Message.Channel, ctx.Message.Message)
			return true
		},
	})
}
