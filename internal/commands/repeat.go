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
		Name:        "repeat",
		Description: "Repeats what you tell it to. Pretty self explanatory.",
		Usage:       "!repeat <text>",
		Category:    bot.CategoryUtility,
		Permission:  bot.BotModerator,
		Examples: []bot.CommandExample{
			{Input: "!repeat meowdy", Output: "meowdy"},
		},
		Handler: func(ctx bot.CommandContext) {
			if len(ctx.Args) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !repeat <text>")
				return
			}
			ctx.Client.Say(ctx.Message.Channel, strings.Join(ctx.Args, " "))
		},
	})
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

			seconds, err := strconv.Atoi(ctx.Args[0])
			if err != nil {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Interval must be a positive number.")
				return
			}

			message := strings.Join(ctx.Args[1:], " ")

			// Stop any existing repeat for this user
			activeRepeatsMu.Lock()
			if ch, ok := activeRepeats[key]; ok {
				close(ch)
			}
			stop := make(chan struct{})
			activeRepeats[key] = stop
			activeRepeatsMu.Unlock()

			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Repeating every %ds. Use !repeatevery stop to stop.", seconds))

			go func() {
				ticker := time.NewTicker(time.Duration(seconds) * time.Second)
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
