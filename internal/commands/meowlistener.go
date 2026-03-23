package commands

import (
	"pizza-son/internal/bot"
	"strings"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "meow",
		Description: "meow's someone back",
		Handler: func(ctx bot.CommandContext) bool {
			if !strings.Contains(strings.ToLower(ctx.Message.Message), "meow") {
				return false
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "meow")
			return true
		},
	})
}
