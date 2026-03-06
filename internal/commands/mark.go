package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
)

func init() {
	Register(bot.Command{
		Name:        "mark",
		Description: "Responds using a markov chain algorithm trained on this chat.",
		Usage:       "!mark [text]",
		Category:    bot.CategoryAI,
		Examples: []bot.CommandExample{
			{Input: "!mark", Output: "t kept calling and I can time , makes serious it ok if i can mean Kappa"},
		},
		Handler: func(ctx bot.CommandContext) {
			text := services.MarkovServiceInstance.Generate(ctx.Message.Channel)
			if text == "" {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Not enough data yet!")
				return
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, text)
		},
	})
}
