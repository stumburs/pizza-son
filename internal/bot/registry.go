package bot

import (
	"pizza-son/internal/config"
	"strings"
	"sync"

	"github.com/gempir/go-twitch-irc/v4"
)

type CommandContext struct {
	Client  *twitch.Client
	Message twitch.PrivateMessage
	Args    []string // command args
}

type CommandHandler func(ctx CommandContext)
type Listener func(ctx CommandContext) bool

type Command struct {
	Name        string
	Description string
	Usage       string
	Permission  Permission
	Handler     CommandHandler
}

type ListenerEntry struct {
	Name        string
	Description string
	Permission  Permission
	Handler     Listener
}

// All commands
type Registry struct {
	mu        sync.RWMutex
	commands  map[string]Command
	listeners []ListenerEntry
	Prefix    string
}

func NewRegistry(prefix string) *Registry {
	return &Registry{
		commands: make(map[string]Command),
		Prefix:   prefix,
	}
}

func (r *Registry) Register(cmd Command) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands[strings.ToLower(cmd.Name)] = cmd
}

func (r *Registry) RegisterListener(l ListenerEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.listeners = append(r.listeners, l)
}

func (r *Registry) Dispatch(client *twitch.Client, msg twitch.PrivateMessage) {

	// Ignored users
	for _, ignored := range config.Get().Bot.IgnoredUsers {
		if strings.EqualFold(msg.User.Name, ignored) {
			return
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	ctx := CommandContext{
		Client:  client,
		Message: msg,
	}

	// Run listeners on every message
	for _, l := range r.listeners {
		if !HasPermission(msg, l.Permission) {
			continue
		}
		l.Handler(ctx)
	}

	// Run commands if starts with prefix
	if !strings.HasPrefix(msg.Message, r.Prefix) {
		return
	}

	parts := strings.Fields(strings.TrimPrefix(msg.Message, r.Prefix))
	if len(parts) == 0 {
		return
	}

	name := strings.ToLower(parts[0])
	cmd, ok := r.commands[name]
	if !ok {
		return
	}

	if !HasPermission(msg, cmd.Permission) {
		client.Reply(msg.Channel, msg.ID, "You don't have permission to use that command.")
		return
	}

	ctx.Args = parts[1:]

	cmd.Handler(ctx)
}

func (r *Registry) Commands() map[string]Command {
	return r.commands
}
