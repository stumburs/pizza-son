package commands

import (
	"fmt"
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
)

func init() {
	Register(bot.Command{
		Name:        "toggle",
		Description: "Enable or disable all commands and listeners in this channel, remembering the previous state.",
		Usage:       "!toggle <on|off>",
		Category:    bot.CategoryModeration,
		Permission:  bot.Moderator,
		Examples: []bot.CommandExample{
			{Input: "!toggle off", Output: "Disabled 85 commands and 21 listeners in this channel. Use !toggle on to restore your previous settings."},
			{Input: "!toggle on", Output: "Restored commands and listeners to how they were before !toggle off."},
		},
		Handler: func(ctx bot.CommandContext) {
			usage := "Usage: !toggle <on|off>"
			if len(ctx.Args) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, usage)
				return
			}

			channel := ctx.Message.Channel
			switch strings.ToLower(ctx.Args[0]) {
			case "off":
				cmdNames := make([]string, 0, len(ctx.Registry.Commands()))
				for name := range ctx.Registry.Commands() {
					if ProtectedCommands[strings.ToLower(name)] {
						continue
					}
					cmdNames = append(cmdNames, name)
				}

				listenerNames := make([]string, 0, len(ctx.Registry.Listeners()))
				for _, l := range ctx.Registry.Listeners() {
					if ProtectedListeners[strings.ToLower(l.Name)] {
						continue
					}
					listenerNames = append(listenerNames, l.Name)
				}

				services.ChannelSettingsInstance.DisableAll(channel, cmdNames, listenerNames)
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID,
					fmt.Sprintf("Disabled %d commands and %d listeners in this channel. Use !toggle on to restore your previous settings.",
						len(cmdNames), len(listenerNames)))

			case "on":
				if services.ChannelSettingsInstance.EnableAll(channel) {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Restored commands and listeners to how they were before !toggle off.")
				} else {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "All commands and listeners enabled in this channel.")
				}

			default:
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, usage)
			}
		},
	})
}
