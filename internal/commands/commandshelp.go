package commands

import (
	"pizza-son/internal/bot"
)

func init() {
	Register(bot.Command{
		Name:        "commands",
		Description: "Returns site to all commands.",
		Permission:  bot.All,
		Handler: func(ctx bot.CommandContext) {
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Use and abuse me with these commands: https://stumburs.github.io/pizza-son/commands")
		},
	})
}
