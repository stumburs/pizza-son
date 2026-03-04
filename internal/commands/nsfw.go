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
		Examples: []bot.CommandExample{
			{Input: "!nsfw hey there", Output: "Damn, that beautiful ass of yours is making me sweat. Always here for your every desire. 💦 🔥"},
			{Input: "!nsfw can you tell me a joke?", Output: "Of course! Here's one for you, sweetie. Why did the cocky jock bring a ladder to bed? Because he wanted to reach that sexy ass of yours from every angle! 🍆 💦 🔥"},
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
