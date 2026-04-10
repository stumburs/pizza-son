package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	Register(bot.Command{
		Name:        "joke",
		Description: "Responds with a random joke.",
		Usage:       "!joke",
		Category:    bot.CategoryFun,
		Permission:  bot.All,
		Examples: []bot.CommandExample{
			{Input: "!joke", Output: "Insert funny joke here."},
		},
		Handler: func(ctx bot.CommandContext) {
			joke, err := services.GetJoke()
			if err != nil {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Failed to get joke: "+err.Error())
				return
			}
			clean := strings.Join(strings.Fields(joke.Joke), " ")
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, clean)
		},
	})
}
