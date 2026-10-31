package services

import (
	"testing"
	"time"
)

func TestIsChannelLiveUsesStreamState(t *testing.T) {
	old := TwitchServiceInstance
	t.Cleanup(func() { TwitchServiceInstance = old })

	TwitchServiceInstance = &TwitchService{
		cache: map[string]cachedStreamInfo{
			"live":    {info: StreamInfo{StartedAt: time.Now().Add(-time.Hour)}, fetchedAt: time.Now()},
			"offline": {info: StreamInfo{}, fetchedAt: time.Now()},
		},
	}

	if !IsChannelLive("live") {
		t.Error("channel with a stream start time should be live")
	}
	if IsChannelLive("offline") {
		t.Error("channel without a stream start time should be offline")
	}
}

func TestIsChannelLiveFallsBackToLive(t *testing.T) {
	old := TwitchServiceInstance
	t.Cleanup(func() { TwitchServiceInstance = old })

	// Discord channels have no Twitch stream state.
	if !IsChannelLive("123456789012345678") {
		t.Error("discord channel should be treated as live")
	}

	// No Twitch service configured.
	TwitchServiceInstance = nil
	if !IsChannelLive("sir_lysergium") {
		t.Error("missing twitch service should be treated as live")
	}

	// Twitch service without a working client reports an error, which must not
	// be mistaken for an offline stream.
	TwitchServiceInstance = &TwitchService{}
	if !IsChannelLive("sir_lysergium") {
		t.Error("twitch lookup failure should be treated as live")
	}
}
