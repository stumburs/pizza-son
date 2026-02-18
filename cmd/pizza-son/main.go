package main

import (
	"log"
	"pizza-son/internal/bot"
	"pizza-son/internal/commands"
	"pizza-son/internal/config"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/stumburs/mgo"
)

func main() {

	log.Println("Initializing Markov...")
	generator := mgo.NewMarkovGenerator()
	generator.ReadNgrams("data.bin")

	log.Println("Loading config...")
	cfg, err := config.Load("config.toml")
	if err != nil {
		panic(err)
	}

	log.Println("Creating client...")
	client := twitch.NewClient(cfg.Twitch.User, cfg.Twitch.OAuth)

	ctx := &commands.Context{
		Reply: func(channel, replyToID, message string) {
			client.Reply(channel, replyToID, message)
		},
	}

	router := bot.NewRouter(ctx)

	router.Register(&commands.MarkCommand{
		Generator: generator,
	})

	client.OnPrivateMessage(router.HandleMessage)
	client.OnConnect(func() {
		log.Println("Connected!")
	})

	log.Println("Joining channel(s):", cfg.Bot.Channels)
	client.Join(cfg.Bot.Channels...)

	log.Println("Connecting...")
	if err := client.Connect(); err != nil {
		panic(err)
	}
}
