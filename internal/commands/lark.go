package commands

import (
	"log"
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
)

func init() {
	Register(bot.Command{
		Name:        "lark",
		Description: "Reinterprets a message from !mark using an LLM.",
		Usage:       "!lark [text]",
		Examples: []bot.CommandExample{
			{Input: "!lark", Output: "the cats will have keys of backwards maps"},
			{Input: "!lark should I get a cat?", Output: "laugh out loud. my friend just did the unthinkable."},
		},
		Handler: func(ctx bot.CommandContext) {
			markMsg := services.MarkovServiceInstance.Generate(ctx.Message.Channel)
			res, err := services.OllamaServiceInstance.GenerateChatResponse(ctx.Message, markMsg)
			if err != nil {
				log.Println("[LLM]", err)
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, *res.Message.Content)
		},
	})
}
