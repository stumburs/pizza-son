package commands

import (
	"math/rand/v2"
	"pizza-son/internal/bot"
	"strings"
)

var berts = []string{
	"bert",
	"bertbert",
	"camembert",
	"cucumbert",
	"deadbertfather",
	"drugbustbert",
	"glitchbert",
	"hackerbert",
	"heisenbert",
	"LiBERTy",
	"londonbert",
	"Lubert",
	"monsterbert",
	"pipebombbert",
	"russiabert",
	"tylenolbert",
	"weinerbert",
	"zazabert",
	"fishbert",
	"Error404bert",
	"bertsittingverycomfortablearoundacampfirewithitsfriends",
	"sorbert",
	"berttosis",
	"Bertnard",
	"robbert",
	"bertsimpson",
	"Bort",
	"numbert",
	"strawberty",
	"cyBERTerrorist",
	"berthquake",
	"alBertEinstein",
	"snipebert",
	"bertJam",
}

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "bertcheck",
		Description: "Responds with a random bert",
		Handler: func(ctx bot.CommandContext) bool {
			if !strings.Contains(strings.ToLower(ctx.Message.Message), "bertcheck") {
				return false
			}
			response := berts[rand.IntN(len(berts))]
			if rand.Float64() < 0.05 {
				response += " zaza"
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, response)
			return true
		},
	})
	RegisterListener(bot.ListenerEntry{
		Name:        "firsttimebert",
		Description: "Detects first time bert'ers",
		Handler: func(ctx bot.CommandContext) bool {
			if !ctx.Message.FirstMessage {
				return false
			}
			msg := strings.ToLower(ctx.Message.Message)
			for _, bert := range berts {
				if strings.Contains(msg, strings.ToLower(bert)) {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "firsttimeberter")
					return true
				}
			}
			return false
		},
	})
}
