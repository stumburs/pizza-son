package commands

import (
	"fmt"
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
)

func init() {
	Register(bot.Command{
		Name:        "fetch7tv",
		Description: "Fetch and cache 7TV emotes for this channel.",
		Usage:       "!fetch7tv",
		Category:    bot.CategoryModeration,
		Permission:  bot.Moderator,
		Examples: []bot.CommandExample{
			{Input: "!fetch7tv", Output: "Fetched 69 emotes for nice_streamer87"},
		},
		Handler: func(ctx bot.CommandContext) {
			count, err := services.SevenTVServiceInstance.Fetch(ctx.Message.Channel)
			if err != nil {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Failed to fetch emotes: "+err.Error())
				return
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Fetched %d emotes for %s", count, ctx.Message.Channel))
		},
	})
}
