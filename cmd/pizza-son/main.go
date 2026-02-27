package main

import (
	"log"
	"pizza-son/internal/bot"
	"pizza-son/internal/commands"
	"pizza-son/internal/config"
)

func main() {
	log.Println("Loading config...")
	config.Load("config.toml")

	// Command registry
	registry := bot.NewRegistry(config.Get().Bot.Prefix)
	commands.SetRegistry(registry)

	b := bot.New(config.Get().Twitch.User, config.Get().Twitch.OAuth, config.Get().Bot.Channels, registry)

	if err := b.Start(); err != nil {
		log.Fatal(err)
	}
}

// log.Println("Initializing Markov...")
// generator := mgo.NewMarkovGenerator()
// generator.ReadNgrams("data.bin")

// log.Println("Initializing Ollama...")
// services.NewOllamaService()

// log.Println("Creating client...")
// client := twitch.NewClient(config.Get().Twitch.User, config.Get().Twitch.OAuth)

// // fricc hook
// router.AddHook(func(msg *message.Message) {
// 	if strings.Contains(strings.ToLower(msg.Content), "fricc") {
// 		router.Ctx.Reply(msg.Channel, msg.MessageID, "fricc u too")
// 	}
// })

// // meow hook
// router.AddHook(func(msg *message.Message) {
// 	if msg.Content == "meow" {
// 		router.Ctx.Reply(msg.Channel, msg.MessageID, "meow")
// 	}
// })

// // ollama on message
// router.AddHook(func(msg *message.Message) {
// 	if msg.IsCommand {
// 		return
// 	}
// 	services.OllamaServiceInstance.OnPrivateMessage(msg.Raw)
// })

// client.OnPrivateMessage(router.HandleMessage)
// client.OnConnect(func() {
// 	log.Println("Connected!")
// })

// log.Println("Joining channel(s):", config.Get().Bot.Channels)
// client.Join(config.Get().Bot.Channels...)

// log.Println("Connecting...")
// if err := client.Connect(); err != nil {
// 	panic(err)
// }
// }
