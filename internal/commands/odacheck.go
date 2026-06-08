package commands

import (
	"math/rand/v2"
	"pizza-son/internal/bot"
	"strings"
)

const specialOdaChance = 0.05

var odaVariants = []string{
	"odaListening",
	"odaer",
}

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
			if rand.Float64() < specialOdaChance {
				emote = odaVariants[rand.IntN(len(odaVariants))]
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
