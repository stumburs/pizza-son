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
			if rand.Float64() < 0.15 {
				response += " zaza"
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, response)
			return true
		},
	})
}
