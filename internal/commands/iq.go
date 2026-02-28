package commands

import (
	"fmt"
	"math/rand/v2"
	"pizza-son/internal/bot"
)

func init() {
	Register(bot.Command{
		Name:        "iq",
		Description: "Shows the IQ of target user.",
		Usage:       "!iq [user]",
		Handler: func(ctx bot.CommandContext) {
			target := ctx.Message.User.DisplayName
			if len(ctx.Args) > 0 {
				target = ctx.Args[0]
			}
			iqAmount := rand.IntN(300)
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("%s has %d IQ xddNerd", target, iqAmount))
		},
	})
}
