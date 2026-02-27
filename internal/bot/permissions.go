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
	if required == All {
		return true
	}

	userLevel := All

	if msg.User.Badges["subscriber"] > 0 {
		userLevel = Subscriber
	}
	if msg.User.IsVip {
		userLevel = VIP
	}
	if msg.User.IsMod {
		userLevel = Moderator
	}
	if slices.Contains(botModerators, msg.User.Name) {
		userLevel = BotModerator
	}
	if msg.User.IsBroadcaster {
		userLevel = Streamer
	}

	return userLevel >= required
}
