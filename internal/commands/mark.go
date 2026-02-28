package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
)

func init() {
	Register(bot.Command{
		Name:        "mark",
		Description: "Responds using a markov chain algorithm trained on this chat.",
		Usage:       "!mark [text]",
		Handler: func(ctx bot.CommandContext) {
			text := services.MarkovServiceInstance.Generate(ctx.Message.Channel)
			if text == "" {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Not enough data yet!")
				return
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, text)
		},
	})
}
