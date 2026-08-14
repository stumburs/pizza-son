package platform

import "time"

type Message struct {
	Platform    string
	ChannelID   string
	UserID      string
	DisplayName string
	Content     string
	Timestamp   time.Time
	Raw         any
}
