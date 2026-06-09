package commands

import (
	"log"
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	Register(bot.Command{
		Name:        "bert",
		Description: "Responds as bert.",
		Usage:       "!bert <text>",
		Category:    bot.CategoryAI,
		Examples: []bot.CommandExample{
			{Input: "!bert how are you bert?", Output: "bert is very cozy today 00 found a warm sunbeam and refuses to leave it :3"},
			{Input: "!bert what's 2+2?", Output: "bert counted on tiny orange paws... 4! 00"},
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
