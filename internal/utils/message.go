package utils

import (
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
)

func GetMessageParts(msg twitch.PrivateMessage) []string {
	parts := strings.SplitN(msg.Message[1:], " ", 2)
	return parts
}

func GetMessageCommandName(msg twitch.PrivateMessage) string {
	parts := GetMessageParts(msg)
	if !strings.HasPrefix(msg.Message, "!") {
		return ""
	}
	name := parts[0]
	return name
}

func GetMessageArgs(msg twitch.PrivateMessage) string {
	parts := GetMessageParts(msg)

	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}
	return args
}
