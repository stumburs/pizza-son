package commands

import (
	"math/rand/v2"
	"pizza-son/internal/bot"
	"strings"
)

const odaListeningChance = 0.05

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "odacheck",
		Description: "Responds with oda",
		Handler: func(ctx bot.CommandContext) bool {
			msg := strings.ToLower(ctx.Message.Text)

			if strings.HasPrefix(msg, "!") || !strings.Contains(msg, "odacheck") {
				return false
			}

			emote := "oda"
			if rand.Float64() < odaListeningChance {
				emote = "odaListening"
			}

			finalMessage := emote
			if rand.Float64() < zazaChance {
				finalMessage += " zaza"
			}

			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, finalMessage)
			return true
		},
	})
}
