package bot

import (
	"pizza-son/internal/config"
	"pizza-son/internal/models"
	"pizza-son/internal/services"
	"strings"
	"sync"
	"time"
)

type Sender interface {
	Say(channel, message string)
	Reply(channel, msgID, message string)
}

type CommandContext struct {
	Client   Sender
	Message  models.Message
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
	Cooldown    time.Duration
	OnCooldown  func(ctx CommandContext, remaining time.Duration)
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
	Cooldown    time.Duration
	// CooldownFor, when set, overrides Cooldown with a duration calculated for
	// the current message. It runs after the matcher, so expensive checks only
	// happen when the listener actually matches.
	CooldownFor func(ctx CommandContext) time.Duration
	Matcher     func(ctx CommandContext) bool
	OnCooldown  func(ctx CommandContext, remaining time.Duration)
	Handler     Listener
	// Passive listeners observe messages without replying or otherwise
	// handling them. They are excluded from activity statistics.
	Passive bool
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

func (r *Registry) Dispatch(client Sender, msg models.Message) {
	// Ignored users
	for _, ignored := range config.Get().Bot.IgnoredUsers {
		if strings.EqualFold(msg.User.Name, ignored) {
			return
		}
	}

	// Remove whitespace including characters used to bypass message duplication checks by Twitch
	msg.Text = SanitizeText(msg.Text)
	message := strings.TrimSpace(msg.Text)

	// Trim leading @mention if this is a reply
	if msg.Reply != nil && msg.Reply.ParentMsgID != "" {
		parts := strings.SplitN(message, " ", 2)
		if len(parts) == 2 && strings.HasPrefix(parts[0], "@") {
			message = parts[1]
		}
	}

	// Record every chat message before dispatching. Commands are still chat
	// messages, but are flagged so the stats site can separate them.
	services.ActivityServiceInstance.LogMessage(msg, strings.HasPrefix(message, r.Prefix))

	r.mu.Lock()
	defer r.mu.Unlock()

	ctx := CommandContext{
		Client:   client,
		Message:  msg,
		Registry: r,
	}

	// Run listeners on every message
	for _, l := range r.listeners {
		if !HasPermission(msg, l.Permission) {
			continue
		}
		// Check for disabled listeners per channel
		if !services.ChannelSettingsInstance.IsListenerEnabled(msg.Channel, l.Name) {
			continue
		}
		l := l

		// cooldown
		go func(ctx CommandContext) {
			// match listener
			if l.Matcher != nil && !l.Matcher(ctx) {
				return
			}

			cooldown := l.Cooldown
			if l.CooldownFor != nil {
				cooldown = l.CooldownFor(ctx)
			}

			// cooldown
			if cooldown > 0 {
				if GlobalCooldowns.IsOnCooldown(l.Name, msg.Channel, msg.User.ID, cooldown) {
					if l.OnCooldown != nil {
						remaining := GlobalCooldowns.Remaining(l.Name, msg.Channel, msg.User.ID, cooldown)
						l.OnCooldown(ctx, remaining)
					}
					return
				}
			}
			handled := l.Handler(ctx)
			if handled && !l.Passive {
				services.ActivityServiceInstance.LogListener(msg, l.Name)
			}
			if cooldown > 0 {
				GlobalCooldowns.Set(l.Name, msg.Channel, msg.User.ID)
			}
		}(ctx)
	}

	if !strings.HasPrefix(message, r.Prefix) {
		return
	}

	parts := strings.Fields(strings.TrimPrefix(message, r.Prefix))
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

	// cooldowns
	if cmd.Cooldown > 0 {
		if GlobalCooldowns.IsOnCooldown(cmd.Name, msg.Channel, msg.User.ID, cmd.Cooldown) {
			if cmd.OnCooldown != nil {
				remaining := GlobalCooldowns.Remaining(cmd.Name, msg.Channel, msg.User.ID, cmd.Cooldown)
				go cmd.OnCooldown(ctx, remaining)
			}
			return
		}
		GlobalCooldowns.Set(cmd.Name, msg.Channel, msg.User.ID)
	}

	services.ActivityServiceInstance.LogCommand(msg, cmd.Name)
	go cmd.Handler(ctx)
}

func (r *Registry) Commands() map[string]Command {
	r.mu.RLock()
	defer r.mu.RUnlock()
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
	services.ActivityServiceInstance.LogCommand(ctx.Message, cmd.Name)
	go cmd.Handler(ctx)
}

func (r *Registry) Listeners() []ListenerEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.listeners
}
