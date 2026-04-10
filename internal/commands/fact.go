package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
)

func init() {
	Register(bot.Command{
		Name:        "fact",
		Description: "Responds with a random fact.",
		Usage:       "!fact",
		Category:    bot.CategoryFun,
		Permission:  bot.All,
		Examples: []bot.CommandExample{
			{Input: "!fact", Output: "Insert random fact here."},
		},
		Handler: func(ctx bot.CommandContext) {
			fact, err := services.GetFact()
			if err != nil {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Failed to get fact: "+err.Error())
				return
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fact.Text)
		},
	})
}
