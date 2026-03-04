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

func init() {
	Register(bot.Command{
		Name:        "command",
		Description: "Enable or disable commands in this channel.",
		Usage:       "!command <enable|disable|list> [command]",
		Permission:  bot.Moderator,
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
				disabled := services.ChannelSettingsInstance.ListDisabled(ctx.Message.Channel)
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
}
