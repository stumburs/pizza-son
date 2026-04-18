package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pizza-son/internal/config"
	"strings"
	"sync"
	"time"
)

type NotificationService struct {
	mu      sync.Mutex
	wasLive map[string]bool
}

var NotificationServiceInstance *NotificationService

func NewNotificationService() {
	NotificationServiceInstance = &NotificationService{
		wasLive: make(map[string]bool),
	}
	go NotificationServiceInstance.run()
	log.Println("[Notifications] Service initialized")
}

func (s *NotificationService) run() {
	// Check on startup, then every 1 minute
	s.check()
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.check()
	}
}

func (s *NotificationService) check() {
	for _, n := range config.Get().Notifications {
		info := TwitchServiceInstance.GetStreamInfo(n.TwitchChannel)
		isLive := info.ViewerCount > 0

		s.mu.Lock()
		alreadyTracked := s.wasLive[n.TwitchChannel]
		s.wasLive[n.TwitchChannel] = isLive
		s.mu.Unlock()

		if isLive && !alreadyTracked {
			// Only notify if stream started within the last 5 minutes
			streamAge := time.Since(info.StartedAt)
			if streamAge <= 5*time.Minute {
				log.Printf("[Notifications] %s went live (%s ago), sending Discord notification", n.TwitchChannel, streamAge.Round(time.Second))
				go s.handleLiveSession(n.DiscordWebhook, info)
			} else {
				log.Printf("[Notifications] %s is live but started %s ago, skipping notification", n.TwitchChannel, streamAge.Round(time.Second))
			}
		}
	}
}

func (s *NotificationService) handleLiveSession(webhookURL string, info StreamInfo) {
	// Initial message
	payload := s.buildPayload(info, false)
	messageID, err := s.executeRequest(http.MethodPost, webhookURL+"?wait=true", payload)
	if err != nil {
		log.Println("[Notifications] Initial send failed:", err)
		return
	}

	// Uptime updates
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	editURL := fmt.Sprintf("%s/messages/%s", webhookURL, messageID)

	for range ticker.C {
		current := TwitchServiceInstance.GetStreamInfoFresh(strings.ToLower(info.ChannelName))

		if current.ViewerCount == 0 {
			log.Printf("[Notifications] %s offline, stopping updater", info.ChannelName)
			return
		}

		updatePayload := s.buildPayload(current, true)

		_, err := s.executeRequest(http.MethodPatch, editURL, updatePayload)
		if err != nil {
			log.Println("[Notofications] Update failed:", err)
			continue
		}
		log.Printf("[Notifications] Updated uptime for %s", info.ChannelName)
	}
}

func (s *NotificationService) buildPayload(info StreamInfo, isUpdate bool) map[string]any {
	content := "@everyone"

	fields := []map[string]any{
		{"name": "Game", "value": info.GameName, "inline": true},
		{"name": "Viewers", "value": fmt.Sprintf("%d", info.ViewerCount), "inline": true},
	}

	if isUpdate {
		uptime := time.Since(info.StartedAt).Round(time.Minute)
		uptimeStr := fmt.Sprintf("%dm", int(uptime.Minutes())%60)
		if uptime.Hours() >= 1 {
			uptimeStr = fmt.Sprintf("%dh %dm", int(uptime.Hours()), int(uptime.Minutes())%60)
		}
		fields = append(fields, map[string]any{"name": "Uptime", "value": uptimeStr, "inline": true})
	}

	return map[string]any{
		"content": content,
		"embeds": []map[string]any{
			{
				"title":       info.ChannelName + " is now live!",
				"description": info.StreamTitle,
				"url":         "https://twitch.tv/" + strings.ToLower(info.ChannelName),
				"color":       0xE01B3C, // red
				"fields":      fields,
				"image": map[string]string{
					"url": strings.ReplaceAll(info.ThumbnailURL, "{width}x{height}", "1280x720"),
				},
			},
		},
	}
}

func (s *NotificationService) executeRequest(method, url string, payload map[string]any) (string, error) {
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("discord API error: %s", resp.Status)
	}

	// Only decode ID if we're doing a POST with ?wait=true
	if method == http.MethodPost {
		var discordMsg struct {
			ID string `json:"id"`
		}
		json.NewDecoder(resp.Body).Decode(&discordMsg)
		return discordMsg.ID, nil
	}

	return "", nil
}
