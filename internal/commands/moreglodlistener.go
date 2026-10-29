package commands

import (
	"pizza-son/internal/bot"
	"strings"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "moreglodlistener",
		Description: "Witty reply to someone asking \"more glod pls\"",
		Handler: func(ctx bot.CommandContext) bool {
			if !strings.Contains(strings.ToLower(ctx.Message.Text), "more glod pls") {
				return false
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "goldenzazabert but only because you asked nicely 0")
			return true
		},
	})
}
