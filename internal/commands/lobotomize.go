package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
)

func init() {
	Register(bot.Command{
		Name:        "lobotomize",
		Description: "Clears the LLM's memory in this channel.",
		Usage:       "!lobotomize",
		Permission:  bot.Moderator,
		Handler: func(ctx bot.CommandContext) {
			services.OllamaServiceInstance.Lobotomize(ctx.Message.Channel)
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "meow")
		},
	})
}
