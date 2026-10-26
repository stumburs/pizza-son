package services

import (
	"fmt"
	"log"
	"pizza-son/internal/config"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/nicklaw5/helix/v2"
)

type TwitchService struct {
	client         *helix.Client
	cache          map[string]cachedStreamInfo
	liveCache      map[string]StreamInfo
	liveCacheAt    map[string]time.Time
	mu             sync.Mutex
	onTokenRefresh func(string)
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
	StartedAt    time.Time
}

const (
	cacheDuration = 1 * time.Minute
	// liveCacheDuration keeps the public site from hitting Helix on every
	// page load while still feeling live.
	liveCacheDuration = 30 * time.Second
)

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
		client:      client,
		cache:       map[string]cachedStreamInfo{},
		liveCache:   map[string]StreamInfo{},
		liveCacheAt: map[string]time.Time{},
	}
	TwitchServiceInstance.startTokenRefresh()
	log.Println("[Twitch] Service initialized")
}

func (s *TwitchService) SetTokenRefreshCallback(fn func(string)) {
	s.onTokenRefresh = fn
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

			if s.onTokenRefresh != nil {
				s.onTokenRefresh(resp.Data.AccessToken)
			}

			log.Printf("[Twitch] App token refreshed, expires in %d seconds", resp.Data.ExpiresIn)
		}
	}()
}

func (s *TwitchService) GetAccessToken() string {
	return s.client.GetUserAccessToken()
}

func (s *TwitchService) GetStreamInfo(channel string) StreamInfo {
	info, _ := s.GetStreamInfoErr(channel)
	return info
}

// GetStreamInfoErr returns the stream state and distinguishes "offline" from
// "Twitch could not be reached". Callers that track uptime must not treat an
// error as an offline stream.
func (s *TwitchService) GetStreamInfoErr(channel string) (StreamInfo, error) {
	s.mu.Lock()
	if cached, ok := s.cache[channel]; ok && time.Since(cached.fetchedAt) < cacheDuration {
		s.mu.Unlock()
		return cached.info, nil
	}
	s.mu.Unlock()

	if s.client == nil {
		return StreamInfo{ChannelName: channel}, fmt.Errorf("twitch client not initialized")
	}

	usersResp, err := s.client.GetUsers(&helix.UsersParams{
		Logins: []string{channel},
	})
	if err != nil {
		log.Printf("[Twitch] Could not resolve user ID for %s: %v", channel, err)
		return StreamInfo{ChannelName: channel}, err
	}
	if len(usersResp.Data.Users) == 0 {
		return StreamInfo{ChannelName: channel}, fmt.Errorf("twitch user not found: %s", channel)
	}

	broadcasterID := usersResp.Data.Users[0].ID

	chanResp, err := s.client.GetChannelInformation(&helix.GetChannelInformationParams{
		BroadcasterIDs: []string{broadcasterID},
	})
	if err != nil {
		log.Printf("[Twitch] Could not get channel info for %s: %v", channel, err)
		return StreamInfo{ChannelName: channel}, err
	}
	if len(chanResp.Data.Channels) == 0 {
		return StreamInfo{ChannelName: channel}, fmt.Errorf("twitch channel info not found: %s", channel)
	}
	ch := chanResp.Data.Channels[0]

	info := StreamInfo{
		GameName:    ch.GameName,
		StreamTitle: ch.Title,
		ChannelName: ch.BroadcasterName,
		ChannelTags: ch.Tags,
	}

	// Get live stream data if possible. An error here means we do not know
	// whether the channel is live, so do not cache a bogus offline result.
	streamResp, err := s.client.GetStreams(&helix.StreamsParams{
		UserLogins: []string{channel},
	})
	if err != nil {
		log.Printf("[Twitch] Could not get live stream for %s: %v", channel, err)
		return info, err
	}
	if streamResp.ErrorMessage != "" {
		err := fmt.Errorf("%s", streamResp.ErrorMessage)
		log.Printf("[Twitch] Could not get live stream for %s: %v", channel, err)
		return info, err
	}
	if len(streamResp.Data.Streams) > 0 {
		info.ViewerCount = streamResp.Data.Streams[0].ViewerCount
		info.ThumbnailURL = streamResp.Data.Streams[0].ThumbnailURL
		info.StartedAt = streamResp.Data.Streams[0].StartedAt
	}

	s.mu.Lock()
	s.cache[channel] = cachedStreamInfo{info: info, fetchedAt: time.Now()}
	s.mu.Unlock()

	return info, nil
}

func (s *TwitchService) GetStreamInfoFresh(channel string) StreamInfo {
	info, _ := s.GetStreamInfoFreshErr(channel)
	return info
}

func (s *TwitchService) GetStreamInfoFreshErr(channel string) (StreamInfo, error) {
	s.mu.Lock()
	delete(s.cache, channel)
	s.mu.Unlock()
	return s.GetStreamInfoErr(channel)
}

