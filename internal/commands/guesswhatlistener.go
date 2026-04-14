package commands

import (
	"pizza-son/internal/bot"
	"strings"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "guesswhat",
		Description: "chicken butt's someone",
		Handler: func(ctx bot.CommandContext) bool {
			if !strings.Contains(strings.ToLower(ctx.Message.Message), "guess what") {
				return false
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "chicken butt")
			return true
		},
	})
}
