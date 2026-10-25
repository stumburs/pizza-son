package services

import (
	"encoding/json"
	"log"
	"maps"
	"os"
	"path/filepath"
	"pizza-son/internal/models"
	"pizza-son/internal/store"
	"strings"
	"sync"
	"time"
)

const (
	activityDir       = "data/activity"
	activityStatsFile = "stats.json"
	activityEventsDir = "events"

	activityDayFormat  = "2006-01-02"
	activityHourFormat = "2006-01-02 15:00"

	// Bump this when the meaning of persisted stats changes. Version 1 fixed
	// listener stats that used to count every listener invocation.
	activitySchemaVersion = 1
)

type ActivityKind string

const (
	ActivityMessage  ActivityKind = "message"
	ActivityCommand  ActivityKind = "command"
	ActivityListener ActivityKind = "listener"
)

// ActivityEvent is one append-only record in the event log. The event log is
// used for analytics only. Message content and role flags are intentionally
// not stored here: LoggerService already writes those to data/logs.
type ActivityEvent struct {
	Kind      ActivityKind    `json:"kind"`
	Timestamp time.Time       `json:"timestamp"`
	Channel   string          `json:"channel"`
	Platform  models.Platform `json:"platform"`
	UserID    string          `json:"user_id,omitempty"`
	Username  string          `json:"username,omitempty"`
	Target    string          `json:"target,omitempty"` // command or listener name
	IsCommand bool            `json:"is_command,omitempty"`
}

// ActivityMetric holds a total plus the same total broken down by local day
// and hour. Hour keys look like "2006-01-02 15:00".
type ActivityMetric struct {
	Total  int            `json:"total"`
	Daily  map[string]int `json:"daily"`
	Hourly map[string]int `json:"hourly"`
}

func (m *ActivityMetric) ensureMaps() {
	if m.Daily == nil {
		m.Daily = make(map[string]int)
	}
	if m.Hourly == nil {
		m.Hourly = make(map[string]int)
	}
}

func (m *ActivityMetric) add(ts time.Time) {
	m.Total++
	m.ensureMaps()
	m.Daily[ts.Format(activityDayFormat)]++
	m.Hourly[ts.Format(activityHourFormat)]++
}

func (m ActivityMetric) clone() ActivityMetric {
	return ActivityMetric{
		Total:  m.Total,
		Daily:  maps.Clone(m.Daily),
		Hourly: maps.Clone(m.Hourly),
	}
}

type ActivityUser struct {
	Messages            int       `json:"messages"`
	Commands            int       `json:"commands"`
	ListenerActivations int       `json:"listener_activations"`
	FirstSeen           time.Time `json:"first_seen"`
	LastSeen            time.Time `json:"last_seen"`
}

type ChannelActivity struct {
	Platform      models.Platform          `json:"platform"`
	Messages      ActivityMetric           `json:"messages"`
	Commands      ActivityMetric           `json:"commands"`
	Listeners     ActivityMetric           `json:"listeners"`
	CommandUsage  map[string]int           `json:"command_usage"`
	ListenerUsage map[string]int           `json:"listener_usage"`
	Users         map[string]*ActivityUser `json:"users"`
}

func (c *ChannelActivity) ensureMaps() {
	c.Messages.ensureMaps()
	c.Commands.ensureMaps()
	c.Listeners.ensureMaps()
	if c.CommandUsage == nil {
		c.CommandUsage = make(map[string]int)
	}
	if c.ListenerUsage == nil {
		c.ListenerUsage = make(map[string]int)
	}
	if c.Users == nil {
		c.Users = make(map[string]*ActivityUser)
	}
	for name, user := range c.Users {
		if user == nil {
			delete(c.Users, name)
		}
	}
}

func (c *ChannelActivity) user(name string) *ActivityUser {
	user, ok := c.Users[name]
	if !ok || user == nil {
		user = &ActivityUser{}
		c.Users[name] = user
	}
	return user
}

