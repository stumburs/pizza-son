package bot

import (
	"log"
	"pizza-son/internal/config"
	"pizza-son/internal/services"

	"github.com/gempir/go-twitch-irc/v4"
)

type Bot struct {
	client   *twitch.Client
	registry *Registry
	channels []string
}

func New(username string, channels []string, registry *Registry) *Bot {
	return &Bot{
		registry: registry,
		channels: channels,
	}
}

func (b *Bot) Start() error {
	token := "oauth:" + services.TwitchServiceInstance.GetAccessToken()
	b.client = twitch.NewClient(config.Get().Twitch.User, token)
	b.setupHandlers()
	log.Println("[Bot] Bot connecting to:", b.channels)
	return b.client.Connect()
}

func (b *Bot) Reconnect(newToken string) {
	log.Println("[Bot] Token refreshed, disconnecting...")
	b.client.Disconnect()
	// Start() loop will reconnect automatically
}

func (b *Bot) setupHandlers() {
	b.client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		b.registry.Dispatch(b.client, message)
	})
	b.client.OnConnect(func() {
		log.Println("[Bot] Connected")
		for _, ch := range b.channels {
			b.client.Join(ch)
		}
	})
}
