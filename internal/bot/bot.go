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
	token := "oauth:" + services.TwitchServiceInstance.GetAccessToken()
	client := twitch.NewClient(username, token)
	b := &Bot{
		client:   client,
		registry: registry,
		channels: channels,
	}
	b.setupHandlers()
	return b
}

func (b *Bot) Start() error {
	log.Println("[Bot] Bot connecting to:", b.channels)
	return b.client.Connect()
}

func (b *Bot) Reconnect(newToken string) {
	log.Println("[Bot] Reconnecting with new token...")
	b.client.Disconnect()
	b.client = twitch.NewClient(config.Get().Twitch.User, "oauth:"+newToken)
	b.setupHandlers()
	go b.client.Connect()
	log.Println("[Bot] Reconnected")
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
