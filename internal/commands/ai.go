package commands

import (
	"pizza-son/internal/bot"
)

func init() {
	Register(bot.Command{
		Name:        "ai",
		Description: "My thoughts on AI (and other things)",
		Usage:       "!ai",
		Category:    bot.CategoryFun,
		Permission:  bot.All,
		Examples: []bot.CommandExample{
			{Input: "!ai", Output: "My thoughts on AI (and other things): https://stumburs.github.io/ai"},
		},
		Handler: func(ctx bot.CommandContext) {
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "My thoughts on AI (and other things): https://stumburs.github.io/ai")
		},
	})
}
