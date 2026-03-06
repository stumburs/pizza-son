package commands

import (
	"log"
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	Register(bot.Command{
		Name:        "math",
		Description: "Responds as a math assistant.",
		Usage:       "!math <text>",
		Category:    bot.CategoryAI,
		Examples: []bot.CommandExample{
			{Input: "!math what's 9+10?", Output: "The result of 9+10 is 19."},
			{Input: "!math what's pi?", Output: "The value of Pi is approximately 3.14."},
		},
		Handler: func(ctx bot.CommandContext) {
			res, err := services.OllamaServiceInstance.GenerateChatResponse(ctx.Message, strings.Join(ctx.Args, " "))
			if err != nil {
				log.Println("[LLM]", err)
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, *res.Message.Content)
		},
	})
}
