package web

import (
	"net/http"
	"net/http/httptest"
	"pizza-son/internal/config"
	"pizza-son/internal/models"
	"pizza-son/internal/services"
	"testing"
	"time"
)

type fakeActivity struct {
	global   services.GlobalActivityStats
	channels map[string]services.ChannelActivity
}

func (f fakeActivity) GlobalStats() services.GlobalActivityStats {
	return f.global
}

func (f fakeActivity) ChannelStats(channel string) (services.ChannelActivity, bool) {
	stats, ok := f.channels[channel]
	return stats, ok
}

func TestBuildSiteSummary(t *testing.T) {
	now := time.Date(2026, 10, 25, 20, 0, 0, 0, time.UTC)
	started := now.Add(-90 * time.Minute)

	cfg := &config.Config{}
	cfg.Bot.Channels = []string{"alpha", "beta"}
	cfg.Bot.Prefix = "!"
	cfg.Twitch.User = "pizza-son"
	cfg.Discord.Channels = []string{"123456789012345678"}

	activity := fakeActivity{
		global: services.GlobalActivityStats{
			Channels: map[string]services.ChannelActivitySummary{
				"alpha":              {Platform: models.PlatformTwitch},
				"beta":               {Platform: models.PlatformTwitch},
				"123456789012345678": {Platform: models.PlatformDiscord},
			},
		},
		channels: map[string]services.ChannelActivity{
			"alpha": {
				Platform:  models.PlatformTwitch,
				Messages:  services.ActivityMetric{Total: 10},
				Commands:  services.ActivityMetric{Total: 3},
				Listeners: services.ActivityMetric{Total: 2},
				Users: map[string]*services.ActivityUser{
					"alice": {},
					"bob":   {},
				},
				CommandUsage:  map[string]int{"hello": 2, "help": 1},
				ListenerUsage: map[string]int{"bert": 2},
			},
			"beta": {
				Platform: models.PlatformTwitch,
				Messages: services.ActivityMetric{Total: 5},
				Users: map[string]*services.ActivityUser{
					"carol": {},
				},
				CommandUsage:  map[string]int{"hello": 1},
				ListenerUsage: map[string]int{"bert": 1},
			},
		},
	}

	berts := &services.BertService{
		Data: map[string]*services.ChannelData{
			"alpha": {
				UserStats: map[string]services.UserStats{
					"alice": {
						TotalActivations: 4,
						BertRecords: map[string]services.BertRecord{
							"beert": {Count: 3},
						},
					},
				},
			},
			"beta": {
				UserStats: map[string]services.UserStats{
					"carol": {
						TotalActivations: 2,
						BertRecords: map[string]services.BertRecord{
							"beert": {Count: 1},
						},
					},
				},
			},
			"123456789012345678": {
				UserStats: map[string]services.UserStats{
					"alice": {TotalActivations: 99},
				},
			},
		},
	}

	live := map[string]services.StreamInfo{
		"alpha": {
			ChannelName:  "alpha",
			StreamTitle:  "painting the ceiling",
			GameName:     "Art",
			ViewerCount:  42,
			ThumbnailURL: "https://example.test/{width}x{height}.jpg",
			StartedAt:    started,
		},
	}

	streams := services.StreamStatsSnapshot{
		TotalSeconds:   9000,
		ChannelSeconds: map[string]int64{"alpha": 5400, "beta": 3600},
	}

	summary := buildSiteSummary(now, cfg, activity, berts, live, streams)

	if summary.Bot.Name != "pizza-son" || summary.Bot.TwitchChannels != 2 || summary.Bot.DiscordChannels != 1 {
		t.Errorf("bot info = %#v", summary.Bot)
	}
	if summary.Totals.Channels != 2 {
		t.Errorf("channels = %d, want 2 (discord excluded)", summary.Totals.Channels)
	}
	if summary.Totals.LiveChannels != 1 {
		t.Errorf("live channels = %d, want 1", summary.Totals.LiveChannels)
	}
	if summary.Totals.Messages != 15 || summary.Totals.Commands != 3 || summary.Totals.ListenerActivations != 2 {
		t.Errorf("totals = %#v", summary.Totals)
	}
	if summary.Totals.UniqueUsers != 3 {
		t.Errorf("unique users = %d, want 3", summary.Totals.UniqueUsers)
	}
	if summary.Totals.Berts != 6 {
		t.Errorf("berts = %d, want 6 (discord excluded)", summary.Totals.Berts)
	}
	if summary.Totals.StreamedSeconds != 9000 {
		t.Errorf("streamed seconds = %d, want 9000", summary.Totals.StreamedSeconds)
	}

	if len(summary.Channels) != 2 {
		t.Fatalf("channels = %d, want 2", len(summary.Channels))
	}
	if !summary.Channels[0].Live || summary.Channels[0].Name != "alpha" {
		t.Errorf("live channel should sort first: %#v", summary.Channels[0])
	}
	if summary.Channels[0].Uptime != 5400 {
		t.Errorf("uptime = %d, want 5400", summary.Channels[0].Uptime)
	}
	if summary.Channels[0].Streamed != 5400 {
		t.Errorf("streamed = %d, want 5400", summary.Channels[0].Streamed)
	}
	if summary.Channels[0].Viewers != 42 || summary.Channels[0].Game != "Art" {
		t.Errorf("live metadata missing: %#v", summary.Channels[0])
	}

	if len(summary.TopCommands) != 2 || summary.TopCommands[0].Name != "hello" || summary.TopCommands[0].Count != 3 {
		t.Errorf("top commands = %#v", summary.TopCommands)
	}
	if len(summary.TopListeners) != 1 || summary.TopListeners[0].Name != "bert" || summary.TopListeners[0].Count != 3 {
		t.Errorf("top listeners = %#v", summary.TopListeners)
	}
	if len(summary.TopBerts) != 1 || summary.TopBerts[0].Name != "beert" || summary.TopBerts[0].Count != 4 {
		t.Errorf("top berts = %#v", summary.TopBerts)
	}
}

func TestBuildSiteSummaryWithoutServices(t *testing.T) {
	summary := buildSiteSummary(time.Now(), nil, nil, nil, nil, services.StreamStatsSnapshot{})

	if summary.Bot.Name != "pizza-son" {
		t.Errorf("bot name = %q", summary.Bot.Name)
	}
	if summary.Totals.Channels != 0 || summary.Totals.Messages != 0 || summary.Totals.StreamedSeconds != 0 {
		t.Errorf("totals = %#v, want zero", summary.Totals)
	}
	if summary.Channels == nil || summary.TopCommands == nil || summary.TopBerts == nil {
		t.Error("empty summaries should be empty slices, not nil")
	}
}

func TestHandleSiteSummary(t *testing.T) {
	ws := &WebService{}
	rec := httptest.NewRecorder()
	ws.handleSiteSummary(rec, httptest.NewRequest(http.MethodGet, "/api/site/summary", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("content type = %q", got)
	}
	if body := rec.Body.Bytes(); len(body) == 0 {
		t.Error("empty response body")
	}
}

func TestTopSiteCounts(t *testing.T) {
	got := topSiteCounts(map[string]int{
		"b": 2,
		"a": 2,
		"c": 5,
		"d": 0,
	}, 2)

	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Name != "c" || got[0].Count != 5 {
		t.Errorf("first = %#v", got[0])
	}
	if got[1].Name != "a" || got[1].Count != 2 {
		t.Errorf("ties should sort by name: %#v", got[1])
	}
}
