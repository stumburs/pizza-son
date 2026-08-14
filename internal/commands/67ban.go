package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
)

func init() {
	Register(bot.Command{
		Name:        "67",
		Description: "Perma-bans users when they type this command (in grubbyyyy channel)",
		Usage:       "!67",
		Category:    bot.CategoryFun,
		Examples: []bot.CommandExample{
			{Input: "!67", Output: "ben"},
		},
		Handler: func(ctx bot.CommandContext) {
			if ctx.Message.Channel == "grubbyyyy" {
				services.TwitchServiceInstance.Ban(ctx.Message.Channel, ctx.Message.User.ID, "ben")
				ctx.Client.Say(ctx.Message.Channel, "ben")
			}
		},
	})
}
