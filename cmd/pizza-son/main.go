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
	"time"
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
	services.NewMudwrestleService()
	services.NewPredictionService()
	services.NewSevenTVService()
	services.NewNotificationService()

	// Command registry
	registry := bot.NewRegistry(config.Get().Bot.Prefix)
	commands.SetRegistry(registry)

	// Twitch bot
	twitchBot := bot.New(
		config.Get().Twitch.User,
		config.Get().Bot.Channels,
		registry,
	)

	services.TwitchServiceInstance.SetTokenRefreshCallback(func(newToken string) {
		twitchBot.Reconnect(newToken)
	})

	// Discord bot
	discordBot, err := bot.NewDiscordBot(
		config.Get().Discord.Token,
		config.Get().Discord.Channels,
		registry,
	)
	if err != nil {
		log.Fatalf("[Main] Failed to initialize Discord bot: %v", err)
	}

	// Run twitch in background
	go func() {
		for {
			if err := twitchBot.Start(); err != nil {
				log.Println("[Twitch] Disconnected:", err)
			}
			time.Sleep(2 * time.Second)
		}
	}()

	// Run Discord in background
	go func() {
		if err := discordBot.Start(); err != nil {
			log.Fatalf("[Discord] Failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc

	log.Println("Shutting down...")
	discordBot.Stop()
	services.LoggerServiceInstance.Close()
}
