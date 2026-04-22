package commands

import (
	"pizza-son/internal/bot"
	"strings"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "dam",
		Description: "responds with 'dam' whenever someone types 'damn'",
		Handler: func(ctx bot.CommandContext) bool {
			if !strings.Contains(strings.ToLower(ctx.Message.Text), "damn") {
				return false
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "dam")
			return true
		},
	})
}
