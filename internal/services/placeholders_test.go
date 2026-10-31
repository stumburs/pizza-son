package services

import (
	"pizza-son/internal/models"
	"testing"
	"time"
)

func TestFillPlaceholdersIsLive(t *testing.T) {
	old := TwitchServiceInstance
	t.Cleanup(func() { TwitchServiceInstance = old })

	now := time.Now()
	TwitchServiceInstance = &TwitchService{
		cache: map[string]cachedStreamInfo{
			"live":    {info: StreamInfo{StartedAt: now.Add(-time.Hour)}, fetchedAt: now},
			"offline": {info: StreamInfo{}, fetchedAt: now},
		},
	}

	tests := []struct {
		name     string
		channel  string
		platform models.Platform
		want     string
	}{
		{name: "live twitch channel", channel: "live", platform: models.PlatformTwitch, want: "true"},
		{name: "offline twitch channel", channel: "offline", platform: models.PlatformTwitch, want: "false"},
		{name: "discord channel", channel: "123456789012345678", platform: models.PlatformDiscord, want: "false"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			msg := models.Message{Channel: tc.channel, Platform: tc.platform}
			got := FillPlaceholders(msg, "is_live={{is_live}}")
			if want := "is_live=" + tc.want; got != want {
				t.Errorf("FillPlaceholders() = %q, want %q", got, want)
			}
		})
	}
}
