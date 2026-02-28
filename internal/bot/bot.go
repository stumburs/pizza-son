package bot

import (
	"log"

	"github.com/gempir/go-twitch-irc/v4"
)

type Bot struct {
	client   *twitch.Client
	registry *Registry
	channels []string
}

func New(username, oauth string, channels []string, registry *Registry) *Bot {
	client := twitch.NewClient(username, oauth)
	return &Bot{
		client:   client,
		registry: registry,
		channels: channels,
	}
}

func (b *Bot) Start() error {
	b.client.OnPrivateMessage(func(msg twitch.PrivateMessage) {
		b.registry.Dispatch(b.client, msg)
	})

	for _, ch := range b.channels {
		b.client.Join(ch)
	}

	log.Println("[Bot] Bot connecting to:", b.channels)
	return b.client.Connect()
}
