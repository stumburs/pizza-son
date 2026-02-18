package bot

import (
	"pizza-son/internal/commands"
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
)

type Router struct {
	Commands map[string]commands.Command
	Ctx      *commands.Context
}

func NewRouter(ctx *commands.Context) *Router {
	return &Router{
		Commands: make(map[string]commands.Command),
		Ctx:      ctx,
	}
}

func (r *Router) Register(cmd commands.Command) {
	r.Commands[cmd.Name()] = cmd
}

func (r *Router) HandleMessage(msg twitch.PrivateMessage) {
	if !strings.HasPrefix(msg.Message, "!") {
		return
	}

	parts := strings.SplitN(msg.Message[1:], " ", 2)
	name := parts[0]

	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}

	if cmd, ok := r.Commands[name]; ok {
		go cmd.Execute(r.Ctx, msg, args)
	}
}
