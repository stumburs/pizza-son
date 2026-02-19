package bot

import (
	"pizza-son/internal/commands"
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

	parts := strings.SplitN(msg.Message[1:], " ", 2)
	name := parts[0]

	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}

	cmd, ok := r.Commands[name]
	if !ok {
		return
	}

	// Permissions
	if !r.hasPermission(cmd.Permission(), msg) {
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

func (r *Router) hasPermission(level commands.Permission, msg twitch.PrivateMessage) bool {
	switch level {
	case commands.All:
		return true
	case commands.Subscriber:
		return msg.User.Badges["subscriber"] == 1
	case commands.VIP:
		return msg.User.IsVip
	case commands.Moderator:
		return msg.User.IsMod
	case commands.Streamer:
		return msg.User.IsBroadcaster
	default:
		return false
	}
}
