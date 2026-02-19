package commands

import (
	"pizza-son/internal/services"
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
)

type NSFWCommand struct {
}

func (c *NSFWCommand) Name() string {
	return "nsfw"
}

func (c *NSFWCommand) Permissions() []Permission {
	return []Permission{All}
}

func (c *NSFWCommand) Execute(ctx *Context, msg twitch.PrivateMessage, args string) {
	res, err := services.OllamaServiceInstance.GenerateChatResponse(msg, strings.TrimSpace(args))
	if err != nil {
		ctx.Reply(msg.Channel, msg.ID, *res.Message.Content)
	}
	ctx.Reply(msg.Channel, msg.ID, *res.Message.Content)
}
