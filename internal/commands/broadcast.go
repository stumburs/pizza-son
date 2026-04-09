package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/config"
	"strings"
)

func init() {
	Register(bot.Command{
		Name:        "broadcast",
		Description: "Broadcasts a message across all connected channels.",
		Usage:       "!broadcast <text>",
		Category:    bot.CategoryModeration,
		Permission:  bot.BotModerator,
		Examples: []bot.CommandExample{
			{Input: "!broadcast Hello!", Output: "Hello! (In every connected channel.)"},
		},
		Handler: func(ctx bot.CommandContext) {
			if len(ctx.Args) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !broadcast <text>")
				return
			}

			channels := config.Get().Bot.Channels
			for _, channel := range channels {
				ctx.Client.Say(channel, strings.Join(ctx.Args, ""))
			}
		},
	})
}
