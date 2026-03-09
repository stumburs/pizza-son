package bot

import (
	"pizza-son/internal/config"
	"pizza-son/internal/services"
	"strings"
	"sync"

	"github.com/gempir/go-twitch-irc/v4"
)

type Sender interface {
	Say(channel, message string)
	Reply(channel, msgID, message string)
}

type TwitchSender struct {
	client *twitch.Client
}

func (t *TwitchSender) Say(channel, message string) {
	t.client.Say(channel, message)
}

func (t *TwitchSender) Reply(channel, msgID, message string) {
	t.client.Reply(channel, msgID, message)
}

type CommandContext struct {
	Client   Sender
	Message  twitch.PrivateMessage
	Args     []string // command args
	Registry *Registry
}

type CommandHandler func(ctx CommandContext)
type Listener func(ctx CommandContext) bool

// Command categories
type Category int

const (
	CategoryUncategorized Category = iota
	CategoryAI
	CategoryCurrency
	CategoryFun
	CategoryGames
	CategoryModeration
	CategoryQuotes
	CategoryUtility
)

type Command struct {
	Name        string
	Description string
	Usage       string
	Permission  Permission
	Category    Category
	Examples    []CommandExample
	Handler     CommandHandler
}

type CommandExample struct {
	Input  string
	Output string
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
		Client:   &TwitchSender{client: client},
		Message:  msg,
		Registry: r,
	}

	// Run listeners on every message
	for _, l := range r.listeners {
		if !HasPermission(msg, l.Permission) {
			continue
		}
		l := l
		go l.Handler(ctx)
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

	if !services.ChannelSettingsInstance.IsCommandEnabled(msg.Channel, name) {
		return
	}

	if !HasPermission(msg, cmd.Permission) {
		client.Reply(msg.Channel, msg.ID, "You don't have permission to use that command.")
		return
	}

	ctx.Args = parts[1:]

	go cmd.Handler(ctx)
}

func (r *Registry) Commands() map[string]Command {
	return r.commands
}

func (c Category) String() string {
	switch c {
	case CategoryAI:
		return "AI"
	case CategoryCurrency:
		return "Currency"
	case CategoryFun:
		return "Fun"
	case CategoryGames:
		return "Games"
	case CategoryModeration:
		return "Moderation"
	case CategoryQuotes:
		return "Quotes"
	case CategoryUtility:
		return "Utility"
	default:
		return "Uncategorized"
	}
}

func (r *Registry) DispatchCommand(ctx CommandContext) {
	if len(ctx.Args) == 0 {
		return
	}

	name := strings.ToLower(ctx.Args[0])
	cmd, ok := r.commands[name]
	if !ok {
		return
	}
	ctx.Args = ctx.Args[1:]
	go cmd.Handler(ctx)
}
