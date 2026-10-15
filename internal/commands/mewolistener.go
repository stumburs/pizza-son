package commands

import (
	"pizza-son/internal/bot"
	"strings"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "mewo",
		Description: "mewo's someone back",
		Handler: func(ctx bot.CommandContext) bool {
			if !strings.Contains(strings.ToLower(ctx.Message.Text), "mewo") {
				return false
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "mewo")
			return true
		},
	})
}
