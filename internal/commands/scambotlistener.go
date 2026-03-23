package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
)

var bannableTerms = []string{
	"streamboo .com",
	"streamboo .live",
}

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "scam-banner",
		Description: "Detects scam bots and automatically bans them.",
		Handler: func(ctx bot.CommandContext) bool {
			if !ctx.Message.FirstMessage {
				return false
			}
			for _, term := range bannableTerms {
				if strings.Contains(ctx.Message.Message, term) {
					services.TwitchServiceInstance.Ban(ctx.Message.Channel, ctx.Message.User.ID, "ben")
					ctx.Client.Say(ctx.Message.Channel, "ben")
					return true
				}
			}
			return false
		},
	})
}
