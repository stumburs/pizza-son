package services

import (
	"log"
	"path/filepath"
	"pizza-son/internal/config"
	"pizza-son/internal/store"
	"strings"
	"sync"
	"time"
)

const (
	streamStatsDir          = "data/streams"
	streamStatsFile         = "stats.json"
	streamStatsPollInterval = time.Minute
)

// StreamSession is a stream that is currently being observed. LastSeen keeps
// the session safe across restarts: if the bot goes down mid-stream, the gap
// is only credited when the same stream is seen live again.
type StreamSession struct {
	StartedAt time.Time `json:"started_at"`
	LastSeen  time.Time `json:"last_seen"`
}

type StreamChannelStats struct {
	Seconds int64 `json:"seconds"`
}

// StreamRecord is one completed stream. EndedAt is the last poll where the
// channel was seen live, and Seconds is the time credited to the totals. The
// history is kept so later features can chart months, compare lengths, etc.
type StreamRecord struct {
	Channel   string    `json:"channel"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
	Seconds   int64     `json:"seconds"`
}

type StreamStatsData struct {
	Channels map[string]*StreamChannelStats `json:"channels"`
	Active   map[string]*StreamSession      `json:"active"`
	Streams  []StreamRecord                 `json:"streams"`
}

// StreamStatsSnapshot is the read-only view used by the public site.
type StreamStatsSnapshot struct {
	TotalSeconds   int64            `json:"total_seconds"`
	ChannelSeconds map[string]int64 `json:"channel_seconds"`
}

// StreamStatsService records how long configured channels have been live. It
// polls Twitch once a minute, closes a session when a channel goes offline,
// records the completed stream and persists totals and history between restarts.
type StreamStatsService struct {
	mu    sync.RWMutex
	store *store.Store[StreamStatsData]
	data  *StreamStatsData
}

var StreamStatsInstance *StreamStatsService

func NewStreamStatsService() {
	StreamStatsInstance = newStreamStatsService(streamStatsDir)
	go StreamStatsInstance.run()
	log.Println("[Streams] Service initialized")
}

func newStreamStatsService(dir string) *StreamStatsService {
	st := store.New(filepath.Join(dir, streamStatsFile), &StreamStatsData{
		Channels: make(map[string]*StreamChannelStats),
		Active:   make(map[string]*StreamSession),
		Streams:  []StreamRecord{},
	})
	s := &StreamStatsService{store: st, data: st.Data()}

	if err := st.EnsureDir(); err != nil {
		log.Println("[Streams] Failed to create data directory:", err)
	}
	ok, err := st.LoadIfExists()
	if err != nil {
		log.Println("[Streams] Failed to load stats:", err)
	}
	s.normalize()
	if ok {
		log.Printf("[Streams] Loaded stream totals for %d channels and %d streams", len(s.data.Channels), len(s.data.Streams))
	}
	return s
}

func (s *StreamStatsService) normalize() {
	if s.data.Channels == nil {
		s.data.Channels = make(map[string]*StreamChannelStats)
	}
	if s.data.Active == nil {
		s.data.Active = make(map[string]*StreamSession)
	}
	if s.data.Streams == nil {
		s.data.Streams = []StreamRecord{}
	}
	for name, stats := range s.data.Channels {
		if stats == nil || name == "" {
			delete(s.data.Channels, name)
		}
	}
	for name, session := range s.data.Active {
		if session == nil || name == "" || session.StartedAt.IsZero() {
			delete(s.data.Active, name)
		}
	}
}

func (s *StreamStatsService) run() {
	s.poll()
	ticker := time.NewTicker(streamStatsPollInterval)
	defer ticker.Stop()
	for range ticker.C {
		s.poll()
	}
}

func (s *StreamStatsService) poll() {
	cfg := config.Current()
	if cfg == nil || TwitchServiceInstance == nil {
		return
	}
	channels := cfg.Bot.Channels
	live, err := TwitchServiceInstance.GetLiveStreams(channels)
	if err != nil {
		// An API failure is not proof that every stream ended. Leaving the
		// active sessions alone prevents one bad poll from closing them and
		// losing stream time.
		log.Printf("[Streams] Skipping update, Twitch lookup failed: %v", err)
		return
	}
	s.Track(live, channels, time.Now())
}

// Track updates the accumulators from one live snapshot. channels is the set
// that was polled, so sessions for removed channels can be closed instead of
// staying active forever.
func (s *StreamStatsService) Track(live map[string]StreamInfo, channels []string, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.trackLocked(live, channels, now)

	if err := s.store.Save(); err != nil {
		log.Println("[Streams] Failed to save stats:", err)
	}
}

func (s *StreamStatsService) trackLocked(live map[string]StreamInfo, channels []string, now time.Time) {
	polled := make(map[string]struct{}, len(channels)+len(live))
	liveByName := make(map[string]StreamInfo, len(live))
	for _, name := range channels {
		name = strings.ToLower(strings.TrimSpace(name))
		if name != "" {
			polled[name] = struct{}{}
		}
	}
	for name, info := range live {
		name = strings.ToLower(strings.TrimSpace(name))
		if name == "" {
			continue
		}
		polled[name] = struct{}{}
		liveByName[name] = info
	}

	for name, session := range s.data.Active {
		if _, ok := polled[name]; !ok {
			s.finishLocked(name, session)
		}
	}

	for name := range polled {
		info, isLive := liveByName[name]
		session := s.data.Active[name]

		if !isLive {
			if session != nil {
				s.finishLocked(name, session)
			}
			continue
		}

		started := info.StartedAt
		if started.IsZero() || started.After(now) {
			started = now
		}

		if session == nil {
			s.data.Active[name] = &StreamSession{StartedAt: started, LastSeen: now}
			continue
		}
		if !session.StartedAt.Equal(started) {
			s.finishLocked(name, session)
			s.data.Active[name] = &StreamSession{StartedAt: started, LastSeen: now}
			continue
		}
		session.LastSeen = now
	}
}

func (s *StreamStatsService) finishLocked(name string, session *StreamSession) {
	delete(s.data.Active, name)
	if session == nil || session.LastSeen.IsZero() || session.LastSeen.Before(session.StartedAt) {
		return
	}

	seconds := int64(session.LastSeen.Sub(session.StartedAt).Seconds())

	stats := s.data.Channels[name]
	if stats == nil {
		stats = &StreamChannelStats{}
		s.data.Channels[name] = stats
	}
	stats.Seconds += seconds

	// Ignore zero-length sessions so the history only contains streams that
	// actually contributed time.
	if seconds > 0 {
		s.data.Streams = append(s.data.Streams, StreamRecord{
			Channel:   name,
			StartedAt: session.StartedAt,
			EndedAt:   session.LastSeen,
			Seconds:   seconds,
		})
	}
}

// History returns a copy of every completed stream. It is intended for
// future reporting such as monthly averages, graphs, and longest streams.
func (s *StreamStatsService) History() []StreamRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history := make([]StreamRecord, len(s.data.Streams))
	copy(history, s.data.Streams)
	return history
}

// Snapshot returns completed time plus the elapsed part of every active
// session. Active sessions use LastSeen instead of the current time so a
// tracker outage cannot balloon the total.
func (s *StreamStatsService) Snapshot() StreamStatsSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot := StreamStatsSnapshot{
		ChannelSeconds: make(map[string]int64, len(s.data.Channels)+len(s.data.Active)),
	}
	for name, stats := range s.data.Channels {
		if stats == nil || stats.Seconds <= 0 {
			continue
		}
		snapshot.ChannelSeconds[name] += stats.Seconds
		snapshot.TotalSeconds += stats.Seconds
	}
	for name, session := range s.data.Active {
		if session == nil || session.LastSeen.IsZero() {
			continue
		}
		seconds := int64(session.LastSeen.Sub(session.StartedAt).Seconds())
		if seconds <= 0 {
			continue
		}
		snapshot.ChannelSeconds[name] += seconds
		snapshot.TotalSeconds += seconds
	}
	return snapshot
}
