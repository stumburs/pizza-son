package commands

import (
	"pizza-son/internal/bot"
	"strings"
)

var friccChain = []struct {
	trigger  string
	response string
}{
	{"fricc u nine", "fricc u ten"},
	{"fricc u eight", "fricc u nine"},
	{"fricc u seven", "fricc u eight"},
	{"fricc u six", "fricc u seven"},
	{"fricc u five", "fricc u six"},
	{"fricc u four", "fricc u five"},
	{"fricc u three", "fricc u four"},
	{"fricc u two", "fricc u three"},
	{"fricc u too", "fricc u three"},
	{"fricc u", "fricc u too"},
	{"fricc", "fricc u too"},
}

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "fricc",
		Description: "fricc's someone back",
		Handler: func(ctx bot.CommandContext) bool {
			msg := strings.ToLower(ctx.Message.Text)
			if !strings.Contains(msg, "fricc") {
				return false
			}
			for _, entry := range friccChain {
				if strings.Contains(msg, entry.trigger) {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, entry.response)
					return true
				}
			}
			return true
		},
	})
}
