package main

import (
	"log"
	"net/url"
	"pizza-son/internal/bot"
	"pizza-son/internal/commands"
	"pizza-son/internal/config"
	"pizza-son/internal/services"
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/stumburs/mgo"
)

func main() {

	log.Println("Initializing Markov...")
	generator := mgo.NewMarkovGenerator()
	generator.ReadNgrams("data.bin")

	log.Println("Loading config...")
	config.Load("config.toml")

	log.Println("Initializing Ollama...")
	url, err := url.Parse("http://192.168.0.101:11434")
	if err != nil {
		panic(err)
	}
	services.InitOllamaService(*url, "mistral:latest", 80)

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
	// !llm
	router.Register(commands.NewAIPromptCommand("llm", commands.All))

	// fricc hook
	router.AddHook(func(msg twitch.PrivateMessage) {
		if strings.Contains(strings.ToLower(msg.Message), "fricc") {
			router.Ctx.Reply(msg.Channel, msg.ID, "fricc u too")
		}
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
