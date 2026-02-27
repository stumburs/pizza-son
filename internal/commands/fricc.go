package commands

import (
	"pizza-son/internal/bot"
	"strings"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "fricc",
		Description: "fricc's someone back",
		Handler: func(ctx bot.CommandContext) bool {
			if !strings.Contains(strings.ToLower(ctx.Message.Message), "fricc") {
				return false
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "fricc u too")
			return true
		},
	})
}
