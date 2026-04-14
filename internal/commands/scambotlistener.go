package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/models"
	"pizza-son/internal/services"
	"strings"
)

var bannableTerms = []string{
	"streamboo .com",
	"streamboo .live",
	"hypesm.online",
}

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "scam-banner",
		Description: "Detects scam bots and automatically bans them.",
		Handler: func(ctx bot.CommandContext) bool {
			// Exclude Discord
			if ctx.Message.Platform == models.PlatformDiscord {
				return false
			}
			if !ctx.Message.FirstMessage {
				return false
			}
			for _, term := range bannableTerms {
				if strings.Contains(ctx.Message.Text, term) {
					services.TwitchServiceInstance.Ban(ctx.Message.Channel, ctx.Message.User.ID, "ben")
					ctx.Client.Say(ctx.Message.Channel, "ben")
					return true
				}
			}
			return false
		},
	})
}
