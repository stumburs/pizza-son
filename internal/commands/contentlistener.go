package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "content",
		Description: "Times out people who say 'content'",
		Handler: func(ctx bot.CommandContext) bool {
			msg := strings.ToLower(ctx.Message.Text)
			if strings.Contains(msg, "content") {
				services.TwitchServiceInstance.Timeout(ctx.Message.Channel, ctx.Message.User.ID, 69, "don't use the C word rar")
				ctx.Client.Say(ctx.Message.Channel, "don't use the C word rar")
				return true
			}
			return false
		},
	})
}
