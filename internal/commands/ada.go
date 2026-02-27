package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	Register(bot.Command{
		Name:        "ada",
		Description: "Chat with Ada, the learning chatbot.",
		Handler: func(ctx bot.CommandContext) {
			if len(ctx.Args) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Say something to Ada!")
				return
			}
			response := services.AdaServiceInstance.GetResponse(ctx.Message.Channel, strings.Join(ctx.Args, " "))
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, response)
		},
	})
}
