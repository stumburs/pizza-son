package commands

import (
	"fmt"
	"pizza-son/internal/bot"
)

func init() {
	Register(bot.Command{
		Name:        "perm",
		Description: "Returns the lowest permission level for calling user.",
		Usage:       "!perm",
		Handler: func(ctx bot.CommandContext) {
			level := bot.GetPermissionLevel(ctx.Message)
			name := bot.PermissionName(level)
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Your permission level is: %s", name))
		},
	})
}
