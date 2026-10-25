package services

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"pizza-son/internal/models"
	"strings"
	"testing"
	"time"
)

func testActivityMessage(channel, name, text string) models.Message {
	return models.Message{
		Channel:  channel,
		Platform: models.PlatformTwitch,
		Text:     text,
		User: models.MessageUser{
			ID:   name + "-id",
			Name: name,
		},
	}
}

func TestActivityServiceRecordsAndAggregates(t *testing.T) {
	dir := t.TempDir()
	svc := newActivityService(dir, 0)
	defer svc.Close()

	at := func(hour, minute int) time.Time {
		return time.Date(2026, 10, 25, hour, minute, 0, 0, time.UTC)
	}

	now := at(13, 5)
	svc.now = func() time.Time { return now }

	alice := testActivityMessage("Pizza_TM", "Alice", "hello")
	svc.LogMessage(alice, false)

	now = at(13, 10)
	svc.LogCommand(alice, "bertcheck")

	now = at(14, 0)
	svc.LogListener(alice, "bert")

	now = at(14, 5)
	bob := testActivityMessage("pizza_tm", "Bob", "hi")
	svc.LogMessage(bob, true)

	channel, ok := svc.ChannelStats("pizza_tm")
	if !ok {
		t.Fatal("channel stats not found")
	}
	if channel.Platform != models.PlatformTwitch {
		t.Errorf("platform = %q, want %q", channel.Platform, models.PlatformTwitch)
	}
	if channel.Messages.Total != 2 {
		t.Errorf("messages = %d, want 2", channel.Messages.Total)
	}
	if channel.Messages.Daily["2026-10-25"] != 2 {
		t.Errorf("daily messages = %d, want 2", channel.Messages.Daily["2026-10-25"])
	}
	if channel.Messages.Hourly["2026-10-25 13:00"] != 1 || channel.Messages.Hourly["2026-10-25 14:00"] != 1 {
		t.Errorf("hourly messages = %#v, want one at 13:00 and one at 14:00", channel.Messages.Hourly)
	}
	if channel.Commands.Total != 1 || channel.Commands.Daily["2026-10-25"] != 1 {
		t.Errorf("commands = %#v, want total 1 for the day", channel.Commands)
	}
	if channel.CommandUsage["bertcheck"] != 1 {
		t.Errorf("command usage = %#v, want bertcheck=1", channel.CommandUsage)
	}
	if channel.Listeners.Total != 1 || channel.ListenerUsage["bert"] != 1 {
		t.Errorf("listeners = %#v, usage = %#v, want bert=1", channel.Listeners, channel.ListenerUsage)
	}
	if len(channel.Users) != 2 {
		t.Fatalf("users = %d, want 2", len(channel.Users))
	}

	aliceStats := channel.Users["alice"]
	if aliceStats == nil {
		t.Fatal("alice stats not found")
	}
	if aliceStats.Messages != 1 || aliceStats.Commands != 1 || aliceStats.ListenerActivations != 1 {
		t.Errorf("alice stats = %#v", aliceStats)
	}
	if !aliceStats.FirstSeen.Equal(at(13, 5)) {
		t.Errorf("alice first seen = %v, want %v", aliceStats.FirstSeen, at(13, 5))
	}
	if !aliceStats.LastSeen.Equal(at(14, 0)) {
		t.Errorf("alice last seen = %v, want %v", aliceStats.LastSeen, at(14, 0))
	}

	user, ok := svc.UserStats("ALICE")
	if !ok {
		t.Fatal("user stats not found")
	}
	if user.Messages != 1 || user.Commands != 1 || user.ListenerActivations != 1 {
		t.Errorf("user stats = %#v", user)
	}
	if _, ok := user.Channels["pizza_tm"]; !ok {
		t.Errorf("user channels = %#v, want pizza_tm", user.Channels)
	}

	global := svc.GlobalStats()
	if global.TotalMessages != 2 || global.TotalCommands != 1 || global.TotalListenerActivations != 1 {
		t.Errorf("global stats = %#v", global)
	}
	if global.UniqueUsers != 2 {
		t.Errorf("unique users = %d, want 2", global.UniqueUsers)
	}
	if global.ChannelCount != 1 {
		t.Errorf("channel count = %d, want 1", global.ChannelCount)
	}
	if global.Channels["pizza_tm"].Messages != 2 {
		t.Errorf("global channel summary = %#v", global.Channels["pizza_tm"])
	}

	// close flushes aggregates and the daily event log
	svc.Close()

	eventPath := filepath.Join(dir, activityEventsDir, "events-2026-10-25.jsonl")
	file, err := os.Open(eventPath)
	if err != nil {
		t.Fatalf("open event log: %v", err)
	}
	defer file.Close()

	var events []ActivityEvent
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Message content belongs to the existing logger, not the activity log.
		if strings.Contains(line, `"message":`) || strings.Contains(line, `"display_name":`) {
			t.Errorf("activity event duplicates message content: %s", line)
		}
		var ev ActivityEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatalf("decode event: %v", err)
		}
		events = append(events, ev)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("read event log: %v", err)
	}
	if len(events) != 4 {
		t.Fatalf("events = %d, want 4", len(events))
	}
	if events[1].Kind != ActivityCommand || events[1].Target != "bertcheck" {
		t.Errorf("second event = %#v, want bertcheck command", events[1])
	}
	if events[2].Kind != ActivityListener || events[2].Target != "bert" {
		t.Errorf("third event = %#v, want bert listener", events[2])
	}
	if !events[3].IsCommand {
		t.Error("fourth event should be flagged as a command message")
	}

	// aggregates survive a restart
	reloaded := newActivityService(dir, 0)
	defer reloaded.Close()

	channel, ok = reloaded.ChannelStats("pizza_tm")
	if !ok {
		t.Fatal("reloaded channel stats not found")
	}
	if channel.Messages.Total != 2 || channel.Commands.Total != 1 || channel.Listeners.Total != 1 {
		t.Errorf("reloaded channel stats = %#v", channel)
	}
	if channel.CommandUsage["bertcheck"] != 1 || channel.ListenerUsage["bert"] != 1 {
		t.Errorf("reloaded usage = %#v / %#v", channel.CommandUsage, channel.ListenerUsage)
	}
}

