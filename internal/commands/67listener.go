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

	sixSevenRegex = regexp.MustCompile(`(?i)` +
		// numeric
		`(?:^|[\s,!?(])` +
		`[6б].{0,20}7` +
		`(?:[\s,!?)]|$)` +
		`|` +
		// written
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
