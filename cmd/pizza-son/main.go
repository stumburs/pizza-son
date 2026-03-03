package main

import (
	"log"
	"os"
	"os/signal"
	"pizza-son/internal/bot"
	"pizza-son/internal/commands"
	"pizza-son/internal/config"
	"pizza-son/internal/services"
	"strings"
	"syscall"
)

func main() {
	log.Println("Loading config...")
	config.Load("config.toml")

	// Twitch API
	services.NewTwitchService()

	// Quotes
	services.NewQuoteService()

	// RNN Model
	services.NewRNNService()

	// Logger
	services.NewLoggerService()

	// Command registry
	registry := bot.NewRegistry(config.Get().Bot.Prefix)
	commands.SetRegistry(registry)

	// Ollama
	services.NewOllamaService()
	// Register on message listener
	registry.RegisterListener(bot.ListenerEntry{
		Name:        "ollama-context",
		Description: "Feeds chat messages to LLM context",
		Permission:  bot.All,
		Handler: func(ctx bot.CommandContext) bool {
			if strings.HasPrefix(ctx.Message.Message, config.Get().Bot.Prefix) {
				return false
			}
			go services.OllamaServiceInstance.OnPrivateMessage(ctx.Message)
			return true
		},
	})

	// Markov
	services.NewMarkovService()
	registry.RegisterListener(bot.ListenerEntry{
		Name:        "markov-learn",
		Description: "Feeds chat messages into Markov generator to train",
		Handler: func(ctx bot.CommandContext) bool {
			if strings.HasPrefix(ctx.Message.Message, config.Get().Bot.Prefix) {
				return false
			}
			go services.MarkovServiceInstance.Learn(ctx.Message.Channel, ctx.Message.Message)
			return true
		},
	})

	// Ada
	services.NewAdaService()

	b := bot.New(
		config.Get().Twitch.User,
		config.Get().Twitch.OAuth,
		config.Get().Bot.Channels,
		registry,
	)

	// Graceful shutdown
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := b.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	<-sc
	log.Println("Shutting down...")
	services.LoggerServiceInstance.Close()
}
