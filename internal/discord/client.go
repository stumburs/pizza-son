package discord

import (
	"fmt"
	"pizza-son/internal/config"

	"github.com/bwmarrin/discordgo"
)

type DiscordClient struct {
	client   *discordgo.Session
	channels map[string]bool
}

func New(config *config.Config) *DiscordClient {
	discord, err := discordgo.New("Bot " + config.Auth.Discord.Token)
	// TODO: improve
	if err != nil {
		panic(err)
	}

	dc := &DiscordClient{
		client: discord,
		channels: map[string]bool{
			"1397219502718193804": true,
		},
	}

	return dc
}

func (dc *DiscordClient) Run() error {
	dc.client.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Only work in specified channels
		if !dc.channels[m.ChannelID] {
			return
		}

		fmt.Println("Discord: " + m.Content)
	})

	dc.client.Identify.Intents = discordgo.IntentGuildMessages |
		discordgo.IntentDirectMessages |
		discordgo.IntentGuildMembers |
		discordgo.IntentMessageContent

	if err := dc.client.Open(); err != nil {
		return err
	}
	return nil
}
