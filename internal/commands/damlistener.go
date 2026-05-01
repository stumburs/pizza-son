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
	RegisterListener(bot.ListenerEntry{
		Name:        "damn",
		Description: "responds with 'damn' when 'dam' appears as a standalone word",
		Handler: func(ctx bot.CommandContext) bool {
			words := strings.FieldsSeq(strings.ToLower(ctx.Message.Text))

			for w := range words {
				if w == "dam" {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "damn")
					return true
				}
			}

			return false
		},
	})
}
