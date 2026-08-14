package main

import (
	"fmt"
	"pizza-son/internal/config"

	"github.com/gempir/go-twitch-irc/v4"
)

func main() {

	config, err := config.Load("config.toml")
	if err != nil {
		panic(err)
	}

	client := twitch.NewClient(config.Auth.Twitch.Username, config.Auth.Twitch.OAuth)

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		fmt.Println(message.Message)
	})

	client.Join("pizza_tm")

	err = client.Connect()
	if err != nil {
		panic(err)
	}
}
