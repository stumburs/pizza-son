package commands

import "github.com/gempir/go-twitch-irc/v4"

type Command interface {
	Name() string
	Permission() Permission
	Execute(ctx *Context, msg twitch.PrivateMessage, args string)
}

type Context struct {
	Reply func(channel, replyToID, message string)
}
