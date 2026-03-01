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
		Name:        "quote",
		Description: "Manage and retrieve quotes. Only VIP's can add new quotes.",
		Usage:       "!quote | !quote <number> | !quote add <text>",
		Handler: func(ctx bot.CommandContext) {
			if len(ctx.Args) == 0 {
				// Random quote
				quote, number, ok := services.QuoteServiceInstance.Random(ctx.Message.Channel)
				if !ok {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "No quotes yet! Add one with !quote add <text>")
					return
				}
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, formatQuote(quote, number))
				return
			}

			switch ctx.Args[0] {
			case "add":
				// Only VIP's can add new quotes
				if !bot.HasPermission(ctx.Message, bot.VIP) {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Only VIP's can add quotes.")
					return
				}
				if len(ctx.Args) < 2 {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !quote add <text>")
					return
				}
				text := strings.Join(ctx.Args[1:], " ")
				number := services.QuoteServiceInstance.Add(ctx.Message.Channel, text, ctx.Message.User.DisplayName)
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Quote #%d added!", number))
			default:
				// Specific quote by number
				number, err := strconv.Atoi(ctx.Args[0])
				if err != nil {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !quote | !quote <number> | !quote add <text>")
					return
				}
				quote, ok := services.QuoteServiceInstance.Get(ctx.Message.Channel, number)
				if !ok {
					count := services.QuoteServiceInstance.Count(ctx.Message.Channel)
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Quote #%d not found. There are %d quotes total.", number, count))
					return
				}
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, formatQuote(quote, number))
			}
		},
	})
}

func formatQuote(q services.Quote, number int) string {
	return fmt.Sprintf("#%d: \" %s \" - added by %s on %s", number, q.Text, q.AddedBy, q.CreatedAt)
}
