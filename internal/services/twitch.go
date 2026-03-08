package services

import (
	"fmt"
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
		RedirectURI:  "https://twitchtokengenerator.com",
	})
	if err != nil {
		log.Fatal("[Twitch] Failed to create Helix client:", err)
	}

	resp, err := client.RefreshUserAccessToken(config.Get().Twitch.RefreshToken)
	if err != nil || resp.Data.AccessToken == "" {
		log.Fatal("[Twitch] Failed to refresh user token:", err, resp.ErrorMessage)
	}

	client.SetUserAccessToken(resp.Data.AccessToken)
	config.Get().Twitch.UserAccessToken = resp.Data.AccessToken
	config.Get().Twitch.RefreshToken = resp.Data.RefreshToken

	// Save to disk
	if err := config.Save(); err != nil {
		log.Println("[Twitch] Failed to save tokens:", err)
	} else {
		log.Println("[Twitch] User token refreshed and saved")
	}

	TwitchServiceInstance = &TwitchService{
		client: client,
		cache:  map[string]cachedStreamInfo{},
	}
	TwitchServiceInstance.startTokenRefresh()
	log.Println("[Twitch] Service initialized")
}

func (s *TwitchService) startTokenRefresh() {
	go func() {
		for {
			time.Sleep(3 * time.Hour)

			resp, err := s.client.RefreshUserAccessToken(config.Get().Twitch.RefreshToken)
			if err != nil || resp.Data.AccessToken == "" {
				log.Println("[Twitch] Failed to refresh app token:", err)
				continue
			}
			s.client.SetUserAccessToken(resp.Data.AccessToken)
			config.Get().Twitch.UserAccessToken = resp.Data.AccessToken
			config.Get().Twitch.RefreshToken = resp.Data.RefreshToken

			// Save to disk
			if err := config.Save(); err != nil {
				log.Println("[Twitch] Failed to save tokens:", err)
			} else {
				log.Println("[Twitch] User token refreshed and saved")
			}

			log.Printf("[Twitch] App token refreshed, expires in %d seconds", resp.Data.ExpiresIn)
		}
	}()
}

func (s *TwitchService) GetAccessToken() string {
	return s.client.GetUserAccessToken()
}

func (s *TwitchService) GetStreamInfo(channel string) StreamInfo {
	s.mu.Lock()

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

func (s *TwitchService) GetUserID(username string) (string, error) {
	resp, err := s.client.GetUsers(&helix.UsersParams{
		Logins: []string{username},
	})
	if err != nil || len(resp.Data.Users) == 0 {
		return "", fmt.Errorf("user not found: %s", username)
	}
	return resp.Data.Users[0].ID, nil
}

func (s *TwitchService) GetUsername(userID string) (string, error) {
	resp, err := s.client.GetUsers(&helix.UsersParams{
		IDs: []string{userID},
	})
	if err != nil || len(resp.Data.Users) == 0 {
		return "", fmt.Errorf("user not found: %s", userID)
	}
	return resp.Data.Users[0].DisplayName, nil
}

func (s *TwitchService) Timeout(channel, userID string, duration int, reason string) {
	broadcasterID, err := s.GetUserID(channel)
	if err != nil {
		log.Printf("[Twitch] Failed to resolve broadcaster ID for %s: %v", channel, err)
		return
	}

	botID, err := s.GetUserID(config.Get().Twitch.User)
	if err != nil {
		log.Printf("[Twitch] Failed to resolve bot user ID: %v", err)
		return
	}

	resp, err := s.client.BanUser(&helix.BanUserParams{
		BroadcasterID: broadcasterID,
		ModeratorId:   botID,
		Body: helix.BanUserRequestBody{
			UserId:   userID,
			Duration: duration,
			Reason:   reason,
		},
	})
	if err != nil || resp.ErrorMessage != "" {
		log.Printf("[Twitch] Failed to timeout %s: %v %s", userID, err, resp.ErrorMessage)
	}
}
