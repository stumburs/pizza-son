package commands

import (
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
)

type IQCommand struct {
}

func (c *IQCommand) Name() string {
	return "iq"
}

func (c *IQCommand) Permissions() []Permission {
	return []Permission{All}
}

func (c *IQCommand) Execute(ctx *Context, msg twitch.PrivateMessage, args string) {
	iq_amount := rand.IntN(300)

	text := strings.TrimSpace(args)
	if text == "" {
		ctx.Reply(msg.Channel, msg.ID, fmt.Sprintf("%s has %d IQ xddNerd", msg.User.DisplayName, iq_amount))
		return
	} else {
		ctx.Reply(msg.Channel, msg.ID, fmt.Sprintf("%s has %d IQ xddNerd", text, iq_amount))
		return
	}
}
