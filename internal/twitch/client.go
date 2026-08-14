package twitch

import (
	"fmt"
	"pizza-son/internal/config"

	"github.com/gempir/go-twitch-irc/v4"
)

type TwitchClient struct {
	client   *twitch.Client
	channels []string
}

func New(config *config.Config) *TwitchClient {
	client := twitch.NewClient(config.Auth.Twitch.Username, config.Auth.Twitch.OAuth)

	tc := &TwitchClient{
		client:   client,
		channels: []string{"pizza_tm"},
	}

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		fmt.Println("Twitch: " + message.Message)
	})

	return tc
}

func (tc *TwitchClient) Run() error {
	tc.client.Join(tc.channels...)
	return tc.client.Connect()
}
