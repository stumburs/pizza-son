package commands

import (
	"fmt"
	"math/rand/v2"
	"pizza-son/internal/bot"
)

func init() {
	Register(bot.Command{
		Name:        "pee",
		Description: "Pees on the target user.",
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
