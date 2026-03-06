package commands

import (
	"fmt"
	"pizza-son/internal/bot"
	"strconv"
	"strings"
	"time"
)

func init() {
	Register(bot.Command{
		Name:        "timer",
		Description: "Start a timer that sends an optional message after it expires. Duration can be formatted using {x}s, {x}m|min, {x}/h.",
		Usage:       "!timer <duration> [message]",
		Examples: []bot.CommandExample{
			{Input: "!timer 5m pizza is ready", Output: "Timer set for 5m. | (after 5 minutes) | pizza_tm dinkDonk pizza is ready dinkDonk"},
			{Input: "!timer 2.5h", Output: "Timer set for 2.5h. | (after 2.5 hours) | dinkDonk Your end is nigh (timer is up) dinkDonk"},
		},
		Handler: func(ctx bot.CommandContext) {
			if len(ctx.Args) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !timer <duration> [message]")
				return
			}

			duration, err := parseDuration(ctx.Args[0])
			if err != nil {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Invalid duration: %q. Examples 5s, 10m, 2.5h, 1min", ctx.Args[0]))
				return
			}

			if duration <= 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Duration must be greater than 0.")
				return
			}

			if duration > 24*time.Hour {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Duration cannot exceed 24 hours.")
				return
			}

			message := "dinkDonk Your end is nigh (timer is up) dinkDonk"
			if len(ctx.Args) > 1 {
				message = strings.Join(ctx.Args[1:], " ")
			}
			displayName := ctx.Message.User.DisplayName
			channel := ctx.Message.Channel
			client := ctx.Client

			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Timer set for %s.", ctx.Args[0]))
			go func() {
				time.Sleep(duration)
				client.Say(channel, fmt.Sprintf("@%s dinkDonk %s dinkDonk", displayName, message))
			}()
		},
	})
}

func parseDuration(s string) (time.Duration, error) {
	s = strings.ToLower(strings.TrimSpace(s))

	units := []struct {
		suffix     string
		multiplier time.Duration
	}{
		{"ms", time.Millisecond},
		{"min", time.Minute},
		{"hrs", time.Hour},
		{"hr", time.Hour},
		{"h", time.Hour},
		{"m", time.Minute},
		{"s", time.Second},
	}

	for _, u := range units {
		if strings.HasSuffix(s, u.suffix) {
			numStr := strings.TrimSuffix(s, u.suffix)
			val, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid number: %q", numStr)
			}
			return time.Duration(val * float64(u.multiplier)), nil
		}
	}

	return 0, fmt.Errorf("unknown unit in %q", s)
}
