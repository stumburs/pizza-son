package commands

import (
	"math/rand/v2"
	"pizza-son/internal/bot"
	"pizza-son/internal/models"
	"pizza-son/internal/services"
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
	"BertlinWall",
	"bertha",
	"hamburgert",
	"nert",
	"bertrayal",
	"adBERTisement",
	"bertday",
	"yogBert",
	"BertRoss",
	"beert",
	"treb",
	"kebabert",
}

type TrollRule struct {
	User     string
	Chance   float64
	Response string
}

var trollRules = []TrollRule{
	{
		User:     "itzkxtee",
		Chance:   0.4,
		Response: "camembert",
	},
}

const zazaChance = 0.05

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "bertcheck",
		Description: "Responds with a random bert",
		Handler: func(ctx bot.CommandContext) bool {
			msg := strings.ToLower(ctx.Message.Text)

			if !strings.Contains(msg, "bertcheck") {
				return false
			}
			if strings.HasPrefix(msg, "!") {
				return false
			}

			response := pickBertResponse(ctx.Message.User.Name)
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
			msg := strings.ToLower(ctx.Message.Text)
			for _, bert := range berts {
				if strings.Contains(msg, strings.ToLower(bert)) {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "firsttimeberter")
					return true
				}
			}
			return false
		},
	})
	RegisterListener(bot.ListenerEntry{
		Name:        "snipebert",
		Description: "Detects snipe messages",
		Handler: func(ctx bot.CommandContext) bool {
			msg := strings.ToLower(ctx.Message.Text)
			if !strings.Contains(msg, "snip") {
				return false
			}
			snipeWords := []string{"sniped", "snipe", "sniper", "get sniped", "got sniped"}
			for _, word := range snipeWords {
				if strings.Contains(msg, word) {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "snipebert")
					return true
				}
			}
			return false
		},
	})
	RegisterListener(bot.ListenerEntry{
		Name:        "!bertcheck corrector",
		Description: "Corrects people who use !bertcheck instead of bertcheck",
		Handler: func(ctx bot.CommandContext) bool {
			msg := strings.ToLower(ctx.Message.Text)
			if msg == "!bertcheck" {
				// Only timeout on Twitch
				if ctx.Message.Platform == models.PlatformTwitch {
					services.TwitchServiceInstance.Timeout(ctx.Message.Channel, ctx.Message.User.ID, 21, "not using emote smh")
				}
				ctx.Client.Say(ctx.Message.Channel, "smh !bertcheck isn't a command")
				return true
			}
			return false
		},
	})
}

func pickBertResponse(user string) string {
	// we do a bit of trolling
	for _, rule := range trollRules {
		if user == rule.User && rand.Float64() < rule.Chance {
			resp := rule.Response
			if rand.Float64() < zazaChance {
				resp += " zaza"
			}
			return resp
		}
	}

	// Normal berts
	resp := berts[rand.IntN(len(berts))]
	if rand.Float64() < zazaChance {
		resp += " zaza"
	}
	return resp
}
