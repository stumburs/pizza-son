package commands

import (
	"fmt"
	"math/rand/v2"
	"pizza-son/internal/bot"
)

func init() {
	Register(bot.Command{
		Name:        "pee",
		Description: "Pees on the target user. Note, for optimal enjoyment, you should add certain 7TV emotes.",
		Usage:       "!pee [user]",
		Category:    bot.CategoryFun,
		Examples: []bot.CommandExample{
			{Input: "!pee", Output: "pizza_tm PEE s all over themselves with 84% flavor smh"},
			{Input: "!pee @water_enjoyer", Output: "pizza_tm PEE s on borpaLick water_enjoyer borpaLickL with 100% flavor LICKA"},
		},
		Handler: func(ctx bot.CommandContext) {
			target := ctx.Message.User.DisplayName
			flavor := rand.IntN(100)
			if len(ctx.Args) > 0 {
				target = ctx.Args[0]
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("%s PEE s on borpaLick %s borpaLickL with %d%% flavor LICKA", ctx.Message.User.DisplayName, target, flavor))
				return
			} else {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("%s PEE s all over themselves with %d%% flavor smh", ctx.Message.User.DisplayName, flavor))
			}
		},
	})
}
