package commands

import (
	"fmt"
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "counter-trigger",
		Description: "Increments a counter when its command is used",
		Handler: func(ctx bot.CommandContext) bool {
			msg := strings.TrimSpace(ctx.Message.Text)
			if !strings.HasPrefix(msg, "!") {
				return false
			}
			name := strings.ToLower(strings.Fields(strings.TrimPrefix(msg, "!"))[0])
			if !services.CounterServiceInstance.Exists(ctx.Message.Channel, name) {
				return false
			}
			value, err := services.CounterServiceInstance.Increment(ctx.Message.Channel, name)
			if err != nil {
				return false
			}
			ctx.Client.Say(ctx.Message.Channel, fmt.Sprintf("%s: %d", name, value))
			return true
		},
	})
}
