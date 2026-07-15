package commands

import (
	"pizza-son/internal/bot"
	"strings"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "lmmlistener",
		Description: "corrects someone when they misspell !llm",
		Handler: func(ctx bot.CommandContext) bool {
			if !strings.Contains(strings.ToLower(ctx.Message.Text), "!lmm") {
				return false
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "it's !llm btw smh")
			return true
		},
	})
}
