package commands

import (
	"pizza-son/internal/bot"
	"regexp"
)

var urlRegex = regexp.MustCompile(`(?i)(https?://(www\.)?[-a-zA-Z0-9@:%._+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_+.~#?&/=]*)|www\.[-a-zA-Z0-9@:%._+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_+.~#?&/=]*)|[-a-zA-Z0-9@:%._+~#=]{1,256}\.(com|net|org|io|tv|gg|dev|app|me|co|uk|de|ru|fr|ca|au|jp|live|stream|chat|games|cloud|tech|ai)([-a-zA-Z0-9()@:%_+.~#?&/=]*))`)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "url-benner",
		Description: "Detects URLs in messages and ben's",
		Handler: func(ctx bot.CommandContext) bool {
			if !urlRegex.MatchString(ctx.Message.Message) {
				return false
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "ben")
			return true
		},
	})
}
