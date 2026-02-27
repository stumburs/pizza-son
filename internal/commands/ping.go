package commands

import "pizza-son/internal/bot"

func init() {
	Register(bot.Command{
		Name:        "ping",
		Description: "Responds with Pong!",
		Handler: func(ctx bot.CommandContext) {
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Pong!")
		},
	})
}
