package commands

import (
	"pizza-son/internal/bot"
	"strings"
)

func init() {
	Register(bot.Command{
		Name:        "repeat",
		Description: "Repeats what you tell it to. Pretty self explanatory.",
		Usage:       "!repeat <text>",
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
}
