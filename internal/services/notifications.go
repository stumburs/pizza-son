package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"pizza-son/internal/config"
	"pizza-son/internal/logging"
	"pizza-son/internal/store"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	notificationsDir      = "data/notifications"
	notificationsFile     = "state.json"
	notificationInterval  = time.Minute
	notificationMaxAge    = 5 * time.Minute
	notificationHTTPLimit = 15 * time.Second
)

// NotificationRecord is the Discord message currently tracking one live
// stream. Persisting it lets the bot resume uptime edits after a restart and
// mark the message as offline when the stream ends.
type NotificationRecord struct {
	Channel    string     `json:"channel"`
	Webhook    string     `json:"webhook"`
	MessageID  string     `json:"message_id"`
	Message    string     `json:"message"`
	StartedAt  time.Time  `json:"started_at"`
	LastLiveAt time.Time  `json:"last_live_at"`
	Info       StreamInfo `json:"info"`
}

type notificationState struct {
	Records map[string]NotificationRecord `json:"records"`
}

type NotificationService struct {
	mu      sync.Mutex
	records map[string]NotificationRecord
	store   *store.Store[notificationState]
	client  *http.Client
}

var NotificationServiceInstance *NotificationService

func NewNotificationService() {
	state := &notificationState{Records: map[string]NotificationRecord{}}
	st := store.New(filepath.Join(notificationsDir, notificationsFile), state)
	if err := st.EnsureDir(); err != nil {
		log.Println("[Notifications] Failed to create state directory:", err)
	}
	if _, err := st.LoadIfExists(); err != nil {
		log.Println("[Notifications] Failed to load state:", err)
	}
	if state.Records == nil {
		state.Records = map[string]NotificationRecord{}
	}

	NotificationServiceInstance = &NotificationService{
		records: state.Records,
		store:   st,
		client:  &http.Client{Timeout: notificationHTTPLimit},
	}
	go NotificationServiceInstance.run()
	log.Println("[Notifications] Service initialized")
}

func (s *NotificationService) run() {
	// Check on startup, then every minute. On startup this both resumes
	// uptime edits for streams that are still live and finalizes messages for
	// streams that ended while the bot was down.
	s.check()
	ticker := time.NewTicker(notificationInterval)
	defer ticker.Stop()
	for range ticker.C {
		s.check()
	}
}

func (s *NotificationService) check() {
	cfg := config.Get()
	if cfg == nil || TwitchServiceInstance == nil {
		return
	}

	configured := make(map[string]struct{}, len(cfg.Notifications))
	for _, n := range cfg.Notifications {
		channel := strings.ToLower(strings.TrimSpace(n.TwitchChannel))
		if channel == "" || strings.TrimSpace(n.DiscordWebhook) == "" {
			continue
		}
		key := notificationKey(channel, n.DiscordWebhook)
		configured[key] = struct{}{}
		s.checkOne(key, channel, n)
	}

	s.prune(configured)
}

func (s *NotificationService) checkOne(key, channel string, n config.Notification) {
	info, err := TwitchServiceInstance.GetStreamInfoFreshErr(channel)
	if err != nil {
		// A failed lookup is not an offline stream. Leaving the record alone
		// means the next successful check can continue the uptime.
		logging.Debugf("[Notifications] %s: Twitch lookup failed: %v", channel, err)
		return
	}
	if strings.TrimSpace(info.ChannelName) == "" {
		info.ChannelName = channel
	}

	s.mu.Lock()
	record, hasRecord := s.records[key]
	s.mu.Unlock()

	if info.ViewerCount <= 0 {
		if hasRecord {
			s.finish(key, record)
		}
		return
	}

	if hasRecord && record.MessageID != "" && !sameStream(record.StartedAt, info.StartedAt) {
		// A new stream started before the old one was marked offline.
		s.finish(key, record)
		hasRecord = false
	}

	if !hasRecord || record.MessageID == "" {
		if info.StartedAt.IsZero() || time.Since(info.StartedAt) > notificationMaxAge {
			logging.Debugf("[Notifications] %s is live but started %s ago, skipping notification", channel, time.Since(info.StartedAt).Round(time.Second))
			return
		}
		s.start(key, channel, n, info)
		return
	}

	s.update(key, record, info)
}

