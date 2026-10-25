package web

import (
	"encoding/json"
	"net/http"
	"pizza-son/internal/config"
	"pizza-son/internal/models"
	"pizza-son/internal/services"
	"sort"
	"strings"
	"time"
)

const siteLeaderboardLimit = 8

// SiteCount is one entry in a public leaderboard.
type SiteCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type SiteBotInfo struct {
	Name            string `json:"name"`
	Prefix          string `json:"prefix"`
	TwitchChannels  int    `json:"twitch_channels"`
	DiscordChannels int    `json:"discord_channels"`
}

type SiteTotals struct {
	Channels            int `json:"channels"`
	LiveChannels        int `json:"live_channels"`
	Messages            int `json:"messages"`
	Commands            int `json:"commands"`
	ListenerActivations int `json:"listener_activations"`
	UniqueUsers         int `json:"unique_users"`
	Berts               int `json:"berts"`
}

type SiteChannel struct {
	Name      string          `json:"name"`
	Platform  models.Platform `json:"platform"`
	Live      bool            `json:"live"`
	Title     string          `json:"title,omitempty"`
	Game      string          `json:"game,omitempty"`
	Viewers   int             `json:"viewers,omitempty"`
	Thumbnail string          `json:"thumbnail,omitempty"`
	StartedAt *time.Time      `json:"started_at,omitempty"`
	Uptime    int64           `json:"uptime_seconds,omitempty"`

	Messages  int `json:"messages"`
	Commands  int `json:"commands"`
	Listeners int `json:"listeners"`
	Users     int `json:"users"`
	Berts     int `json:"berts"`
}

type SiteSummary struct {
	GeneratedAt  time.Time     `json:"generated_at"`
	Bot          SiteBotInfo   `json:"bot"`
	Totals       SiteTotals    `json:"totals"`
	Channels     []SiteChannel `json:"channels"`
	TopCommands  []SiteCount   `json:"top_commands"`
	TopListeners []SiteCount   `json:"top_listeners"`
	TopBerts     []SiteCount   `json:"top_berts"`
}

// siteActivityReader is the slice of the activity service the public summary
// needs. It exists so the summary can be tested without touching disk.
type siteActivityReader interface {
	GlobalStats() services.GlobalActivityStats
	ChannelStats(channel string) (services.ChannelActivity, bool)
}

// GET /api/site/summary
func (ws *WebService) handleSiteSummary(w http.ResponseWriter, r *http.Request) {
	cfg := config.Current()

	twitchChannels := []string{}
	if cfg != nil {
		twitchChannels = cfg.Bot.Channels
	}

	live := map[string]services.StreamInfo{}
	if services.TwitchServiceInstance != nil {
		live = services.TwitchServiceInstance.GetLiveStreams(twitchChannels)
	}

	summary := buildSiteSummary(time.Now(), cfg, services.ActivityServiceInstance, services.BertServiceInstance, live)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(summary)
}

