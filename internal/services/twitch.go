package services

import (
	"log"
	"pizza-son/internal/config"

	"github.com/nicklaw5/helix/v2"
)

type TwitchService struct {
	client *helix.Client
}

type StreamInfo struct {
	GameName     string
	StreamTitle  string
	ChannelName  string
	ViewerCount  int
	ChannelTags  []string
	ThumbnailURL string
}

var TwitchServiceInstance *TwitchService

func NewTwitchService() {
	client, err := helix.NewClient(&helix.Options{
		ClientID:     config.Get().Twitch.ClientID,
		ClientSecret: config.Get().Twitch.ClientSecret,
	})
	if err != nil {
		log.Fatal("[Twitch] Failed to create Helix client:", err)
	}

	token, err := client.RequestAppAccessToken([]string{})
	if err != nil {
		log.Fatal("[Twitch] Failed to get app access token:", err)
	}

	log.Printf("[Twitch] Got access token, expires in %d seconds", token.Data.ExpiresIn)
	client.SetAppAccessToken(token.Data.AccessToken)

	TwitchServiceInstance = &TwitchService{client: client}
	log.Println("[Twitch] Service initialized")
}

func (s *TwitchService) GetStreamInfo(channel string) StreamInfo {
	usersResp, err := s.client.GetUsers(&helix.UsersParams{
		Logins: []string{channel},
	})

	if err != nil || len(usersResp.Data.Users) == 0 {
		log.Printf("[Twitch] Could not resolve user ID for %s: %v", channel, err)
		return StreamInfo{ChannelName: channel}
	}

	broadcasterID := usersResp.Data.Users[0].ID

	chanResp, err := s.client.GetChannelInformation(&helix.GetChannelInformationParams{
		BroadcasterIDs: []string{broadcasterID},
	})
	if err != nil || len(chanResp.Data.Channels) == 0 {
		log.Printf("[Twitch] Could not get channel info for %s: %v", channel, err)
		return StreamInfo{ChannelName: channel}
	}
	ch := chanResp.Data.Channels[0]

	info := StreamInfo{
		GameName:    ch.GameName,
		StreamTitle: ch.Title,
		ChannelName: ch.BroadcasterName,
		ChannelTags: ch.Tags,
	}

	// Get live stream data if possible
	streamResp, err := s.client.GetStreams(&helix.StreamsParams{
		UserLogins: []string{channel},
	})
	if err == nil && len(streamResp.Data.Streams) > 0 {
		info.ViewerCount = streamResp.Data.Streams[0].ViewerCount
		info.ThumbnailURL = streamResp.Data.Streams[0].ThumbnailURL
	}

	return info
}
