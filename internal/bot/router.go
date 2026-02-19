package bot

import (
	"pizza-son/internal/commands"
	"pizza-son/internal/utils"
	"strings"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
)

type Router struct {
	Commands map[string]commands.Command
	Ctx      *commands.Context

	// Runs on every message before commands
	MessageHooks []func(msg twitch.PrivateMessage)
	Cooldowns    map[string]time.Time
}

func NewRouter(ctx *commands.Context) *Router {
	return &Router{
		Commands:  make(map[string]commands.Command),
		Ctx:       ctx,
		Cooldowns: make(map[string]time.Time),
	}
}

func (r *Router) Register(cmd commands.Command) {
	r.Commands[cmd.Name()] = cmd
}

func (r *Router) AddHook(hook func(msg twitch.PrivateMessage)) {
	r.MessageHooks = append(r.MessageHooks, hook)
}

func (r *Router) HandleMessage(msg twitch.PrivateMessage) {
	// Hooks
	for _, hook := range r.MessageHooks {
		go hook(msg)
	}

	// Commands
	if !strings.HasPrefix(msg.Message, "!") {
		return
	}

	commandName := utils.GetMessageCommandName(msg)
	args := utils.GetMessageArgs(msg)

	cmd, ok := r.Commands[commandName]
	if !ok {
		return
	}

	// Permissions
	if !r.hasPermission(cmd.Permissions(), msg) {
		r.Ctx.Reply(msg.Channel, msg.ID, "You don't have permission to use this command.")
		return
	}

	// Cooldown per command per user
	key := msg.User.Name + ":" + cmd.Name()
	if t, exists := r.Cooldowns[key]; exists && time.Since(t) < 5*time.Second {
		return
	}
	r.Cooldowns[key] = time.Now()

	go cmd.Execute(r.Ctx, msg, args)
}

func (r *Router) hasPermission(levels []commands.Permission, msg twitch.PrivateMessage) bool {
	for _, level := range levels {
		switch level {
		case commands.All:
			return true
		case commands.Subscriber:
			if msg.User.Badges["subscriber"] == 1 {
				return true
			}
		case commands.VIP:
			if msg.User.IsVip {
				return true
			}
		case commands.Moderator:
			if msg.User.IsMod {
				return true
			}
		case commands.Streamer:
			if msg.User.IsBroadcaster {
				return true
			}
		}
	}
	return false
}
