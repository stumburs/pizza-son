package commands

import (
	"fmt"
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	Register(bot.Command{
		Name:        "time",
		Description: "Gets the current time for a location.",
		Usage:       "!time <location>",
		Category:    bot.CategoryUtility,
		Examples: []bot.CommandExample{
			{Input: "!time London", Output: "London: 16:20"},
		},
		Handler: func(ctx bot.CommandContext) {
			if len(ctx.Args) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !time <location>")
				return
			}
			location := strings.Join(ctx.Args, " ")
			time, _, err := services.GetTimeForCity(location)
			if err != nil {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Failed to get time: "+err.Error())
				return
			}
			formattedTime, err := services.DatetimeToHourMinute(time.Datetime)
			if err != nil {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, err.Error())
				return
			}
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("%s: %s", location, formattedTime))
		},
	})
}
