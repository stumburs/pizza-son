package main

import (
	"log"
	"pizza-son/internal/bot"
	"pizza-son/internal/commands"
	"pizza-son/internal/config"
	"pizza-son/internal/services"
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/stumburs/mgo"
)

func main() {
	log.Println("Loading config...")
	config.Load("config.toml")

	log.Println("Initializing Markov...")
	generator := mgo.NewMarkovGenerator()
	generator.ReadNgrams("data.bin")

	log.Println("Initializing Ollama...")
	services.NewOllamaService()

	log.Println("Creating client...")
	client := twitch.NewClient(config.Get().Twitch.User, config.Get().Twitch.OAuth)

	ctx := &commands.Context{
		Reply: func(channel, replyToID, message string) {
			client.Reply(channel, replyToID, message)
		},
	}

	router := bot.NewRouter(ctx)

	// !mark
	router.Register(&commands.MarkCommand{
		Generator: generator,
	})
	// !iq
	router.Register(&commands.IQCommand{})
	// !llm
	router.Register(&commands.LLMCommand{})
	// !cat
	router.Register(&commands.CatCommand{})
	// !nsfw
	router.Register(&commands.NSFWCommand{})
	// !lark
	router.Register(&commands.LarkCommand{})
	// !joni
	router.Register(&commands.JoniCommand{})
	// !jark
	router.Register(&commands.JarkCommand{})
	// !math
	router.Register(&commands.MathCommand{})
	// !lobotomize
	router.Register(&commands.LobotomizeCommand{})

	// fricc hook
	router.AddHook(func(msg twitch.PrivateMessage) {
		if strings.Contains(strings.ToLower(msg.Message), "fricc") {
			router.Ctx.Reply(msg.Channel, msg.ID, "fricc u too")
		}
	})

	// meow hook
	router.AddHook(func(msg twitch.PrivateMessage) {
		if msg.Message == "meow" {
			router.Ctx.Reply(msg.Channel, msg.ID, "meow")
		}
	})

	// ollama on message
	router.AddHook(func(msg twitch.PrivateMessage) {
		if strings.HasPrefix(msg.Message, "!") {
			return
		}
		services.OllamaServiceInstance.OnPrivateMessage(msg)
	})

	client.OnPrivateMessage(router.HandleMessage)
	client.OnConnect(func() {
		log.Println("Connected!")
	})

	log.Println("Joining channel(s):", config.Get().Bot.Channels)
	client.Join(config.Get().Bot.Channels...)

	log.Println("Connecting...")
	if err := client.Connect(); err != nil {
		panic(err)
	}
}
