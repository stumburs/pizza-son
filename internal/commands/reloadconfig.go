package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/config"
)

func init() {
	Register(bot.Command{
		Name:        "reloadconfig",
		Description: "Reloads the config file.",
		Usage:       "!reloadconfig",
		Permission:  bot.BotModerator,
		Examples: []bot.CommandExample{
			{Input: "!reloadconfig", Output: "Config reloaded! meow"},
		},
		Handler: func(ctx bot.CommandContext) {
			config.Reload("config.toml")
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Config reloaded! meow")
		},
	})
}
