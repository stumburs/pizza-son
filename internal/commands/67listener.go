package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"regexp"
	"strings"
)

var (
	// very basic
	basicUrlRegex = regexp.MustCompile(`(?i)https?:\\/S+|www\.\S+`)

	// AI save us (TODO: Improve and de-jankify)
	sixSevenRegex = regexp.MustCompile(`(?i)` +
		// Numeric: 67, 6-7, 6 uh 7, 67676767, etc.
		`\b[6б].{0,20}7\b` +
		`|` +
		// Mixed: six...7, 6...seven
		`\bsix.{0,20}7\b` +
		`|` +
		`\b[6б].{0,20}seven\b` +
		`|` +
		// Written: six...seven, sixty...seven
		`six.{0,20}seven` +
		`|` +
		`sixty.{0,20}seven`,
	)
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "67",
		Description: "Detects 67s in messages and ben's",
		Handler: func(ctx bot.CommandContext) bool {
			msg := strings.TrimSpace(ctx.Message.Message)

			// strip URLs
			msg = basicUrlRegex.ReplaceAllString(msg, "")

			if !sixSevenRegex.MatchString(msg) {
				return false
			}
			ctx.Client.Say(ctx.Message.Channel, "ben")
			services.TwitchServiceInstance.Timeout(ctx.Message.Channel, ctx.Message.User.ID, 69, "ben")
			return true
		},
	})
}
