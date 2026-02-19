package commands

import (
	"pizza-son/internal/services"

	"github.com/gempir/go-twitch-irc/v4"
)

type LobotomizeCommand struct {
}

func (c *LobotomizeCommand) Name() string {
	return "lobotomize"
}

func (c *LobotomizeCommand) Permissions() []Permission {
	return []Permission{Streamer, Moderator}
}

func (c *LobotomizeCommand) Execute(ctx *Context, msg twitch.PrivateMessage, args string) {
	services.OllamaServiceInstance.Lobotomize(msg.Channel)

	ctx.Reply(msg.Channel, msg.ID, "pizza_son has been lobotomized meow")
}