func (c *ChannelActivity) clone() ChannelActivity {
	users := make(map[string]*ActivityUser, len(c.Users))
	for name, user := range c.Users {
		if user == nil {
			continue
		}
		u := *user
		users[name] = &u
	}
	return ChannelActivity{
		Platform:      c.Platform,
		Messages:      c.Messages.clone(),
		Commands:      c.Commands.clone(),
		Listeners:     c.Listeners.clone(),
		CommandUsage:  maps.Clone(c.CommandUsage),
		ListenerUsage: maps.Clone(c.ListenerUsage),
		Users:         users,
	}
}

type ActivityData struct {
	Version  int                         `json:"version"`
	Channels map[string]*ChannelActivity `json:"channels"`
}

// GlobalActivityStats is the aggregate view used by the stats site.
type GlobalActivityStats struct {
	TotalMessages            int                               `json:"total_messages"`
	TotalCommands            int                               `json:"total_commands"`
	TotalListenerActivations int                               `json:"total_listener_activations"`
	UniqueUsers              int                               `json:"unique_users"`
	ChannelCount             int                               `json:"channel_count"`
	Channels                 map[string]ChannelActivitySummary `json:"channels"`
	Commands                 map[string]int                    `json:"commands"`
	Listeners                map[string]int                    `json:"listeners"`
	DailyMessages            map[string]int                    `json:"daily_messages"`
	HourlyMessages           map[string]int                    `json:"hourly_messages"`
	DailyCommands            map[string]int                    `json:"daily_commands"`
	HourlyCommands           map[string]int                    `json:"hourly_commands"`
	DailyListeners           map[string]int                    `json:"daily_listeners"`
	HourlyListeners          map[string]int                    `json:"hourly_listeners"`
}

type ChannelActivitySummary struct {
	Platform            models.Platform `json:"platform"`
	Messages            int             `json:"messages"`
	Commands            int             `json:"commands"`
	ListenerActivations int             `json:"listener_activations"`
	Users               int             `json:"users"`
}

// UserActivityStats is the per-user view split by channel.
type UserActivityStats struct {
	Username            string                   `json:"username"`
	Messages            int                      `json:"messages"`
	Commands            int                      `json:"commands"`
	ListenerActivations int                      `json:"listener_activations"`
	Channels            map[string]*ActivityUser `json:"channels"`
}

// ActivityService records every chat message, command invocation and listener
// activation. Events are appended to a daily JSONL file, while aggregate
// counters are kept in memory and periodically flushed to stats.json.
type ActivityService struct {
	mu        sync.RWMutex
	store     *store.Store[ActivityData]
	data      *ActivityData
	eventsDir string
	now       func() time.Time

	file    *os.File
	encoder *json.Encoder
	fileDay string

	stopCh    chan struct{}
	doneCh    chan struct{}
	closeOnce sync.Once
}

var ActivityServiceInstance *ActivityService

func NewActivityService() {
	svc := newActivityService(activityDir, 30*time.Second)
	ActivityServiceInstance = svc
	log.Println("[Activity] Service initialized")
}

func newActivityService(dir string, flushInterval time.Duration) *ActivityService {
	s := store.New(filepath.Join(dir, activityStatsFile), &ActivityData{
		Channels: make(map[string]*ChannelActivity),
	})
	svc := &ActivityService{
		store:     s,
		data:      s.Data(),
		eventsDir: filepath.Join(dir, activityEventsDir),
		now:       time.Now,
		stopCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
	if err := s.EnsureDir(); err != nil {
		log.Println("[Activity] Failed to create data directory:", err)
	}
	if err := os.MkdirAll(svc.eventsDir, 0755); err != nil {
		log.Println("[Activity] Failed to create events directory:", err)
	}
	svc.load()
	if flushInterval > 0 {
		go svc.flushLoop(flushInterval)
	} else {
		close(svc.doneCh)
	}
	return svc
}

func (s *ActivityService) load() {
	ok, err := s.store.LoadIfExists()
	if err != nil {
		log.Println("[Activity] Failed to load stats:", err)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data.Channels == nil {
		s.data.Channels = make(map[string]*ChannelActivity)
	}

	// Version 0 counted every listener invocation, including listeners that
	// did not handle the message. Drop those stats once. A missing file also
	// starts at version 0, so only migrate when a file was actually loaded.
	migrated := ok && s.data.Version < activitySchemaVersion
	for name, ch := range s.data.Channels {
		if ch == nil {
			delete(s.data.Channels, name)
			continue
		}
		if migrated {
			ch.Listeners = ActivityMetric{}
			ch.ListenerUsage = make(map[string]int)
			for _, user := range ch.Users {
				if user != nil {
					user.ListenerActivations = 0
				}
			}
		}
		ch.ensureMaps()
	}
	s.data.Version = activitySchemaVersion
	if migrated {
		log.Println("[Activity] Reset listener stats from the previous schema version")
		s.saveLocked()
	}
	if ok {
		log.Printf("[Activity] Loaded stats for %d channels", len(s.data.Channels))
	}
}

func (s *ActivityService) flushLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer close(s.doneCh)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			s.saveLocked()
			s.mu.Unlock()
		case <-s.stopCh:
			return
		}
	}
}

