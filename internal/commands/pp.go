package commands

import (
	"fmt"
	"math/rand/v2"
	"pizza-son/internal/bot"
)

func init() {
	Register(bot.Command{
		Name:        "pp",
		Description: "Shows the pp size of target user.",
		Usage:       "!pp [user]",
		Category:    bot.CategoryFun,
		Examples: []bot.CommandExample{
			{Input: "!pp", Output: "pizza_tm has 34cm kok SpringlesLong"},
			{Input: "!pp @big_daddy69", Output: "big_daddy69 has 1cm kok SpringlesLong"},
		},
		Handler: func(ctx bot.CommandContext) {
			target := ctx.Message.User.DisplayName
			if len(ctx.Args) > 0 {
				target = ctx.Args[0]
			}
			ppSize := rand.IntN(50)

			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("%s has %dcm kok SpringlesLong", target, ppSize))
		},
	})
}
