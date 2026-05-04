package commands

import (
	"pizza-son/internal/bot"
)

func init() {
	Register(bot.Command{
		Name:        "bertle",
		Description: "Returns link to bertle game.",
		Usage:       "!bertle",
		Category:    bot.CategoryFun,
		Permission:  bot.All,
		Examples: []bot.CommandExample{
			{Input: "!bertle", Output: "Play bertle here: https://stumburs.id.lv"},
		},
		Handler: func(ctx bot.CommandContext) {
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Play bertle here: https://stumburs.id.lv")
		},
	})
}
