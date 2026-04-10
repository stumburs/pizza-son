package commands

import (
	"log"
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	Register(bot.Command{
		Name:        "joni",
		Description: "Responds using the almighty Jon's prompt.",
		Usage:       "!joni <text>",
		Category:    bot.CategoryAI,
		Examples: []bot.CommandExample{
			{Input: "!joni hello", Output: "Imagine some thoughtful response here."},
			{Input: "!joni should I get a cat?", Output: "Imagine some thoughtful response here."},
		},
		Handler: func(ctx bot.CommandContext) {
			res, err := services.OllamaServiceInstance.GenerateChatResponse(ctx.Message, strings.Join(ctx.Args, " "))
			if err != nil {
				log.Println("[LLM]", err)
			}
			clean := strings.Join(strings.Fields(*res.Message.Content), " ")
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, clean)
		},
	})
}
