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
		Name:        "tmknowledge",
		Description: "Trackmania knowledge facts. Bot moderators can add and remove facts.",
		Usage:       "!tmknowledge | !tmknowledge <number> | !tmknowledge add <text> | !tmknowledge remove <number>",
		Category:    bot.CategoryAI,
		Examples: []bot.CommandExample{
			{Input: "!tmknowledge", Output: "#3: Uberlap is a technique where you respawn at the last checkpoint - by meower"},
			{Input: "!tmknowledge 2", Output: "#2: Ice slides are faster than drifting on ice - by meower"},
			{Input: "!tmknowledge add Uberlap is a technique...", Output: "Trackmania knowledge #4 added!"},
			{Input: "!tmknowledge remove 4", Output: "Removed Trackmania knowledge #4."},
		},
		Handler: func(ctx bot.CommandContext) {
			if len(ctx.Args) == 0 {
				entry, number, ok := services.TMKnowledgeServiceInstance.Random()
				if !ok {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "No Trackmania knowledge yet. Bot moderators can add with !tmknowledge add <text>")
					return
				}
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, formatTMKnowledge(entry, number))
				return
			}

			switch ctx.Args[0] {
			case "add":
				if !bot.HasPermission(ctx.Message, bot.BotModerator) {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Only bot moderators can add Trackmania knowledge.")
					return
				}
				if len(ctx.Args) < 2 {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !tmknowledge add <text>")
					return
				}
				text := strings.Join(ctx.Args[1:], " ")
				number := services.TMKnowledgeServiceInstance.Add(text, ctx.Message.User.DisplayName)
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Trackmania knowledge #%d added!", number))

			case "remove", "delete", "rm":
				if !bot.HasPermission(ctx.Message, bot.BotModerator) {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Only bot moderators can remove Trackmania knowledge.")
					return
				}
				if len(ctx.Args) < 2 {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !tmknowledge remove <number>")
					return
				}
				number, err := strconv.Atoi(ctx.Args[1])
				if err != nil {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !tmknowledge remove <number>")
					return
				}
				if services.TMKnowledgeServiceInstance.Remove(number) {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Removed Trackmania knowledge #%d.", number))
				} else {
					count := services.TMKnowledgeServiceInstance.Count()
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Trackmania knowledge #%d not found. There are %d entries total.", number, count))
				}

			default:
				number, err := strconv.Atoi(ctx.Args[0])
				if err != nil {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !tmknowledge | !tmknowledge <number> | !tmknowledge add <text> | !tmknowledge remove <number>")
					return
				}
				entry, ok := services.TMKnowledgeServiceInstance.Get(number)
				if !ok {
					count := services.TMKnowledgeServiceInstance.Count()
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Trackmania knowledge #%d not found. There are %d entries total.", number, count))
					return
				}
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, formatTMKnowledge(entry, number))
			}
		},
	})
}

func formatTMKnowledge(e services.TMKnowledgeEntry, number int) string {
	return fmt.Sprintf("#%d: %s - by %s", number, e.Content, e.AddedBy)
}
