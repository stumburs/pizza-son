package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/models"
	"pizza-son/internal/services"
	"regexp"
	"strings"
)

var (
	// very basic
	basicUrlRegex = regexp.MustCompile(`(?i)https?:\\/S+|www\.\S+`)

	// AI save us (TODO: Improve and de-jankify)
	sixSevenRegex = regexp.MustCompile(`(?i)` +
		// "six" in any form: numeric (6/б), roman (VI), written (six/sick/sex/sicks), tens (sixty/sicksty)
		`(?:^|[\s,!?(])` +
		`(?:[6б]|VI|six|sick|sex|sicks|sixty|sicksty)` +
		// loose distance so 6 and 7 in the same sentence still count
		`.{0,50}` +
		// then "seven" in any form: 7, VII, seven
		`(?:7|VII|seven)`,
	)
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "67",
		Description: "Detects 67s in messages and ben's",
		Handler: func(ctx bot.CommandContext) bool {
			msg := strings.TrimSpace(ctx.Message.Text)

			// strip URLs
			msg = basicUrlRegex.ReplaceAllString(msg, "")

			if !sixSevenRegex.MatchString(msg) {
				return false
			}

			ctx.Client.Say(ctx.Message.Channel, "ben")

			// Only timeout on Twitch
			if ctx.Message.Platform == models.PlatformTwitch {
				services.TwitchServiceInstance.Timeout(ctx.Message.Channel, ctx.Message.User.ID, 69, "ben")
			}
			return true
		},
	})
}
