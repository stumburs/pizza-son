package bot

import (
	"pizza-son/internal/config"
	"pizza-son/internal/models"
	"slices"
)

type Permission int

const (
	All Permission = iota
	Subscriber
	VIP
	Moderator
	BotModerator
	Streamer
)

func HasPermission(msg models.Message, required Permission) bool {
	return GetPermissionLevel(msg) >= required
}

func GetPermissionLevel(msg models.Message) Permission {
	if msg.User.IsBroadcaster {
		return Streamer
	}

	botMods := config.Get().Bot.Moderators

	if slices.Contains(botMods, msg.User.Name) || slices.Contains(botMods, msg.User.ID) {
		return BotModerator
	}
	if msg.User.IsMod {
		return Moderator
	}
	if msg.User.IsVIP {
		return VIP
	}
	if msg.User.IsSubscriber {
		return Subscriber
	}
	return All
}

func PermissionName(p Permission) string {
	switch p {
	case Subscriber:
		return "Subscriber"
	case VIP:
		return "VIP"
	case Moderator:
		return "Moderator"
	case BotModerator:
		return "Bot Moderator"
	case Streamer:
		return "Streamer"
	default:
		return "Everyone"
	}
}
