package commands

import (
	"pizza-son/internal/bot"
)

func init() {
	Register(bot.Command{
		Name:        "support",
		Description: "Give child support to pizza. http://ko-fi.com/stumburs/tiers",
		Usage:       "!support",
		Category:    bot.CategoryUncategorized,
		Examples: []bot.CommandExample{
			{Input: "!support", Output: "Give child support to pizza. http://ko-fi.com/stumburs/tiers"},
		},
		Handler: func(ctx bot.CommandContext) {
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Give child support to pizza. http://ko-fi.com/stumburs/tiers")
		},
	})
}
