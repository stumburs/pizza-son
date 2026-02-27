package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
)

func init() {
	Register(bot.Command{
		Name:        "pee",
		Description: "Pees on the target user.",
		Permission:  bot.Moderator,
		Handler: func(ctx bot.CommandContext) {
			services.OllamaServiceInstance.Lobotomize(ctx.Message.Channel)
		},
	})
}
