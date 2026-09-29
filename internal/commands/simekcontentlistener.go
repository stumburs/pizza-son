package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "simekcontent",
		Description: "special timeout for simek",
		Handler: func(ctx bot.CommandContext) bool {
			if !strings.Contains(ctx.Message.Text, "ƈᴏɴᴛᴇɴᴛ") {
				return false
			}
			services.TwitchServiceInstance.Timeout(ctx.Message.Channel, ctx.Message.User.ID, 69, "ben Simek")
			ctx.Client.Say(ctx.Message.Channel, "smb Simek")
			return true
		},
	})
}
