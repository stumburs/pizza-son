package commands

import (
	"fmt"
	"math/rand/v2"
	"pizza-son/internal/bot"
	"pizza-son/internal/models"
	"pizza-son/internal/services"
	"strings"
)

type TrollRule struct {
	User     string
	Chance   float64
	Response string
}

var trollRules = []TrollRule{
	{
		User:     "itzkxtee",
		Chance:   0.2,
		Response: "camembert",
	},
	{
		User:     "itzkxtee",
		Chance:   0.2,
		Response: "cheesebert",
	},
}

const zazaChance = 0.05

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "bertcheck",
		Description: "Responds with a random bert",
		Handler: func(ctx bot.CommandContext) bool {
			msg := strings.ToLower(ctx.Message.Text)

			// Ignore if command like !bertstats or !bertcheck add
			if strings.HasPrefix(msg, "!") || !strings.Contains(msg, "bertcheck") {
				return false
			}

			// bertcheckcheck funnies
			if strings.Contains(msg, "bertcheckcheck") {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "bertcheck")
				return true
			}

			available := services.BertServiceInstance.GetBerts(ctx.Message.Channel)

			baseResponse := pickBertResponse(ctx.Message.User.Name, available)

			if baseResponse == "" {
				return false
			}

			services.BertServiceInstance.RegisterActivation(ctx.Message.Channel, ctx.Message.User.Name, baseResponse)

			finalMessage := baseResponse
			if rand.Float64() < zazaChance {
				finalMessage += " zaza"
			}

			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, finalMessage)
			return true
		},
	})
	// Add/remove berts
	Register(bot.Command{
		Name:        "bertcheck",
		Description: "Add or remove berts used by 'bertcheck' on this channel.",
		Usage:       "!bertcheck <add|remove> <bert>",
		Permission:  bot.Moderator,
		Category:    bot.CategoryFun,
		Examples: []bot.CommandExample{
			{Input: "!bertcheck add bertnard", Output: "Added bertnard to the bert log"},
			{Input: "!bertcheck remove beert", Output: "Removed bertnard from the bert log"},
		},
		Handler: func(ctx bot.CommandContext) {
			args := strings.Fields(ctx.Message.Text)
			if len(args) < 3 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !bertcheck <add|remove> <bert>")
				return
			}

			action, name := strings.ToLower(args[1]), args[2]

			switch action {
			case "add":
				services.BertServiceInstance.AddBert(ctx.Message.Channel, name)
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Added %s to the bert log", name))
			case "remove":
				if services.BertServiceInstance.RemoveBert(ctx.Message.Channel, name) {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Removed %s from the bert log", name))
				} else {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Could not find %s in the bert log", name))
				}
			}
		},
	})

	Register(bot.Command{
		Name:        "bertstats",
		Description: "Shows bertcheck stats for any user.",
		Usage:       "!bertstats [user]",
		Permission:  bot.All,
		Category:    bot.CategoryFun,
		Examples: []bot.CommandExample{
			{Input: "!bertstats", Output: "casual_berter: 12 total berts. Most common: bertday (2x)"},
			{Input: "!bertstats @bertman", Output: "bertman: 69420 total berts. Most common: bert (5325x)"},
		},
		Handler: func(ctx bot.CommandContext) {
			target := ctx.Message.User.Name
			args := strings.Fields(ctx.Message.Text)
			if len(args) > 1 {
				target = strings.TrimPrefix(args[1], "@")
			}

			total, common, count := services.BertServiceInstance.GetUserStats(ctx.Message.Channel, target)
			if total == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("%s has never bertchecked smh", target))
				return
			}

			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf(
				"%s: %d total berts. Most common: %s (%dx)",
				target, total, common, count,
			))
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

			// separate bertcheck test so we don't add 'bertcheck' to the list of berts
			if strings.Contains(msg, "bertcheck") {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "firsttimeberter")
				return true
			}

			availableBerts := services.BertServiceInstance.GetBerts(ctx.Message.Channel)

			for _, bert := range availableBerts {
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
		Name:        "!bertcheckcorrector",
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

func pickBertResponse(user string, availableBerts []string) string {
	// we do a bit of trolling
	for _, rule := range trollRules {
		if user == rule.User && rand.Float64() < rule.Chance {
			return rule.Response
		}
	}

	// Normal berts
	if len(availableBerts) > 0 {
		return availableBerts[rand.IntN(len(availableBerts))]
	}

	return "No berts on this channel :("
}
