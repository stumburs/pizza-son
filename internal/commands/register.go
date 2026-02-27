package commands

import "pizza-son/internal/bot"

var (
	globalRegistry   *bot.Registry
	pendingCommands  []bot.Command
	pendingListeners []bot.ListenerEntry
)

func SetRegistry(r *bot.Registry) {
	globalRegistry = r

	for _, cmd := range pendingCommands {
		globalRegistry.Register(cmd)
	}
	pendingCommands = nil

	for _, l := range pendingListeners {
		globalRegistry.RegisterListener(l)
	}
	pendingListeners = nil
}

func Register(cmd bot.Command) {
	if globalRegistry == nil {
		pendingCommands = append(pendingCommands, cmd)
		return
	}
	globalRegistry.Register(cmd)
}

func RegisterListener(l bot.ListenerEntry) {
	if globalRegistry == nil {
		pendingListeners = append(pendingListeners, l)
		return
	}
	globalRegistry.RegisterListener(l)
}
