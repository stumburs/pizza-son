package commands

import (
	"pizza-son/internal/bot"
)

const docsPageURL = "https://pizzason.stumburs.id.lv/docs"

func init() {
	Register(bot.Command{
		Name:        "commands",
		Description: "Links to the command and listener documentation.",
		Usage:       "!commands",
		Category:    bot.CategoryUtility,
		Permission:  bot.All,
		Examples: []bot.CommandExample{
			{Input: "!commands", Output: "Use and abuse me with these commands: " + docsPageURL},
		},
		Handler: func(ctx bot.CommandContext) {
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Use and abuse me with these commands: "+docsPageURL)
		},
	})
}
