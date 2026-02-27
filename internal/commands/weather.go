package commands

import (
	"fmt"
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	Register(bot.Command{
		Name:        "weather",
		Description: "Gets the temperature for a location. Usage: !weather <location>",
		Handler: func(ctx bot.CommandContext) {
			if len(ctx.Args) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !weather <location>")
				return
			}
			location := strings.Join(ctx.Args, " ")
			temp, err := services.GetTemperature(location)
			if err != nil {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Failed to get weather.")
				return
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("%s: %s", location, temp))
		},
	})
}
