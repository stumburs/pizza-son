package services

import (
	"log"
	"pizza-son/internal/config"
	"sync"
	"time"

	"github.com/nicklaw5/helix/v2"
)

type TwitchService struct {
	client *helix.Client
	cache  map[string]cachedStreamInfo
	mu     sync.Mutex
}

type cachedStreamInfo struct {
	info      StreamInfo
	fetchedAt time.Time
}

type StreamInfo struct {
	GameName     string
	StreamTitle  string
	ChannelName  string
	ViewerCount  int
	ChannelTags  []string
	ThumbnailURL string
}

const cacheDuration = 10 * time.Minute

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
	if err != nil || token.Data.AccessToken == "" {
		log.Fatal("[Twitch] Failed to get app access token:", err, token.ErrorMessage)
	}

	log.Printf("[Twitch] Got access token, expires in %d seconds", token.Data.ExpiresIn)
	client.SetAppAccessToken(token.Data.AccessToken)

	TwitchServiceInstance = &TwitchService{client: client}
	TwitchServiceInstance.startTokenRefresh(client, token.Data.ExpiresIn)
	log.Println("[Twitch] Service initialized")
}

func (s *TwitchService) startTokenRefresh(client *helix.Client, expiresIn int) {
	go func() {
		for {
			time.Sleep(time.Duration(expiresIn-300) * time.Second)

			token, err := client.RequestAppAccessToken([]string{})
			if err != nil || token.Data.AccessToken == "" {
				log.Println("[Twitch] Failed to refresh app token:", err)
				time.Sleep(30 * time.Second)
				continue
			}
			client.SetAppAccessToken(token.Data.AccessToken)
			expiresIn = token.Data.ExpiresIn
			log.Printf("[Twitch] App token refreshed, expires in %d seconds", expiresIn)
		}
	}()
}

func (s *TwitchService) GetStreamInfo(channel string) StreamInfo {
	s.mu.Lock()

	if s.cache == nil {
		s.cache = make(map[string]cachedStreamInfo)
	}

	if cached, ok := s.cache[channel]; ok {
		if time.Since(cached.fetchedAt) < cacheDuration {
			s.mu.Unlock()
			log.Printf("[Twitch] Cache hit for %s", channel)
			return cached.info
		}
	}
	s.mu.Unlock()

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

	s.mu.Lock()
	s.cache[channel] = cachedStreamInfo{info: info, fetchedAt: time.Now()}
	s.mu.Unlock()

	log.Printf("[Twitch] Cached stream info for %s", channel)

	return info
}
