package commands

import (
	"pizza-son/internal/bot"
)

func init() {
	Register(bot.Command{
		Name:        "help",
		Description: "there is no help",
		Usage:       "!help",
		Category:    bot.CategoryUtility,
		Permission:  bot.All,
		Examples: []bot.CommandExample{
			{Input: "!help", Output: "there is no help"},
		},
		Handler: func(ctx bot.CommandContext) {
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "there is no help")
		},
	})
}
