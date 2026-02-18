package main

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/commands"
	"pizza-son/internal/config"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/stumburs/mgo"
)

func main() {

	generator := mgo.NewMarkovGenerator()
	generator.ReadNgrams("data.bin")

	cfg, err := config.Load("config.toml")
	if err != nil {
		panic(err)
	}

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

	client.Join(cfg.Bot.Channels...)

	if err := client.Connect(); err != nil {
		panic(err)
	}
}
