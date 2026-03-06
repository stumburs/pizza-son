package commands

import (
	"fmt"
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	Register(bot.Command{
		Name:        "slicerboard",
		Description: "Shows the top 5 slice holders.",
		Usage:       "!slicerboard",
		Category:    bot.CategoryCurrency,
		Examples: []bot.CommandExample{
			{Input: "!slicerboard", Output: "Top slicers: #1 pizza_tm (123574 🍕) | #2 your_mom (99 🍕) | #3 pineapplesonpizza (0 🍕)"},
		},
		Handler: func(ctx bot.CommandContext) {
			top := services.CurrencyServiceInstance.TopBalances(5)
			if len(top) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "No slice holders yet!")
				return
			}

			parts := make([]string, len(top))
			for i, entry := range top {
				username, err := services.TwitchServiceInstance.GetUsername(entry.UserID)
				if err != nil {
					username = entry.UserID
				}
				parts[i] = fmt.Sprintf("#%d %s (%d 🍕)", i+1, username, entry.Balance)
			}

			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Top slicers: "+strings.Join(parts, " | "))
		},
	})
}