func (s *ActivityService) saveLocked() {
	if err := s.store.Save(); err != nil {
		log.Println("[Activity] Failed to save stats:", err)
	}
}

// LogMessage records an incoming chat message. isCommand reports whether the
// message was a command invocation; the message is still counted as chat.
func (s *ActivityService) LogMessage(msg models.Message, isCommand bool) {
	if s == nil {
		return
	}
	s.record(s.eventFromMessage(msg, ActivityMessage, "", isCommand))
}

// LogCommand records a command that passed all checks and is about to run.
func (s *ActivityService) LogCommand(msg models.Message, name string) {
	if s == nil {
		return
	}
	s.record(s.eventFromMessage(msg, ActivityCommand, name, false))
}

// LogListener records a listener that ran and returned true, meaning it
// actually handled the message rather than just being checked.
func (s *ActivityService) LogListener(msg models.Message, name string) {
	if s == nil {
		return
	}
	s.record(s.eventFromMessage(msg, ActivityListener, name, false))
}

func (s *ActivityService) eventFromMessage(msg models.Message, kind ActivityKind, target string, isCommand bool) ActivityEvent {
	return ActivityEvent{
		Kind:      kind,
		Timestamp: s.now(),
		Channel:   msg.Channel,
		Platform:  msg.Platform,
		UserID:    msg.User.ID,
		Username:  msg.User.Name,
		Target:    target,
		IsCommand: isCommand,
	}
}

func (s *ActivityService) record(ev ActivityEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.writeEventLocked(ev)

	channelKey := strings.ToLower(ev.Channel)
	ch, ok := s.data.Channels[channelKey]
	if !ok || ch == nil {
		ch = &ChannelActivity{
			Platform:      ev.Platform,
			CommandUsage:  make(map[string]int),
			ListenerUsage: make(map[string]int),
			Users:         make(map[string]*ActivityUser),
		}
		s.data.Channels[channelKey] = ch
	}
	if ch.Platform == "" {
		ch.Platform = ev.Platform
	}
	ch.ensureMaps()

	switch ev.Kind {
	case ActivityMessage:
		ch.Messages.add(ev.Timestamp)
	case ActivityCommand:
		ch.Commands.add(ev.Timestamp)
		if ev.Target != "" {
			ch.CommandUsage[strings.ToLower(ev.Target)]++
		}
	case ActivityListener:
		ch.Listeners.add(ev.Timestamp)
		if ev.Target != "" {
			ch.ListenerUsage[strings.ToLower(ev.Target)]++
		}
	}

	username := strings.ToLower(ev.Username)
	if username == "" {
		return
	}
	user := ch.user(username)
	if user.FirstSeen.IsZero() || ev.Timestamp.Before(user.FirstSeen) {
		user.FirstSeen = ev.Timestamp
	}
	if ev.Timestamp.After(user.LastSeen) {
		user.LastSeen = ev.Timestamp
	}
	switch ev.Kind {
	case ActivityMessage:
		user.Messages++
	case ActivityCommand:
		user.Commands++
	case ActivityListener:
		user.ListenerActivations++
	}
}

