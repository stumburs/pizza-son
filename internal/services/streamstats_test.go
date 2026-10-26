package services

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStreamStatsTracksLiveAndOffline(t *testing.T) {
	now := time.Date(2026, 10, 25, 20, 0, 0, 0, time.UTC)
	started := now.Add(-2 * time.Hour)

	svc := newStreamStatsService(t.TempDir())
	channels := []string{"alpha", "beta"}

	svc.Track(map[string]StreamInfo{
		"alpha": {ChannelName: "alpha", StartedAt: started},
	}, channels, now)

	snapshot := svc.Snapshot()
	if snapshot.TotalSeconds != 7200 {
		t.Fatalf("total = %d, want 7200", snapshot.TotalSeconds)
	}
	if snapshot.ChannelSeconds["alpha"] != 7200 {
		t.Errorf("alpha = %d, want 7200", snapshot.ChannelSeconds["alpha"])
	}
	if snapshot.ChannelSeconds["beta"] != 0 {
		t.Errorf("beta = %d, want 0", snapshot.ChannelSeconds["beta"])
	}

	// One more poll while the same stream is live extends the session.
	svc.Track(map[string]StreamInfo{
		"alpha": {ChannelName: "alpha", StartedAt: started},
	}, channels, now.Add(time.Minute))

	if got := svc.Snapshot().TotalSeconds; got != 7260 {
		t.Fatalf("total after second poll = %d, want 7260", got)
	}

	// Going offline closes the session at the last time it was observed.
	svc.Track(nil, channels, now.Add(2*time.Minute))

	snapshot = svc.Snapshot()
	if snapshot.TotalSeconds != 7260 {
		t.Errorf("total after offline = %d, want 7260", snapshot.TotalSeconds)
	}
	if len(svc.data.Active) != 0 {
		t.Errorf("active sessions = %d, want 0", len(svc.data.Active))
	}

	if len(svc.data.Streams) != 1 {
		t.Fatalf("history records = %d, want 1", len(svc.data.Streams))
	}
	record := svc.data.Streams[0]
	if record.Channel != "alpha" {
		t.Errorf("history channel = %q, want alpha", record.Channel)
	}
	if !record.StartedAt.Equal(started) {
		t.Errorf("history start = %v, want %v", record.StartedAt, started)
	}
	if !record.EndedAt.Equal(now.Add(time.Minute)) {
		t.Errorf("history end = %v, want %v", record.EndedAt, now.Add(time.Minute))
	}
	if record.Seconds != 7260 {
		t.Errorf("history seconds = %d, want 7260", record.Seconds)
	}

	history := svc.History()
	if len(history) != 1 {
		t.Fatalf("History() returned %d records, want 1", len(history))
	}
	history[0].Channel = "mutated"
	if svc.data.Streams[0].Channel != "alpha" {
		t.Error("History() should return a copy")
	}
}

func TestStreamStatsNewStreamReplacesOldSession(t *testing.T) {
	now := time.Date(2026, 10, 25, 20, 0, 0, 0, time.UTC)
	firstStart := now.Add(-time.Hour)

	svc := newStreamStatsService(t.TempDir())
	channels := []string{"alpha"}

	svc.Track(map[string]StreamInfo{
		"alpha": {StartedAt: firstStart},
	}, channels, now)

	// Same channel, but Twitch reports a different start time: the old
	// session ends at the previous poll and a new one begins.
	secondStart := now.Add(time.Minute)
	svc.Track(map[string]StreamInfo{
		"alpha": {StartedAt: secondStart},
	}, channels, now.Add(10*time.Minute))

	snapshot := svc.Snapshot()
	want := int64(3600 + 9*60) // old session 1h, new session 9m
	if snapshot.TotalSeconds != want {
		t.Errorf("total = %d, want %d", snapshot.TotalSeconds, want)
	}

	if len(svc.data.Streams) != 1 {
		t.Fatalf("history records = %d, want 1", len(svc.data.Streams))
	}
	if got := svc.data.Streams[0].Seconds; got != 3600 {
		t.Errorf("finished stream seconds = %d, want 3600", got)
	}
}

