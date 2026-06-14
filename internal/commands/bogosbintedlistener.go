package commands

import (
	"pizza-son/internal/bot"
	"strings"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "bogosbinted",
		Description: "glorps when someone says bogos binted",
		Handler: func(ctx bot.CommandContext) bool {
			if !strings.Contains(strings.ToLower(ctx.Message.Text), "bogos binted") {
				return false
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "glorp")
			return true
		},
	})
}
