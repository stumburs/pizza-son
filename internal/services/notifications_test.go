package services

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		in   time.Duration
		want string
	}{
		{"negative", -time.Minute, "0m"},
		{"seconds round down", 29 * time.Second, "0m"},
		{"seconds round up", 90 * time.Second, "2m"},
		{"under an hour", 59*time.Minute + 20*time.Second, "59m"},
		{"hour", time.Hour, "1h 0m"},
		{"hours and minutes", 2*time.Hour + 35*time.Minute, "2h 35m"},
		{"day", 25*time.Hour + 10*time.Minute, "1d 1h 10m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDuration(tt.in); got != tt.want {
				t.Fatalf("formatDuration(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestBuildLivePayload(t *testing.T) {
	started := time.Now().Add(-2*time.Hour - 15*time.Minute)
	info := StreamInfo{
		ChannelName:  "Sir_Lysergium",
		GameName:     "Pizza",
		StreamTitle:  "Making pizza",
		ViewerCount:  42,
		ThumbnailURL: "https://example.com/thumb-{width}x{height}.jpg",
		StartedAt:    started,
	}

	payload := buildLivePayload(info, true, "hey {{channel}}, playing {{game}}")
	if payload["content"] != "hey Sir_Lysergium, playing Pizza" {
		t.Fatalf("content = %q", payload["content"])
	}

	embed := embedFromPayload(t, payload)
	if embed["title"] != "Sir_Lysergium is live" {
		t.Fatalf("title = %q", embed["title"])
	}

	fields := fieldsFromEmbed(t, embed)
	if got := fieldByName(fields, "Game"); got != "Pizza" {
		t.Fatalf("Game = %q", got)
	}
	if got := fieldByName(fields, "Uptime"); got != "2h 15m" {
		t.Fatalf("Uptime = %q", got)
	}

	image, ok := embed["image"].(map[string]string)
	if !ok || !strings.HasPrefix(image["url"], "https://example.com/thumb-1280x720.jpg") {
		t.Fatalf("image = %#v", embed["image"])
	}
	imageURL, err := url.Parse(image["url"])
	if err != nil {
		t.Fatal(err)
	}
	if imageURL.Query().Get("t") == "" {
		t.Fatalf("update image URL has no cache buster: %s", image["url"])
	}

	// Uptime edits keep the original content; Discord does not re-ping on edit.
	if got := buildLivePayload(info, true, "")["content"]; got != "@everyone" {
		t.Fatalf("default update content = %q, want @everyone", got)
	}
}

func TestBuildLivePayloadFirstMessageHasNoUptime(t *testing.T) {
	payload := buildLivePayload(StreamInfo{
		ChannelName: "someone",
		GameName:    "Game",
		ViewerCount: 3,
	}, false, "")

	if payload["content"] != "@everyone" {
		t.Fatalf("content = %q", payload["content"])
	}
	embed := embedFromPayload(t, payload)
	if embed["title"] != "someone is now live!" {
		t.Fatalf("title = %q", embed["title"])
	}
	if got := fieldByName(fieldsFromEmbed(t, embed), "Uptime"); got != "" {
		t.Fatalf("Uptime should be absent, got %q", got)
	}
}

func TestBuildOfflinePayload(t *testing.T) {
	started := time.Unix(1_700_000_000, 0)
	lastLive := started.Add(90 * time.Minute)
	record := NotificationRecord{
		Channel:    "someone",
		StartedAt:  started,
		LastLiveAt: lastLive,
		Info: StreamInfo{
			ChannelName:  "Someone",
			GameName:     "Pizza",
			ThumbnailURL: "https://example.com/thumb-{width}x{height}.jpg",
		},
	}

	payload := buildOfflinePayload(record)

	// The edit keeps the original content; Discord does not re-ping on edit.
	if payload["content"] != "@everyone" {
		t.Fatalf("content = %q, want @everyone", payload["content"])
	}

	embed := embedFromPayload(t, payload)
	if embed["title"] != "Someone is no longer live" {
		t.Fatalf("title = %q", embed["title"])
	}
	description, _ := embed["description"].(string)
	if !strings.Contains(description, "<t:1700005400:f>") {
		t.Fatalf("description missing last-live timestamp: %q", description)
	}

	fields := fieldsFromEmbed(t, embed)
	if got := fieldByName(fields, "Streamed"); got != "1h 30m" {
		t.Fatalf("Streamed = %q", got)
	}
	if got := fieldByName(fields, "Game"); got != "Pizza" {
		t.Fatalf("Game = %q", got)
	}

	image, ok := embed["image"].(map[string]string)
	if !ok {
		t.Fatalf("image = %#v", embed["image"])
	}
	imageURL, err := url.Parse(image["url"])
	if err != nil {
		t.Fatal(err)
	}
	if got := imageURL.Query().Get("t"); got != "1700005400" {
		t.Fatalf("offline image cache buster = %q", got)
	}
}

func TestWebhookWithQueryPreservesExistingQuery(t *testing.T) {
	got, err := webhookWithQuery(
		"https://discord.com/api/webhooks/123/abc?thread_id=456",
		"wait", "true",
	)
	if err != nil {
		t.Fatal(err)
	}

	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if u.Query().Get("thread_id") != "456" {
		t.Fatalf("thread_id lost: %s", got)
	}
	if u.Query().Get("wait") != "true" {
		t.Fatalf("wait missing: %s", got)
	}
}

func TestWebhookMessageURL(t *testing.T) {
	got, err := webhookMessageURL(
		"https://discord.com/api/webhooks/123/abc?thread_id=456",
		"789",
	)
	if err != nil {
		t.Fatal(err)
	}

	u, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(u.Path, "/webhooks/123/abc/messages/789") {
		t.Fatalf("path = %q", u.Path)
	}
	if u.Query().Get("thread_id") != "456" {
		t.Fatalf("thread_id lost: %s", got)
	}
}

func TestSameStream(t *testing.T) {
	now := time.Now()
	if !sameStream(time.Time{}, now) {
		t.Fatal("zero timestamps should be treated as the same stream")
	}
	if !sameStream(now, now) {
		t.Fatal("equal timestamps should be the same stream")
	}
	if sameStream(now, now.Add(time.Second)) {
		t.Fatal("different start times should be different streams")
	}
}

func embedFromPayload(t *testing.T, payload map[string]any) map[string]any {
	t.Helper()
	embeds, ok := payload["embeds"].([]map[string]any)
	if !ok || len(embeds) != 1 {
		t.Fatalf("embeds = %#v", payload["embeds"])
	}
	return embeds[0]
}

func fieldsFromEmbed(t *testing.T, embed map[string]any) []map[string]any {
	t.Helper()
	fields, ok := embed["fields"].([]map[string]any)
	if !ok {
		t.Fatalf("fields = %#v", embed["fields"])
	}
	return fields
}

func fieldByName(fields []map[string]any, name string) string {
	for _, field := range fields {
		if field["name"] == name {
			value, _ := field["value"].(string)
			return value
		}
	}
	return ""
}
