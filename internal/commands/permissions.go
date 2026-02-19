package commands

type Permission int

const (
	All Permission = iota
	Subscriber
	VIP
	Moderator
	Streamer
)
