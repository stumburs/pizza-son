package commands

import (
	"math/rand/v2"
	"pizza-son/internal/bot"
	"strings"
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "0-0",
		Description: "responds to 0-0 with 00",
		Handler: func(ctx bot.CommandContext) bool {
			if ctx.Message.User.Name != "itzkxtee" {
				return false
			}
			if !strings.Contains(strings.ToLower(ctx.Message.Text), "0-0") {
				return false
			}
			if rand.Float64() < 0.25 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "00")
				return true
			}
			return false
		},
	})

	RegisterListener(bot.ListenerEntry{
		Name:        "00",
		Description: "responds to 00 with 0-0",
		Handler: func(ctx bot.CommandContext) bool {
			if ctx.Message.User.Name != "itzkxtee" {
				return false
			}
			if !strings.Contains(strings.ToLower(ctx.Message.Text), "00") {
				return false
			}
			if rand.Float64() < 0.25 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "0-0")
				return true
			}
			return false
		},
	})
}
