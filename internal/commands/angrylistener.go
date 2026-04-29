package commands

import (
	"pizza-son/internal/bot"
	"strings"
)

var triggers = []string{"angry", "angy", "madge", "angri"}

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "rar",
		Description: "Responds with 'rar' to messages that contain 'angry' or 'angy'",
		Handler: func(ctx bot.CommandContext) bool {
			msg := strings.ToLower(ctx.Message.Text)
			for _, t := range triggers {
				if strings.Contains(msg, t) {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "rar")
					return true
				}
			}
			return false
		},
	})
}