func (s *ActivityService) writeEventLocked(ev ActivityEvent) {
	day := ev.Timestamp.Format(activityDayFormat)
	if s.encoder == nil || s.fileDay != day {
		if s.file != nil {
			if err := s.file.Close(); err != nil {
				log.Println("[Activity] Failed to close event log:", err)
			}
			s.file = nil
			s.encoder = nil
		}

		path := filepath.Join(s.eventsDir, "events-"+day+".jsonl")
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("[Activity] Failed to open event log for %s: %v", day, err)
			return
		}
		s.file = f
		s.encoder = json.NewEncoder(f)
		s.fileDay = day
	}

	if err := s.encoder.Encode(ev); err != nil {
		log.Println("[Activity] Failed to write event:", err)
	}
}

// GlobalStats returns totals for every channel. Discord channels are included;
// consumers can filter by platform.
func (s *ActivityService) GlobalStats() GlobalActivityStats {
	stats := GlobalActivityStats{
		Channels:        make(map[string]ChannelActivitySummary),
		Commands:        make(map[string]int),
		Listeners:       make(map[string]int),
		DailyMessages:   make(map[string]int),
		HourlyMessages:  make(map[string]int),
		DailyCommands:   make(map[string]int),
		HourlyCommands:  make(map[string]int),
		DailyListeners:  make(map[string]int),
		HourlyListeners: make(map[string]int),
	}
	if s == nil || s.data == nil {
		return stats
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	seenUsers := make(map[string]struct{})
	for name, ch := range s.data.Channels {
		if ch == nil {
			continue
		}
		stats.ChannelCount++
		stats.TotalMessages += ch.Messages.Total
		stats.TotalCommands += ch.Commands.Total
		stats.TotalListenerActivations += ch.Listeners.Total

		stats.Channels[name] = ChannelActivitySummary{
			Platform:            ch.Platform,
			Messages:            ch.Messages.Total,
			Commands:            ch.Commands.Total,
			ListenerActivations: ch.Listeners.Total,
			Users:               len(ch.Users),
		}

		addCounts(stats.Commands, ch.CommandUsage)
		addCounts(stats.Listeners, ch.ListenerUsage)
		addCounts(stats.DailyMessages, ch.Messages.Daily)
		addCounts(stats.HourlyMessages, ch.Messages.Hourly)
		addCounts(stats.DailyCommands, ch.Commands.Daily)
		addCounts(stats.HourlyCommands, ch.Commands.Hourly)
		addCounts(stats.DailyListeners, ch.Listeners.Daily)
		addCounts(stats.HourlyListeners, ch.Listeners.Hourly)

		for username := range ch.Users {
			seenUsers[username] = struct{}{}
		}
	}
	stats.UniqueUsers = len(seenUsers)
	return stats
}

// ChannelStats returns a copy of one channel's aggregate data.
func (s *ActivityService) ChannelStats(channel string) (ChannelActivity, bool) {
	if s == nil || s.data == nil {
		return ChannelActivity{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	ch, ok := s.data.Channels[strings.ToLower(channel)]
	if !ok || ch == nil {
		return ChannelActivity{}, false
	}
	return ch.clone(), true
}

// UserStats returns one user's activity across every channel.
func (s *ActivityService) UserStats(username string) (UserActivityStats, bool) {
	stats := UserActivityStats{
		Username: strings.ToLower(username),
		Channels: make(map[string]*ActivityUser),
	}
	if s == nil || s.data == nil {
		return stats, false
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for name, ch := range s.data.Channels {
		if ch == nil {
			continue
		}
		user, ok := ch.Users[stats.Username]
		if !ok || user == nil {
			continue
		}
		copied := *user
		stats.Channels[name] = &copied
		stats.Messages += copied.Messages
		stats.Commands += copied.Commands
		stats.ListenerActivations += copied.ListenerActivations
	}
	return stats, len(stats.Channels) > 0
}

func addCounts(dst, src map[string]int) {
	for key, count := range src {
		dst[key] += count
	}
}

// Close flushes the final aggregate state and closes the current event log.
func (s *ActivityService) Close() {
	if s == nil {
		return
	}
	s.closeOnce.Do(func() {
		if s.stopCh != nil {
			close(s.stopCh)
			<-s.doneCh
		}

		s.mu.Lock()
		defer s.mu.Unlock()

		if s.file != nil {
			if err := s.file.Close(); err != nil {
				log.Println("[Activity] Failed to close event log:", err)
			}
			s.file = nil
			s.encoder = nil
		}
		s.saveLocked()
	})
}
