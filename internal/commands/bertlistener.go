package commands

import (
	"fmt"
	"log"
	"math/rand/v2"
	"pizza-son/internal/bot"
	"pizza-son/internal/models"
	"pizza-son/internal/services"
	"strings"
	"time"
)

type TrollRule struct {
	User     string
	Chance   float64
	Response string
	Channel  string
}

var trollRules = []TrollRule{
	{
		User:     "itzkxtee",
		Chance:   0.2,
		Response: "camembert",
		Channel:  "sir_lysergium",
	},
	{
		User:     "itzkxtee",
		Chance:   0.2,
		Response: "cheesebert",
		Channel:  "sir_lysergium",
	},
}

var trollRulesEnabled = make(map[string]bool) // channel -> enabled/disabled

func isTrollEnabled(channel string) bool {
	enabled, exists := trollRulesEnabled[channel]
	if !exists {
		return true
	}
	return enabled
}

const zazaChance = 0.05

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "bertcheck",
		Description: "Responds with a random bert",
		Cooldown:    69 * time.Second,
		Matcher: func(ctx bot.CommandContext) bool {
			msg := strings.ToLower(ctx.Message.Text)
			return !strings.HasPrefix(msg, "!") && strings.Contains(msg, "bertcheck") && !strings.Contains(msg, "bertcheckcheck")
		},
		OnCooldown: func(ctx bot.CommandContext, remaining time.Duration) {
			services.TwitchServiceInstance.Timeout(ctx.Message.Channel, ctx.Message.User.ID, 69, "overberting smh")
		},
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

			baseResponse := pickBertResponse(ctx.Message.User.Name, available, ctx.Message.Channel)

			if baseResponse == "" {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "No berts on this channel :(")
				return false
			}

			finalMessage := baseResponse

			var isZazaRoll bool = false

			if rand.Float64() < zazaChance {
				finalMessage += " zaza"
				isZazaRoll = true
			}

			if rand.Float64() < zazaChance {
				finalMessage += " zazaL"
				isZazaRoll = true
			}

			services.BertServiceInstance.RegisterActivation(ctx.Message.Channel, ctx.Message.User.Name, baseResponse, isZazaRoll)

			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, finalMessage)

			// lyser
			if ctx.Message.Channel == "sir_lysergium" {
				// goldenzazabert things
				if strings.Contains(finalMessage, "goldenzazabert") {
					ctx.Client.Say(ctx.Message.Channel, fmt.Sprintf("oda %s got a %s oda !!! omg ur so cool BigDog oda", ctx.Message.User.DisplayName, finalMessage))
				}

				if strings.Contains(finalMessage, "Angine-de-Bertrine") {
					ctx.Client.Say(ctx.Message.Channel, "oda ADP oda")
				}
			}

			return true
		},
	})
	RegisterListener(bot.ListenerEntry{
		Name:        "bertcheckcheck",
		Description: "Responds with bertcheck",
		Cooldown:    69 * time.Second,
		Matcher: func(ctx bot.CommandContext) bool {
			msg := strings.ToLower(ctx.Message.Text)
			return !strings.HasPrefix(msg, "!") && strings.Contains(msg, "bertcheckcheck")
		},
		OnCooldown: func(ctx bot.CommandContext, remaining time.Duration) {
			services.TwitchServiceInstance.Timeout(ctx.Message.Channel, ctx.Message.User.ID, 69, "overberting^2 smh")
		},
		Handler: func(ctx bot.CommandContext) bool {
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "bertcheck")
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
				emoteCount, err := services.SevenTVServiceInstance.Fetch(ctx.Message.Channel)
				if err != nil {
					log.Printf("[!bertcheck add] Failed to update emote set for %s: %s", ctx.Message.Channel, err)
				} else {
					log.Printf("[!bertcheck add] Updated emote set for %s: %d emotes", ctx.Message.Channel, emoteCount)
				}
			case "remove":
				if services.BertServiceInstance.RemoveBert(ctx.Message.Channel, name) {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Removed %s from the bert log", name))
					emoteCount, err := services.SevenTVServiceInstance.Fetch(ctx.Message.Channel)
					if err != nil {
						log.Printf("[!bertcheck remove] Failed to update emote set for %s: %s", ctx.Message.Channel, err)
					} else {
						log.Printf("[!bertcheck remove] Updated emote set for %s: %d emotes", ctx.Message.Channel, emoteCount)
					}
				} else {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Could not find %s in the bert log", name))
				}
			}
		},
	})

	Register(bot.Command{
		Name:        "bertstats",
		Description: "Shows bertcheck stats for any user.",
		Usage:       "!bertstats",
		Permission:  bot.All,
		Category:    bot.CategoryFun,
		Examples: []bot.CommandExample{
			{Input: "!bertstats", Output: "View all bert stats at https://bertstats.stumburs.id.lv"},
		},
		Handler: func(ctx bot.CommandContext) {
			target := ctx.Message.User.Name
			args := strings.Fields(ctx.Message.Text)
			if len(args) > 1 {
				target = strings.ToLower(strings.TrimPrefix(args[1], target))
			}

			stats := services.BertServiceInstance.GetUserStats(ctx.Message.Channel, target)
			if stats.TotalBertchecks == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("%s has never bertchecked smh based berters here: https://bertstats.stumburs.id.lv", target))
				return
			}

			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf(
				"%s: %d total berts. Most common: %s (%dx). %d/%d collected. %d berts by all chatters. View all bert stats at https://bertstats.stumburs.id.lv",
				target,
				stats.TotalBertchecks,
				stats.MostCommonBert,
				stats.MostCommonCount,
				stats.BertsCollectedOutOfAll,
				stats.TotalBerts,
				stats.ChannelTotalBertchecks,
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
			matched := false

			// separate bertcheck test so we don't add 'bertcheck' to the list of berts
			if strings.Contains(msg, "bertcheck") {
				matched = true
			} else {
				availableBerts := services.BertServiceInstance.GetBerts(ctx.Message.Channel)
				for _, bert := range availableBerts {
					if strings.Contains(msg, strings.ToLower(bert)) {
						matched = true
						break
					}
				}
			}

			if matched {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "firsttimeberter")
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
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "AWPert")
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
					services.TwitchServiceInstance.Timeout(ctx.Message.Channel, ctx.Message.User.ID, 1, "not using emote smh")
				}
				ctx.Client.Say(ctx.Message.Channel, "smh !bertcheck isn't a command")
				return true
			}
			return false
		},
	})

	Register(bot.Command{
		Name:        "rigbert",
		Description: "Toggle bertcheck rigged responses on/off for this channel.",
		Usage:       "!rigbert",
		Permission:  bot.Moderator,
		Category:    bot.CategoryFun,
		Examples: []bot.CommandExample{
			{Input: "!rigbert", Output: "bertcheck rigging enabled/disabled"},
		},
		Handler: func(ctx bot.CommandContext) {
			current := isTrollEnabled(ctx.Message.Channel)
			trollRulesEnabled[ctx.Message.Channel] = !current
			if !current {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "bertcheck rigging enabled")
			} else {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "bertcheck rigging disabled")
			}
		},
	})
}

func pickBertResponse(user string, availableBerts []string, channel string) string {
	// we do a bit of trolling
	if isTrollEnabled(channel) {
		for _, rule := range trollRules {
			if user == rule.User && rand.Float64() < rule.Chance && rule.Channel == channel {
				return rule.Response
			}
		}
	}

	// goldenzazabert - 0.005% chance in sir_lysergium
	if channel == "sir_lysergium" && rand.Float64() < 0.00005 {
		return "goldenzazabert"
	}

	var filteredBerts []string
	for _, b := range availableBerts {
		if b != "goldenzazabert" {
			filteredBerts = append(filteredBerts, b)
		}
	}

	// Normal berts
	if len(filteredBerts) > 0 {
		return filteredBerts[rand.IntN(len(filteredBerts))]
	}

	return ""
}
