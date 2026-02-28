package commands

import (
	"log"
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	Register(bot.Command{
		Name:        "nsfw",
		Description: "Responds as a spicy LLM ;)",
		Usage:       "!nsfw <text>",
		Handler: func(ctx bot.CommandContext) {
			res, err := services.OllamaServiceInstance.GenerateChatResponse(ctx.Message, strings.Join(ctx.Args, " "))
			if err != nil {
				log.Println("[LLM]", err)
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, *res.Message.Content)
		},
	})
}
