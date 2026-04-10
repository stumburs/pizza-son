package commands

import (
	"log"
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	Register(bot.Command{
		Name:        "llm",
		Description: "Responds as a basic LLM.",
		Usage:       "!llm <text>",
		Category:    bot.CategoryAI,
		Examples: []bot.CommandExample{
			{Input: "!llm hello!", Output: "Hello, there, pizza_tm."},
			{Input: "!llm what's the current stream title?", Output: "The current stream title is: 'Playing with kitties'"},
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
