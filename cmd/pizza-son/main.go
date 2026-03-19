package main

import (
	"log"
	"os"
	"os/signal"
	"pizza-son/internal/bot"
	"pizza-son/internal/commands"
	"pizza-son/internal/config"
	"pizza-son/internal/services"
	"syscall"
)

func main() {
	config.Load("config.toml")

	// Services
	services.NewTwitchService()
	services.NewOllamaService()
	services.NewMarkovService()
	services.NewAdaService()
	services.NewQuoteService()
	services.NewCurrencyService()
	services.NewRNNService()
	services.NewLoggerService()
	services.NewChannelSettingsService()

	// Command registry
	registry := bot.NewRegistry(config.Get().Bot.Prefix)
	commands.SetRegistry(registry)

	// Bot
	b := bot.New(
		config.Get().Twitch.User,
		config.Get().Bot.Channels,
		registry,
	)

	services.TwitchServiceInstance.SetTokenRefreshCallback(func(newToken string) {
		b.Reconnect(newToken)
	})

	go func() {
		if err := b.Start(); err != nil {
			log.Fatal(err)
		}
	}()

	// Graceful shutdown
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc

	log.Println("Shutting down...")
	services.LoggerServiceInstance.Close()
}
