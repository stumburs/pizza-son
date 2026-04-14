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
		wasLive := s.wasLive[n.TwitchChannel]

		if isLive && !wasLive {
			// Only notify if stream started within the last 5 minutes
			streamAge := time.Since(info.StartedAt)
			if streamAge <= 5*time.Minute {
				log.Printf("[Notifications] %s went live (%s ago), sending Discord notification", n.TwitchChannel, streamAge.Round(time.Second))
				go s.sendDiscord(n.DiscordWebhook, info)
			} else {
				log.Printf("[Notifications] %s is live but started %s ago, skipping notification", n.TwitchChannel, streamAge.Round(time.Second))
			}
		}
		s.wasLive[n.TwitchChannel] = isLive
		s.mu.Unlock()
	}
}

func (s *NotificationService) sendDiscord(webhookURL string, info StreamInfo) {
	payload := map[string]any{
		"embeds": []map[string]any{
			{
				"title":       info.ChannelName + " is now live!",
				"description": "@everyone " + info.StreamTitle,
				"url":         "https://twitch.tv/" + strings.ToLower(info.ChannelName),
				"color":       0x9146FF, // purple
				"fields": []map[string]any{
					{"name": "Game", "value": info.GameName, "inline": true},
					{"name": "Viewers", "value": fmt.Sprintf("%d", info.ViewerCount), "inline": true},
				},
				"image": map[string]string{
					"url": strings.ReplaceAll(info.ThumbnailURL, "{width}x{height}", "1280x720"),
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Println("[Notifications] Failed to marshal payload:", err)
		return
	}

	// ?wait=true to get message ID back
	resp, err := http.Post(webhookURL+"?wait=true", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Println("[Notifications] Failed to send Discord notification:", err)
		return
	}
	defer resp.Body.Close()

	var discordMsg struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&discordMsg); err != nil || discordMsg.ID == "" {
		log.Println("[Notifications] Failed to get message ID:", err)
		return
	}

	log.Printf("[Notifications] Message sent, ID: %s", discordMsg.ID)

	go s.updateUptime(webhookURL, discordMsg.ID, info)
}

func (s *NotificationService) updateUptime(webhookURL, messageID string, info StreamInfo) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// check if live
		current := TwitchServiceInstance.GetStreamInfo(strings.ToLower(info.ChannelName))
		if current.ViewerCount == 0 {
			log.Printf("[Notifications] %s is no longer live, stopping uptime updater", info.ChannelName)
			return
		}

		uptime := time.Since(info.StartedAt).Round(time.Minute)
		hours := int(uptime.Hours())
		mins := int(uptime.Minutes()) % 60

		uptimeStr := fmt.Sprintf("%dm", mins)
		if hours > 0 {
			uptimeStr = fmt.Sprintf("%dh %dm", hours, mins)
		}

		payload := map[string]any{
			"embeds": []map[string]any{
				{
					"title":       info.ChannelName + " is now live!",
					"description": "@everyone " + info.StreamTitle,
					"url":         "https://twitch.tv/" + strings.ToLower(info.ChannelName),
					"color":       0x9146FF, // purple
					"fields": []map[string]any{
						{"name": "Game", "value": info.GameName, "inline": true},
						{"name": "Viewers", "value": fmt.Sprintf("%d", info.ViewerCount), "inline": true},
						{"name": "Uptime", "value": uptimeStr, "inline": true},
					},
					"image": map[string]string{
						"url": strings.ReplaceAll(info.ThumbnailURL, "{width}x{height}", "1280x720"),
					},
				},
			},
		}

		body, err := json.Marshal(payload)
		if err != nil {
			continue
		}

		editURL := fmt.Sprintf("%s/messages/%s", webhookURL, messageID)
		req, err := http.NewRequest(http.MethodPatch, editURL, bytes.NewReader(body))
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("[Notifications] Failed to update message:", err)
			continue
		}
		resp.Body.Close()
		log.Printf("[Notifications] Updated uptime for %s: %s", info.ChannelName, uptimeStr)
	}
}
