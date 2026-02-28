package commands

import (
	"fmt"
	"math/rand/v2"
	"pizza-son/internal/bot"
)

func init() {
	Register(bot.Command{
		Name:        "hug",
		Description: "Hugs the target user.",
		Usage:       "!hug [user]",
		Handler: func(ctx bot.CommandContext) {
			target := ctx.Message.User.DisplayName
			loveAmount := rand.IntN(100)
			if len(ctx.Args) > 0 {
				target = ctx.Args[0]
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("%s hugs %s with %d%% love :3", ctx.Message.User.DisplayName, target, loveAmount))
				return
			} else {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("%s hugs themselves with %d%% love... what a loser smh", ctx.Message.User.DisplayName, loveAmount))
			}
		},
	})
}
