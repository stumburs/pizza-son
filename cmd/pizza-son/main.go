package main

import (
	"fmt"
	"pizza-son/internal/config"
	"pizza-son/internal/discord"
	"pizza-son/internal/twitch"
)

func main() {

	config, err := config.Load("config.toml")
	if err != nil {
		panic(err)
	}

	twitchClient := twitch.New(config)
	go func() {
		if err := twitchClient.Run(); err != nil {
			fmt.Println("twitch error:", err)
		}
	}()

	discordClient := discord.New(config)
	go func() {
		if err := discordClient.Run(); err != nil {
			fmt.Println("discord error:", err)
		}
	}()

	select {}
}
