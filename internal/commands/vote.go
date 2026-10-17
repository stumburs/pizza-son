package commands

import (
	"pizza-son/internal/bot"
)

func init() {
	Register(bot.Command{
		Name:        "vote",
		Description: "Vote for the best berts",
		Usage:       "!vote",
		Category:    bot.CategoryFun,
		Permission:  bot.All,
		Examples: []bot.CommandExample{
			{Input: "!vote", Output: "Vote for your favorite berts! https://pizzason.stumburs.id.lv/vote"},
		},
		Handler: func(ctx bot.CommandContext) {
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Vote for your favorite berts! https://pizzason.stumburs.id.lv/vote")
		},
	})
}
