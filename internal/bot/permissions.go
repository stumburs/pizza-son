package bot

import (
	"slices"

	"github.com/gempir/go-twitch-irc/v4"
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

// TODO: Add config file support
var botModerators = []string{"pizza_tm"}

func HasPermission(msg twitch.PrivateMessage, required Permission) bool {
	return GetPermissionLevel(msg) >= required
}

func GetPermissionLevel(msg twitch.PrivateMessage) Permission {
	if msg.User.IsBroadcaster {
		return Streamer
	}
	if slices.Contains(botModerators, msg.User.Name) {
		return BotModerator
	}
	if msg.User.IsMod {
		return Moderator
	}
	if msg.User.IsVip {
		return VIP
	}
	if msg.User.Badges["subscriber"] > 0 {
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
