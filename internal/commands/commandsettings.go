package commands

import (
	"fmt"
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
)

// These commands can never be disabled
var protectedCommands = map[string]bool{
	"command": true,
	"ping":    true,
}

// These listeners can never be disabled
var protectedListeners = map[string]bool{
	"log": true,
}

func init() {
	Register(bot.Command{
		Name:        "command",
		Description: "Enable or disable commands in this channel.",
		Usage:       "!command <enable|disable|list> [command]",
		Category:    bot.CategoryModeration,
		Examples: []bot.CommandExample{
			{Input: "!command enable hug", Output: "Command !hug enabled."},
			{Input: "!command disable mark", Output: "Command !mark disabled."},
			{Input: "!command list", Output: "Disabled commands: !mark, !nsfw, !quote"},
		},
		Permission: bot.Moderator,
		Handler: func(ctx bot.CommandContext) {
			if len(ctx.Args) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !command <enable|disable|list> [command]")
				return
			}

			switch ctx.Args[0] {
			case "enable":
				if len(ctx.Args) < 2 {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !command enable <command>")
					return
				}
				cmd := strings.ToLower(ctx.Args[1])
				services.ChannelSettingsInstance.EnableCommand(ctx.Message.Channel, cmd)
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Command !%s enabled.", cmd))

			case "disable":
				if len(ctx.Args) < 2 {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !command disable <command>")
					return
				}
				cmd := strings.ToLower(ctx.Args[1])
				if protectedCommands[cmd] {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Command !%s cannot be disabled.", cmd))
					return
				}
				services.ChannelSettingsInstance.DisableCommand(ctx.Message.Channel, cmd)
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Command !%s disabled.", cmd))

			case "list":
				disabled := services.ChannelSettingsInstance.ListDisabledCommands(ctx.Message.Channel)
				if len(disabled) == 0 {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "No commands are disabled in this channel.")
					return
				}
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Disabled commands: !"+strings.Join(disabled, ", !"))

			default:
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !command <enable|disable|list> [command]")
			}
		},
	})

	// Listener toggle
	Register(bot.Command{
		Name:        "listener",
		Description: "Enable or disable listeners in this channel.",
		Usage:       "!listener <enable|disable|list> [name]",
		Category:    bot.CategoryModeration,
		Examples: []bot.CommandExample{
			{Input: "!listener enable meow", Output: "Listener enabled: meow"},
			{Input: "!listener disable mark", Output: "Command disabled: mark."},
			{Input: "!listener list", Output: "Listeners: mark (disabled), meow"},
		},
		Permission: bot.Moderator,
		Handler: func(ctx bot.CommandContext) {
			if len(ctx.Args) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !listener <enable|disable|list> [name]")
				return
			}

			switch strings.ToLower(ctx.Args[0]) {
			case "enable":
				if len(ctx.Args) < 2 {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !listener enable <name>")
					return
				}
				services.ChannelSettingsInstance.EnableListener(ctx.Message.Channel, ctx.Args[1])
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Listener enabled: "+ctx.Args[1])

			case "disable":
				if len(ctx.Args) < 2 {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !command disable <command>")
					return
				}
				cmd := strings.ToLower(ctx.Args[1])
				if protectedListeners[cmd] {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Listener !%s cannot be disabled.", cmd))
					return
				}
				services.ChannelSettingsInstance.DisableListener(ctx.Message.Channel, ctx.Args[1])
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Listener disabled: "+ctx.Args[1])

			case "list":
				disabled := services.ChannelSettingsInstance.ListDisabledListeners(ctx.Message.Channel)
				all := ctx.Registry.Listeners()

				names := make([]string, 0, len(all))
				disabledSet := make(map[string]bool, len(disabled))
				for _, d := range disabled {
					disabledSet[d] = true
				}
				for _, l := range all {
					if disabledSet[l.Name] {
						names = append(names, l.Name+" (disabled)")
					} else {
						names = append(names, l.Name)
					}
				}
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Listeners: "+strings.Join(names, ", "))

			default:
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !command <enable|disable|list> [command]")
			}
		},
	})
}