func (s *NotificationService) start(key, channel string, n config.Notification, info StreamInfo) {
	postURL, err := webhookWithQuery(n.DiscordWebhook, "wait", "true")
	if err != nil {
		log.Printf("[Notifications] Invalid webhook for %s: %v", channel, err)
		return
	}

	messageID, err := s.executeRequest(http.MethodPost, postURL, buildLivePayload(info, false, n.Message))
	if err != nil {
		log.Printf("[Notifications] Initial send failed for %s: %v", channel, err)
		return
	}
	if messageID == "" {
		log.Printf("[Notifications] Initial send for %s returned no message ID", channel)
		return
	}

	startedAt := info.StartedAt
	if startedAt.IsZero() {
		startedAt = time.Now()
	}
	record := NotificationRecord{
		Channel:    channel,
		Webhook:    n.DiscordWebhook,
		MessageID:  messageID,
		Message:    n.Message,
		StartedAt:  startedAt,
		LastLiveAt: time.Now(),
		Info:       info,
	}

	s.mu.Lock()
	s.records[key] = record
	s.saveLocked()
	s.mu.Unlock()

	log.Printf("[Notifications] %s went live, sent Discord notification", channel)
}

func (s *NotificationService) update(key string, record NotificationRecord, info StreamInfo) {
	editURL, err := webhookMessageURL(record.Webhook, record.MessageID)
	if err != nil {
		log.Printf("[Notifications] Invalid webhook for %s: %v", record.Channel, err)
		return
	}

	record.LastLiveAt = time.Now()
	record.Info = info
	if record.StartedAt.IsZero() && !info.StartedAt.IsZero() {
		record.StartedAt = info.StartedAt
	}

	if _, err := s.executeRequest(http.MethodPatch, editURL, buildLivePayload(info, true, record.Message)); err != nil {
		log.Printf("[Notifications] Uptime update failed for %s: %v", record.Channel, err)
		return
	}

	s.mu.Lock()
	s.records[key] = record
	s.saveLocked()
	s.mu.Unlock()

	logging.Debugf("[Notifications] Updated uptime for %s", record.Channel)
}

func (s *NotificationService) finish(key string, record NotificationRecord) {
	if record.MessageID != "" && record.Webhook != "" {
		editURL, err := webhookMessageURL(record.Webhook, record.MessageID)
		if err != nil {
			log.Printf("[Notifications] Invalid webhook for %s: %v", record.Channel, err)
		} else if _, err := s.executeRequest(http.MethodPatch, editURL, buildOfflinePayload(record)); err != nil {
			// Keep the record so the final edit is retried on the next poll.
			log.Printf("[Notifications] Failed to mark %s offline: %v", record.Channel, err)
			return
		}
	}

	s.mu.Lock()
	delete(s.records, key)
	s.saveLocked()
	s.mu.Unlock()

	log.Printf("[Notifications] %s is no longer live", record.Channel)
}

// prune removes records whose notification target was deleted from the
// config. There is nothing useful to edit in that case.
func (s *NotificationService) prune(configured map[string]struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	changed := false
	for key := range s.records {
		if _, ok := configured[key]; !ok {
			delete(s.records, key)
			changed = true
		}
	}
	if changed {
		s.saveLocked()
	}
}

func (s *NotificationService) saveLocked() {
	if s.store == nil {
		return
	}
	if err := s.store.Save(); err != nil {
		log.Println("[Notifications] Failed to save state:", err)
	}
}

func notificationKey(channel, webhook string) string {
	return strings.ToLower(strings.TrimSpace(channel)) + "|" + strings.TrimSpace(webhook)
}

func sameStream(a, b time.Time) bool {
	if a.IsZero() || b.IsZero() {
		return true
	}
	return a.Equal(b)
}

