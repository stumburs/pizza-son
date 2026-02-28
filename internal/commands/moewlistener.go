package commands

import (
	"pizza-son/internal/bot"
	"strings"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "moew",
		Description: "moew's someone back",
		Handler: func(ctx bot.CommandContext) bool {
			if !strings.Contains(strings.ToLower(ctx.Message.Message), "moew") {
				return false
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "moew")
			return true
		},
	})
}