// buildSiteSummary merges the configured channel list with the live Twitch
// state, the activity aggregates and the bert statistics.
func buildSiteSummary(
	now time.Time,
	cfg *config.Config,
	activity siteActivityReader,
	berts *services.BertService,
	live map[string]services.StreamInfo,
) SiteSummary {
	summary := SiteSummary{
		GeneratedAt:  now,
		Channels:     []SiteChannel{},
		TopCommands:  []SiteCount{},
		TopListeners: []SiteCount{},
		TopBerts:     []SiteCount{},
	}
	summary.Bot = SiteBotInfo{Name: "pizza-son", Prefix: "!"}

	if cfg != nil {
		if cfg.Twitch.User != "" {
			summary.Bot.Name = cfg.Twitch.User
		}
		if cfg.Bot.Prefix != "" {
			summary.Bot.Prefix = cfg.Bot.Prefix
		}
		summary.Bot.TwitchChannels = len(cfg.Bot.Channels)
		summary.Bot.DiscordChannels = len(cfg.Discord.Channels)
	}

	activityStats := services.GlobalActivityStats{}
	if activity != nil {
		activityStats = activity.GlobalStats()
	}

	channels := make(map[string]struct{})
	if cfg != nil {
		for _, name := range cfg.Bot.Channels {
			name = strings.ToLower(strings.TrimSpace(name))
			if name != "" {
				channels[name] = struct{}{}
			}
		}
	}
	for name, stats := range activityStats.Channels {
		if stats.Platform == models.PlatformDiscord {
			continue
		}
		channels[name] = struct{}{}
	}
	if cfg == nil {
		summary.Bot.TwitchChannels = len(channels)
	}

	bertTotalsByChannel, bertUsage := collectBertTotals(berts)

	seenUsers := make(map[string]struct{})
	commandUsage := make(map[string]int)
	listenerUsage := make(map[string]int)

	for name := range channels {
		var stats services.ChannelActivity
		if activity != nil {
			stats, _ = activity.ChannelStats(name)
		}

		channel := SiteChannel{
			Name:      name,
			Platform:  models.PlatformTwitch,
			Messages:  stats.Messages.Total,
			Commands:  stats.Commands.Total,
			Listeners: stats.Listeners.Total,
			Users:     len(stats.Users),
			Berts:     bertTotalsByChannel[name],
		}

		if info, ok := live[name]; ok && !info.StartedAt.IsZero() {
			channel.Live = true
			channel.Title = info.StreamTitle
			channel.Game = info.GameName
			channel.Viewers = info.ViewerCount
			channel.Thumbnail = info.ThumbnailURL
			started := info.StartedAt
			channel.StartedAt = &started
			if now.After(started) {
				channel.Uptime = int64(now.Sub(started).Seconds())
			}
			summary.Totals.LiveChannels++
		}

		summary.Channels = append(summary.Channels, channel)
		summary.Totals.Messages += channel.Messages
		summary.Totals.Commands += channel.Commands
		summary.Totals.ListenerActivations += channel.Listeners

		for user := range stats.Users {
			seenUsers[user] = struct{}{}
		}
		for command, count := range stats.CommandUsage {
			commandUsage[command] += count
		}
		for listener, count := range stats.ListenerUsage {
			listenerUsage[listener] += count
		}
	}

	summary.Totals.Channels = len(summary.Channels)
	summary.Totals.UniqueUsers = len(seenUsers)
	summary.Totals.Berts = sumSiteCounts(bertTotalsByChannel)

	sort.Slice(summary.Channels, func(i, j int) bool {
		a, b := summary.Channels[i], summary.Channels[j]
		if a.Live != b.Live {
			return a.Live
		}
		if a.Messages != b.Messages {
			return a.Messages > b.Messages
		}
		if a.Berts != b.Berts {
			return a.Berts > b.Berts
		}
		return a.Name < b.Name
	})

	summary.TopCommands = topSiteCounts(commandUsage, siteLeaderboardLimit)
	summary.TopListeners = topSiteCounts(listenerUsage, siteLeaderboardLimit)
	summary.TopBerts = topSiteCounts(bertUsage, siteLeaderboardLimit)
	return summary
}

func collectBertTotals(berts *services.BertService) (map[string]int, map[string]int) {
	byChannel := make(map[string]int)
	usage := make(map[string]int)
	if berts == nil {
		return byChannel, usage
	}

	berts.Mu.RLock()
	defer berts.Mu.RUnlock()

	for name, data := range berts.Data {
		if data == nil || services.IsDiscordChannel(name) {
			continue
		}
		key := strings.ToLower(name)
		for _, stats := range data.UserStats {
			byChannel[key] += stats.TotalActivations
			for bert, record := range stats.BertRecords {
				if services.IsDiscordBert(bert) {
					continue
				}
				usage[bert] += record.Count
			}
		}
	}
	return byChannel, usage
}

func topSiteCounts(counts map[string]int, limit int) []SiteCount {
	result := make([]SiteCount, 0, len(counts))
	for name, count := range counts {
		if count <= 0 {
			continue
		}
		result = append(result, SiteCount{Name: name, Count: count})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count != result[j].Count {
			return result[i].Count > result[j].Count
		}
		return result[i].Name < result[j].Name
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}

func sumSiteCounts(counts map[string]int) int {
	total := 0
	for _, count := range counts {
		total += count
	}
	return total
}
