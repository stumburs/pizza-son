package commands

import (
	"pizza-son/internal/services"
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
)

type MathCommand struct {
}

func (c *MathCommand) Name() string {
	return "math"
}

func (c *MathCommand) Permissions() []Permission {
	return []Permission{All}
}

func (c *MathCommand) Execute(ctx *Context, msg twitch.PrivateMessage, args string) {
	res, err := services.OllamaServiceInstance.GenerateChatResponse(msg, strings.TrimSpace(args))
	if err != nil {
		ctx.Reply(msg.Channel, msg.ID, *res.Message.Content)
	}
	ctx.Reply(msg.Channel, msg.ID, *res.Message.Content)
}
