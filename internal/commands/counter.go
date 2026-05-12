package commands

import (
	"fmt"
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strconv"
	"strings"
)

func init() {
	Register(bot.Command{
		Name:        "counter",
		Description: "Manage counters for this channel. If you want to include the count in your custom message use '{}', for example. 'This counter has been activated {} times'.",
		Usage:       "!counter add <name> [message] | !counter remove <name> | !counter set <name> <value> | !counter list",
		Category:    bot.CategoryModeration,
		Permission:  bot.Moderator,
		Examples: []bot.CommandExample{
			{Input: "!counter add deaths i've died {} times", Output: "Counter 'deaths' created. Use !deaths to increment it."},
			{Input: "!counter remove deaths", Output: "Counter 'deaths' removed."},
			{Input: "!counter set deaths 5", Output: "Counter 'deaths' set to 5."},
			{Input: "!counter list", Output: "Counters: deaths (3), wins (7)"},
			{Input: "!counter", Output: "i've died 4 times"},
		},
		Handler: func(ctx bot.CommandContext) {
			if len(ctx.Args) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !counter add <name> [message] | !counter remove <name> | !counter set <name> <value> | !counter list")
				return
			}
			switch strings.ToLower(ctx.Args[0]) {
			case "add":
				if len(ctx.Args) < 2 {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !counter add <name> [message]")
					return
				}
				name := strings.ToLower(ctx.Args[1])
				message := ""
				if len(ctx.Args) > 2 {
					message = strings.Join(ctx.Args[2:], " ")
				}
				if err := services.CounterServiceInstance.Add(ctx.Message.Channel, name, message); err != nil {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, err.Error())
					return
				}
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Counter '%s' created. Use !%s to increment it.", name, name))

			case "remove":
				if len(ctx.Args) < 2 {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !counter remove <name>")
					return
				}
				name := strings.ToLower(ctx.Args[1])
				if err := services.CounterServiceInstance.Remove(ctx.Message.Channel, name); err != nil {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, err.Error())
					return
				}
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Counter '%s' removed.", name))

			case "set":
				if len(ctx.Args) < 3 {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !counter set <name> <value>")
					return
				}
				name := strings.ToLower(ctx.Args[1])
				value, err := strconv.Atoi(ctx.Args[2])
				if err != nil {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Value must be a number.")
					return
				}
				if err := services.CounterServiceInstance.Set(ctx.Message.Channel, name, value); err != nil {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, err.Error())
					return
				}
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Counter '%s' set to %d.", name, value))

			case "list":
				counters := services.CounterServiceInstance.List(ctx.Message.Channel)
				if len(counters) == 0 {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "No counters yet. Add one with !counter add <name>")
					return
				}
				parts := make([]string, len(counters))
				for i, c := range counters {
					parts[i] = fmt.Sprintf("%s (%d)", c.Name, c.Value)
				}
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Counters: "+strings.Join(parts, ", "))

			default:
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !counter add | remove | set | list")
			}
		},
	})
}