func TestActivityServiceMigratesListenerStats(t *testing.T) {
	dir := t.TempDir()

	legacy := struct {
		Channels map[string]*ChannelActivity `json:"channels"`
	}{
		Channels: map[string]*ChannelActivity{
			"testchannel": {
				Platform:  models.PlatformTwitch,
				Messages:  ActivityMetric{Total: 5},
				Listeners: ActivityMetric{Total: 27},
				ListenerUsage: map[string]int{
					"noop": 27,
				},
				Users: map[string]*ActivityUser{
					"alice": {
						Messages:            5,
						ListenerActivations: 27,
					},
				},
			},
		},
	}
	raw, err := json.Marshal(legacy)
	if err != nil {
		t.Fatalf("marshal old stats: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, activityStatsFile), raw, 0644); err != nil {
		t.Fatalf("write old stats: %v", err)
	}

	svc := newActivityService(dir, 0)
	defer svc.Close()

	channel, ok := svc.ChannelStats("testchannel")
	if !ok {
		t.Fatal("channel stats not found")
	}
	if channel.Messages.Total != 5 {
		t.Errorf("messages = %d, want 5 (message stats should be retained)", channel.Messages.Total)
	}
	if channel.Listeners.Total != 0 {
		t.Errorf("listener total = %d, want 0 after migration", channel.Listeners.Total)
	}
	if len(channel.ListenerUsage) != 0 {
		t.Errorf("listener usage = %#v, want empty after migration", channel.ListenerUsage)
	}
	if channel.Users["alice"].ListenerActivations != 0 {
		t.Errorf("user listener activations = %d, want 0 after migration", channel.Users["alice"].ListenerActivations)
	}
	if channel.Users["alice"].Messages != 5 {
		t.Errorf("user messages = %d, want 5 after migration", channel.Users["alice"].Messages)
	}

	// The migrated version is written back so the reset only happens once.
	raw, err = os.ReadFile(filepath.Join(dir, activityStatsFile))
	if err != nil {
		t.Fatalf("read migrated stats: %v", err)
	}
	var migrated ActivityData
	if err := json.Unmarshal(raw, &migrated); err != nil {
		t.Fatalf("decode migrated stats: %v", err)
	}
	if migrated.Version != activitySchemaVersion {
		t.Errorf("version = %d, want %d", migrated.Version, activitySchemaVersion)
	}
}

func TestActivityMetricAdd(t *testing.T) {
	var metric ActivityMetric
	ts := time.Date(2026, 10, 25, 23, 59, 0, 0, time.UTC)
	metric.add(ts)
	metric.add(ts)

	if metric.Total != 2 {
		t.Errorf("total = %d, want 2", metric.Total)
	}
	if metric.Daily["2026-10-25"] != 2 {
		t.Errorf("daily = %#v", metric.Daily)
	}
	if metric.Hourly["2026-10-25 23:00"] != 2 {
		t.Errorf("hourly = %#v", metric.Hourly)
	}
}
