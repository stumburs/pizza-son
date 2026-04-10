package commands

import (
	"fmt"
	"pizza-son/internal/bot"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	activeRepeats   = make(map[string]chan struct{}) // key: channel:user
	activeRepeatsMu sync.Mutex
)

func init() {

	Register(bot.Command{
		Name:        "repeatevery",
		Description: "Repeatedly send a message every N seconds. Use !repeatevery stop to stop.",
		Usage:       "!repeatevery <seconds> <message> | !repeatevery stop",
		Category:    bot.CategoryUtility,
		Permission:  bot.Moderator,
		Examples: []bot.CommandExample{
			{Input: "!repeatevery 10 meow", Output: "Repeating every 10s. Use !repeatevery stop to stop."},
			{Input: "!repeatevery stop", Output: "Stopped repeating."},
		},
		Handler: func(ctx bot.CommandContext) {
			if len(ctx.Args) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !repeatevery <seconds> <message> | !repeatevery stop")
				return
			}

			key := ctx.Message.Channel + ":" + ctx.Message.User.ID

			if strings.ToLower(ctx.Args[0]) == "stop" {
				activeRepeatsMu.Lock()
				if ch, ok := activeRepeats[key]; ok {
					close(ch)
					delete(activeRepeats, key)
					activeRepeatsMu.Unlock()
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Stopped repeating.")
				} else {
					activeRepeatsMu.Unlock()
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "No active repeat to stop.")
				}
				return
			}

			if len(ctx.Args) < 2 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !repeatevery <seconds> <message> | !repeatevery stop")
				return
			}

			seconds, err := strconv.ParseFloat(ctx.Args[0], 64)
			if err != nil || seconds < 0.1 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Interval must be at least 0.1 seconds.")
				return
			}

			interval := time.Duration(seconds * float64(time.Second))

			message := strings.Join(ctx.Args[1:], " ")

			// Stop any existing repeat for this user
			activeRepeatsMu.Lock()
			if ch, ok := activeRepeats[key]; ok {
				close(ch)
			}
			stop := make(chan struct{})
			activeRepeats[key] = stop
			activeRepeatsMu.Unlock()

			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Repeating every %gs. Use !repeatevery stop to stop.", seconds))

			go func() {
				ticker := time.NewTicker(interval)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						ctx.Client.Say(ctx.Message.Channel, message)
					case <-stop:
						return
					}
				}
			}()
		},
	})
}
