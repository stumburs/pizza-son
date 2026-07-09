package commands

import (
	"fmt"
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	Register(bot.Command{
		Name:        "streamlisten",
		Description: "Toggle the bot's ability to listen to stream audio and respond via LLM. Disabled by default.",
		Usage:       "!streamlisten <on|off>",
		Category:    bot.CategoryFun,
		Permission:  bot.Moderator,
		Examples: []bot.CommandExample{
			{Input: "!streamlisten on", Output: "Stream listening enabled in this channel."},
			{Input: "!streamlisten off", Output: "Stream listening disabled in this channel."},
		},
		Handler: func(ctx bot.CommandContext) {
			if len(ctx.Args) == 0 {
				enabled := services.STTServiceInstance.IsEnabled(ctx.Message.Channel)
				status := "disabled"
				if enabled {
					status = "enabled"
				}
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID,
					fmt.Sprintf("Stream listening is currently %s in this channel.", status))
				return
			}

			switch strings.ToLower(ctx.Args[0]) {
			case "on", "enable":
				services.STTServiceInstance.SetEnabled(ctx.Message.Channel, true)
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Stream listening enabled in this channel.")
			case "off", "disable":
				services.STTServiceInstance.SetEnabled(ctx.Message.Channel, false)
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Stream listening disabled in this channel.")
			case "test":
				if len(ctx.Args) < 2 {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !streamlisten test <text the streamer said>")
					return
				}
				text := strings.Join(ctx.Args[1:], " ")
				go services.STTServiceInstance.TestTranscript(ctx.Message.Channel, text)
			default:
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !streamlisten <on|off|test <text>>")
			}
		},
	})
}
