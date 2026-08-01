package commands

import (
	"pizza-son/internal/bot"
	"strings"
)

func init() {

	RegisterListener(bot.ListenerEntry{
		Name:        "firsttimeben",
		Description: "Detects first time ben'ers",
		Handler: func(ctx bot.CommandContext) bool {
			if !ctx.Message.FirstMessage {
				return false
			}

			msg := strings.ToLower(ctx.Message.Text)
			matched := false

			if strings.Contains(msg, "ben") {
				matched = true
			}
			if matched {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "FirstTimeBen")
			}

			return false
		},
	})
}