// GetLiveStreams fetches live status for many channels in a single Helix
// request. Results are cached briefly so the public site can poll freely. The
// returned map is keyed by lowercase channel login and only contains channels
// that are currently live. An error means at least one Helix request failed,
// so callers that track stream time should leave existing sessions untouched.
func (s *TwitchService) GetLiveStreams(channels []string) (map[string]StreamInfo, error) {
	result := make(map[string]StreamInfo)
	if len(channels) == 0 {
		return result, nil
	}

	normalized := make([]string, 0, len(channels))
	seen := make(map[string]struct{}, len(channels))
	for _, ch := range channels {
		ch = strings.ToLower(strings.TrimSpace(ch))
		if ch == "" {
			continue
		}
		if _, ok := seen[ch]; ok {
			continue
		}
		seen[ch] = struct{}{}
		normalized = append(normalized, ch)
	}
	if len(normalized) == 0 {
		return result, nil
	}

	now := time.Now()
	var stale []string
	s.mu.Lock()
	for _, ch := range normalized {
		if at, ok := s.liveCacheAt[ch]; !ok || now.Sub(at) >= liveCacheDuration {
			stale = append(stale, ch)
		}
	}
	if len(stale) == 0 {
		for _, ch := range normalized {
			if info, ok := s.liveCache[ch]; ok && info.ViewerCount > 0 {
				result[ch] = info
			}
		}
		s.mu.Unlock()
		return result, nil
	}
	s.mu.Unlock()

	if s.client == nil {
		return result, fmt.Errorf("twitch client not initialized")
	}

	fetched := make(map[string]StreamInfo)
	settled := make(map[string]bool, len(stale))
	var firstErr error
	for start := 0; start < len(stale); start += 100 {
		end := start + 100
		if end > len(stale) {
			end = len(stale)
		}
		chunk := stale[start:end]

		resp, err := s.client.GetStreams(&helix.StreamsParams{
			UserLogins: chunk,
			Type:       "live",
			First:      100,
		})
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			log.Printf("[Twitch] Could not fetch live streams: %v", err)
			continue
		}
		if resp.ErrorMessage != "" {
			if firstErr == nil {
				firstErr = fmt.Errorf("%s", resp.ErrorMessage)
			}
			log.Printf("[Twitch] Could not fetch live streams: %s", resp.ErrorMessage)
			continue
		}

		for _, ch := range chunk {
			settled[ch] = true
		}
		for _, stream := range resp.Data.Streams {
			login := strings.ToLower(stream.UserLogin)
			if login == "" {
				continue
			}
			fetched[login] = StreamInfo{
				GameName:     stream.GameName,
				StreamTitle:  stream.Title,
				ChannelName:  stream.UserName,
				ViewerCount:  stream.ViewerCount,
				ChannelTags:  stream.Tags,
				ThumbnailURL: stream.ThumbnailURL,
				StartedAt:    stream.StartedAt,
			}
		}
	}

	fetchedAt := time.Now()
	s.mu.Lock()
	for _, ch := range stale {
		if !settled[ch] {
			continue
		}
		if info, ok := fetched[ch]; ok {
			s.liveCache[ch] = info
		} else {
			delete(s.liveCache, ch)
		}
		s.liveCacheAt[ch] = fetchedAt
	}
	for _, ch := range normalized {
		if info, ok := s.liveCache[ch]; ok && info.ViewerCount > 0 {
			result[ch] = info
		}
	}
	s.mu.Unlock()

	return result, firstErr
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

func (s *TwitchService) GetModerators(broadcasterID, userID string) (bool, error) {
	resp, err := s.client.GetModerators(&helix.GetModeratorsParams{
		BroadcasterID: broadcasterID,
		UserIDs:       []string{userID},
	})
	if err != nil {
		return false, err
	}
	if resp.ErrorMessage != "" {
		return false, fmt.Errorf("%s", resp.ErrorMessage)
	}
	return len(resp.Data.Moderators) > 0, nil
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
	// Skip supproters
	username, err := s.GetUsername(userID)
	if err == nil {
		username = strings.ToLower(username)
		supporters := config.Get().Bot.Supporters
		botMods := config.Get().Bot.Moderators
		if slices.Contains(supporters, username) || slices.Contains(supporters, userID) ||
			slices.Contains(botMods, username) || slices.Contains(botMods, userID) {
			log.Printf("[Twitch] Skipping timeout for protected user %s", username)
			return
		}
	}

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

func (s *TwitchService) Ban(channel, userID string, reason string) {
	// Skip supproters
	username, err := s.GetUsername(userID)
	if err == nil {
		username = strings.ToLower(username)
		supporters := config.Get().Bot.Supporters
		botMods := config.Get().Bot.Moderators
		if slices.Contains(supporters, username) || slices.Contains(supporters, userID) ||
			slices.Contains(botMods, username) || slices.Contains(botMods, userID) {
			log.Printf("[Twitch] Skipping timeout for protected user %s", username)
			return
		}
	}

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
			UserId: userID,
			Reason: reason,
		},
	})
	if err != nil || resp.ErrorMessage != "" {
		log.Printf("[Twitch] Failed to ban %s: %v %s", userID, err, resp.ErrorMessage)
	}
}

func (s *TwitchService) Shoutout(channel, raiderUsername string) {
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

	raiderID, err := s.GetUserID(raiderUsername)
	if err != nil {
		log.Printf("[Twitch] Failed to resolve raider ID for %s: %v", raiderUsername, err)
		return
	}

	resp, err := s.client.SendShoutout(&helix.SendShoutoutParams{
		FromBroadcasterID: broadcasterID,
		ToBroadcasterID:   raiderID,
		ModeratorID:       botID,
	})

	if err != nil || resp.ErrorMessage != "" {
		log.Printf("[Twitch] Failed to execute shoutout for %s: %v %s", raiderUsername, err, resp.ErrorMessage)
		return
	}

	log.Printf("[Twitch] Auto-shoutout successfully triggered for %s in channel #%s", raiderUsername, channel)
}
