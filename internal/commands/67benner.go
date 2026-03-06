package commands

import (
	"pizza-son/internal/bot"
	"strings"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "67-benner",
		Description: "Detects 67s in messages and ben's",
		Handler: func(ctx bot.CommandContext) bool {
			if strings.Contains(ctx.Message.Message, "67") {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "ben")
				return true
			}
			return false
		},
	})
}