func buildLivePayload(info StreamInfo, isUpdate bool, message string) map[string]any {
	channelName := displayChannelName(info)
	fields := []map[string]any{
		{"name": "Game", "value": fieldValue(info.GameName), "inline": true},
		{"name": "Viewers", "value": fmt.Sprintf("%d", info.ViewerCount), "inline": true},
	}
	if isUpdate {
		if !info.StartedAt.IsZero() {
			fields = append(fields, map[string]any{
				"name": "Uptime", "value": formatDuration(time.Since(info.StartedAt)), "inline": true,
			})
		}
	}

	title := channelName + " is now live!"
	if isUpdate {
		title = channelName + " is live"
	}

	embed := map[string]any{
		"title":       title,
		"description": info.StreamTitle,
		"url":         "https://twitch.tv/" + strings.ToLower(channelName),
		"color":       0xE01B3C, // red
		"fields":      fields,
	}
	// Discord caches embed images by URL. Include a changing marker so the
	// preview reflects the current stream instead of the first frame.
	addThumbnail(embed, info.ThumbnailURL, time.Now().Unix())

	// Uptime updates edit the original message, so keeping the content
	// (including @everyone) does not ping anyone again.
	return map[string]any{
		"content": renderMessage(message, info),
		"embeds":  []map[string]any{embed},
	}
}

func buildOfflinePayload(record NotificationRecord) map[string]any {
	info := record.Info
	channelName := displayChannelName(info)
	if channelName == "" {
		channelName = record.Channel
	}

	lastLive := record.LastLiveAt
	if lastLive.IsZero() {
		lastLive = record.StartedAt
	}

	description := "The stream has ended."
	if !lastLive.IsZero() {
		description = fmt.Sprintf("Last live <t:%d:f> (<t:%d:R>).", lastLive.Unix(), lastLive.Unix())
	}

	fields := make([]map[string]any, 0, 3)
	if info.GameName != "" {
		fields = append(fields, map[string]any{"name": "Game", "value": info.GameName, "inline": true})
	}
	if !record.StartedAt.IsZero() && !lastLive.IsZero() && !lastLive.Before(record.StartedAt) {
		fields = append(fields, map[string]any{
			"name": "Streamed", "value": formatDuration(lastLive.Sub(record.StartedAt)), "inline": true,
		})
	}

	embed := map[string]any{
		"title":       channelName + " is no longer live",
		"description": description,
		"url":         "https://twitch.tv/" + strings.ToLower(channelName),
		"color":       0x95A5A6, // gray
		"fields":      fields,
	}
	cacheBust := record.LastLiveAt.Unix()
	if cacheBust <= 0 {
		cacheBust = time.Now().Unix()
	}
	addThumbnail(embed, info.ThumbnailURL, cacheBust)

	// This is an edit of the original message, so the content can stay as-is
	// without pinging anyone again.
	return map[string]any{
		"content": renderMessage(record.Message, info),
		"embeds":  []map[string]any{embed},
	}
}

func displayChannelName(info StreamInfo) string {
	return strings.TrimSpace(info.ChannelName)
}

func fieldValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "Unknown"
	}
	return value
}

func addThumbnail(embed map[string]any, thumbnailURL string, cacheBust int64) {
	if strings.TrimSpace(thumbnailURL) == "" {
		return
	}

	imageURL := strings.ReplaceAll(thumbnailURL, "{width}x{height}", "1280x720")
	if cacheBust > 0 {
		if u, err := url.Parse(imageURL); err == nil {
			q := u.Query()
			q.Set("t", strconv.FormatInt(cacheBust, 10))
			u.RawQuery = q.Encode()
			imageURL = u.String()
		}
	}

	embed["image"] = map[string]string{"url": imageURL}
}

func renderMessage(message string, info StreamInfo) string {
	if message == "" {
		message = "@everyone"
	}
	r := strings.NewReplacer(
		"{{channel}}", displayChannelName(info),
		"{{game}}", info.GameName,
		"{{title}}", info.StreamTitle,
	)
	return r.Replace(message)
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	d = d.Round(time.Minute)

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60

	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, minutes)
	default:
		return fmt.Sprintf("%dm", minutes)
	}
}

func webhookWithQuery(rawURL, key, value string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func webhookMessageURL(rawURL, messageID string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	u.Path = strings.TrimRight(u.Path, "/") + "/messages/" + url.PathEscape(messageID)
	return u.String(), nil
}

func (s *NotificationService) executeRequest(method, requestURL string, payload map[string]any) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(method, requestURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("discord API error: %s", resp.Status)
	}

	if method == http.MethodPost {
		var discordMsg struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&discordMsg); err != nil {
			return "", err
		}
		return discordMsg.ID, nil
	}

	return "", nil
}
