package commands

import (
	"pizza-son/internal/bot"
)

func init() {
	Register(bot.Command{
		Name:        "mark",
		Description: "Responds using a markov chain algorithm trained on this chat.",
		Handler: func(ctx bot.CommandContext) {
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Not yet implemented :(")
		},
	})
}
