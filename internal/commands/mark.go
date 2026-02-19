package commands

import (
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/stumburs/mgo"
)

type MarkCommand struct {
	Generator *mgo.MarkovGenerator
}

func (c *MarkCommand) Name() string {
	return "mark"
}

func (c *MarkCommand) Permissions() []Permission {
	return []Permission{All}
}

func (c *MarkCommand) Execute(ctx *Context, msg twitch.PrivateMessage, args string) {
	text := strings.TrimSpace(args)
	c.Generator.SourceText = text

	output := c.Generator.BuildNgrams(mgo.SplitByNCharacters, 4).GenerateText(100)

	ctx.Reply(msg.Channel, msg.ID, output)
}
