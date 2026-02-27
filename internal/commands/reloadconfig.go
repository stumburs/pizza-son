package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/config"
)

func init() {
	Register(bot.Command{
		Name:        "reloadconfig",
		Description: "Reloads the config file.",
		Permission:  bot.BotModerator,
		Handler: func(ctx bot.CommandContext) {
			config.Reload("config.toml")
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Config reloaded! meow")
		},
	})
}
