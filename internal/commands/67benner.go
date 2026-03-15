package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"regexp"
	"strings"
)

var sixSevenRegex = regexp.MustCompile(`(?i)(^|\s)67(\s|$)|sixty[-\s]?seven|six[-\s]?seven`)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "67-benner",
		Description: "Detects 67s in messages and ben's",
		Handler: func(ctx bot.CommandContext) bool {
			if !sixSevenRegex.MatchString(strings.TrimSpace(ctx.Message.Message)) {
				return false
			}
			ctx.Client.Say(ctx.Message.Channel, "ben")
			services.TwitchServiceInstance.Timeout(ctx.Message.Channel, ctx.Message.User.ID, 69, "ben")
			return true
		},
	})
}