func TestStreamStatsClosesRemovedChannels(t *testing.T) {
	now := time.Date(2026, 10, 25, 20, 0, 0, 0, time.UTC)
	started := now.Add(-time.Hour)

	svc := newStreamStatsService(t.TempDir())

	svc.Track(map[string]StreamInfo{
		"alpha": {StartedAt: started},
	}, []string{"alpha", "beta"}, now)

	// alpha is no longer polled, so its session is closed at the last seen
	// timestamp instead of staying active forever.
	svc.Track(nil, []string{"beta"}, now.Add(time.Minute))

	snapshot := svc.Snapshot()
	if snapshot.TotalSeconds != 3600 {
		t.Errorf("total = %d, want 3600", snapshot.TotalSeconds)
	}
	if len(svc.data.Active) != 0 {
		t.Errorf("active sessions = %d, want 0", len(svc.data.Active))
	}
	if len(svc.data.Streams) != 1 || svc.data.Streams[0].Channel != "alpha" {
		t.Errorf("history = %#v, want one alpha record", svc.data.Streams)
	}
}

func TestStreamStatsPersistsAcrossRestart(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 10, 25, 20, 0, 0, 0, time.UTC)
	started := now.Add(-90 * time.Minute)

	svc := newStreamStatsService(dir)
	svc.Track(map[string]StreamInfo{
		"alpha": {StartedAt: started},
	}, []string{"alpha"}, now)

	reloaded := newStreamStatsService(dir)
	snapshot := reloaded.Snapshot()
	if snapshot.TotalSeconds != 5400 {
		t.Fatalf("total after reload = %d, want 5400", snapshot.TotalSeconds)
	}
	if reloaded.data.Active["alpha"] == nil {
		t.Error("active session was not restored")
	}
	if len(reloaded.data.Streams) != 0 {
		t.Errorf("history before stream end = %d, want 0", len(reloaded.data.Streams))
	}

	// A poll after the restart continues the same session instead of
	// double-counting it.
	reloaded.Track(map[string]StreamInfo{
		"alpha": {StartedAt: started},
	}, []string{"alpha"}, now.Add(time.Minute))

	if got := reloaded.Snapshot().TotalSeconds; got != 5460 {
		t.Errorf("total after restart poll = %d, want 5460", got)
	}

	// Ending the stream writes exactly one history record, and a second
	// restart must not duplicate it.
	reloaded.Track(nil, []string{"alpha"}, now.Add(2*time.Minute))
	if len(reloaded.data.Streams) != 1 {
		t.Fatalf("history records after end = %d, want 1", len(reloaded.data.Streams))
	}

	reloadedAgain := newStreamStatsService(dir)
	if len(reloadedAgain.data.Streams) != 1 {
		t.Fatalf("history records after second reload = %d, want 1", len(reloadedAgain.data.Streams))
	}
	if got := reloadedAgain.data.Streams[0].Seconds; got != 5460 {
		t.Errorf("persisted history seconds = %d, want 5460", got)
	}
	if got := reloadedAgain.Snapshot().TotalSeconds; got != 5460 {
		t.Errorf("total after second reload = %d, want 5460", got)
	}
}

func TestStreamStatsLoadsLegacyDataWithoutHistory(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	legacy := `{"channels":{"alpha":{"seconds":3600}},"active":{}}`
	if err := os.WriteFile(filepath.Join(dir, streamStatsFile), []byte(legacy), 0644); err != nil {
		t.Fatal(err)
	}

	svc := newStreamStatsService(dir)
	if svc.data.Streams == nil {
		t.Fatal("history should be initialized for legacy data")
	}
	if len(svc.data.Streams) != 0 {
		t.Errorf("legacy history = %d, want 0", len(svc.data.Streams))
	}
	if got := svc.Snapshot().TotalSeconds; got != 3600 {
		t.Errorf("legacy total = %d, want 3600", got)
	}
}

func TestStreamStatsIgnoresInvalidSession(t *testing.T) {
	svc := newStreamStatsService(t.TempDir())
	now := time.Date(2026, 10, 25, 20, 0, 0, 0, time.UTC)

	svc.mu.Lock()
	svc.data.Active["glitch"] = &StreamSession{
		StartedAt: now,
		LastSeen:  now.Add(-time.Minute),
	}
	svc.mu.Unlock()

	if got := svc.Snapshot().TotalSeconds; got != 0 {
		t.Errorf("total = %d, want 0", got)
	}
	if len(svc.data.Streams) != 0 {
		t.Errorf("invalid session created history: %#v", svc.data.Streams)
	}
}
