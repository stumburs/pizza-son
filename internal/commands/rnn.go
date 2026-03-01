package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	Register(bot.Command{
		Name:        "rnn",
		Description: "Generates text using a pre-trained RNN model based off over 97k messages across few channels the bot has lurked in over a few months.",
		Usage:       "!rnn [text-to-continue]",
		Handler: func(ctx bot.CommandContext) {
			seed := "i"
			if len(ctx.Args) > 0 {
				seed = strings.Join(ctx.Args, " ")
			}
			text := services.RNNServiceInstance.Generate(seed, 150, 0.8)
			if text == "" {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Failed to generate text.")
				return
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, text)
		},
	})
}
