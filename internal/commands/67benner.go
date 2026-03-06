package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "67-benner",
		Description: "Detects 67s in messages and ben's",
		Handler: func(ctx bot.CommandContext) bool {
			if strings.TrimSpace(ctx.Message.Message) == "67" {
				ctx.Client.Say(ctx.Message.Channel, "ben")
				services.TwitchServiceInstance.Timeout(ctx.Message.Channel, ctx.Message.User.ID, 69, "ben")
				return true
			}
			return false
		},
	})
}
