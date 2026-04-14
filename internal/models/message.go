package models

type Platform string

const (
	PlatformTwitch  Platform = "twitch"
	PlatformDiscord Platform = "discord"
)

type Message struct {
	ID           string
	Channel      string
	Platform     Platform
	User         MessageUser
	Text         string
	Reply        *ParentMessage
	FirstMessage bool
}

type MessageUser struct {
	ID            string
	Name          string
	DisplayName   string
	IsBroadcaster bool
	IsMod         bool
	IsVIP         bool
	IsSubscriber  bool
}

type ParentMessage struct {
	ParentMsgID       string
	ParentMsgBody     string
	ParentDisplayName string
}
