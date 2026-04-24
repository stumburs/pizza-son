package commands

import (
	"fmt"
	"math/rand/v2"
	"pizza-son/internal/bot"
	"pizza-son/internal/models"
	"pizza-son/internal/services"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	activeEmoteSpams   = make(map[string]chan struct{})
	activeEmoteSpamsMu sync.Mutex
)

func init() {
	Register(bot.Command{
		Name:        "emotespam",
		Description: "Repeatedly send a random 7TV emote from this channel. Use !emotespam stop to stop",
		Usage:       "!emotespam <seconds> | !emotespam stop",
		Category:    bot.CategoryFun,
		Permission:  bot.Moderator,
		Examples: []bot.CommandExample{
			{Input: "!emotespam 2", Output: "Spamming emotes every 2s. Use !emotespam stop to stop."},
			{Input: "!emotespam 0.5", Output: "Spamming emotes every 0.5s. Use !emotespam stop to stop."},
			{Input: "!emotespam stop", Output: "Stopped emote spam."},
		},
		Handler: func(ctx bot.CommandContext) {
			// Exclude Discord
			if ctx.Message.Platform == models.PlatformDiscord {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "This is a Twitch exclusive command.")
				return
			}

			if len(ctx.Args) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !emotespam <seconds> | !emotespam stop")
				return
			}

			key := ctx.Message.Channel + ":" + ctx.Message.User.ID

			if strings.ToLower(ctx.Args[0]) == "stop" {
				activeEmoteSpamsMu.Lock()
				if ch, ok := activeEmoteSpams[key]; ok {
					close(ch)
					delete(activeEmoteSpams, key)
					activeEmoteSpamsMu.Unlock()
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Stopped emote spam.")
				} else {
					activeEmoteSpamsMu.Unlock()
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "No active emote spam to stop.")
				}
				return
			}

			emotes := services.SevenTVServiceInstance.GetEmotes(ctx.Message.Channel)
			if len(emotes) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "No 7TV emotes cached for this channel. Run !fetch7tv first.")
				return
			}

			seconds, err := strconv.ParseFloat(ctx.Args[0], 64)
			if err != nil || seconds < 0.1 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Interval must be at least 0.1 seconds.")
				return
			}

			interval := time.Duration(seconds * float64(time.Second))

			// Stop existing spam for this user
			activeEmoteSpamsMu.Lock()
			if ch, ok := activeEmoteSpams[key]; ok {
				close(ch)
			}
			stop := make(chan struct{})
			activeEmoteSpams[key] = stop
			activeEmoteSpamsMu.Unlock()

			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Spamming emotes every %gs. Use !emotespam stop to stop.", seconds))

			go func() {
				ticker := time.NewTicker(interval)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						emote := emotes[rand.IntN(len(emotes))]
						ctx.Client.Say(ctx.Message.Channel, emote.Name)
					case <-stop:
						return
					}
				}
			}()
		},
	})
}
