package commands

import (
	"fmt"
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strconv"
	"strings"
)

func init() {
	Register(bot.Command{
		Name:        "slices",
		Description: "Check your or another user's pizza slice balance, or give slices to another user. Only Moderators can set amounts.",
		Usage:       "!slices | !slices <user> | !slices give <user> <amount> | !slices set <user> <amount>",
		Examples: []bot.CommandExample{
			{Input: "!slices", Output: "You have 69 pizza slices."},
			{Input: "!slices @big_bob", Output: "big_bob has 420 pizza slices."},
			{Input: "!slices give @cat_enjoyer123 500", Output: "Gave 500 pizza slices to cat_enjoyer123."},
			{Input: "!slices set @naughty_person 0", Output: "Set naughty_person slices to 0."},
		},
		Handler: func(ctx bot.CommandContext) {
			if len(ctx.Args) == 0 {
				balance := services.CurrencyServiceInstance.Balance(ctx.Message.User.ID)
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("You have %d pizza slices.", balance))
				return
			}

			switch ctx.Args[0] {
			case "give":
				if len(ctx.Args) < 3 {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !slices give <user> <amount>")
					return
				}
				target := strings.ToLower(strings.TrimPrefix(ctx.Args[1], "@"))

				// Prevent self-give
				if target == strings.ToLower(ctx.Message.User.Name) {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "You can't give slices to yourself.")
					return
				}

				amount, err := strconv.Atoi(ctx.Args[2])
				if err != nil || amount <= 0 {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Amount must be a positive number.")
					return
				}
				userID, err := services.TwitchServiceInstance.GetUserID(target)
				if err != nil {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Could not find user: %s", target))
					return
				}
				remaining, ok := services.CurrencyServiceInstance.Give(ctx.Message.User.ID, userID, amount)
				if !ok {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("You don't have enough slices. You only have %d slices.", remaining))
					return
				}
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Gave %d pizza slices to %s.", amount, target))
			case "set":
				if !bot.HasPermission(ctx.Message, bot.Moderator) {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Only moderators can set slices.")
					return
				}
				if len(ctx.Args) < 3 {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !slices set <user> <amount>")
					return
				}
				target := strings.ToLower(strings.TrimPrefix(ctx.Args[1], "@"))
				amount, err := strconv.Atoi(ctx.Args[2])
				if err != nil || amount < 0 {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Amount must be a positive number or 0.")
					return
				}
				userID, err := services.TwitchServiceInstance.GetUserID(target)
				if err != nil {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Could not find user: %s", target))
					return
				}
				services.CurrencyServiceInstance.Set(userID, amount)
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Set %s slices to %d.", target, amount))
			// Check another user's balance
			default:
				target := strings.ToLower(strings.TrimPrefix(ctx.Args[0], "@"))
				userID, err := services.TwitchServiceInstance.GetUserID(target)
				if err != nil {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Could not find user: %s", target))
					return
				}
				balance := services.CurrencyServiceInstance.Balance(userID)
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("%s has %d pizza slices.", target, balance))
			}
		},
	})
}
