package web

import (
	"encoding/json"
	"net/http"
	"pizza-son/internal/services"
	"strings"
)

// GET /api/activity/global
func (ws *WebService) handleActivityGlobal(w http.ResponseWriter, r *http.Request) {
	stats := services.ActivityServiceInstance.GlobalStats()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GET /api/activity/channel?name=<channel>
func (ws *WebService) handleActivityChannel(w http.ResponseWriter, r *http.Request) {
	channel := strings.TrimSpace(r.URL.Query().Get("name"))
	if channel == "" {
		http.Error(w, "Missing channel name", http.StatusBadRequest)
		return
	}

	stats, ok := services.ActivityServiceInstance.ChannelStats(channel)
	if !ok {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"channel":        strings.ToLower(channel),
		"platform":       stats.Platform,
		"messages":       stats.Messages,
		"commands":       stats.Commands,
		"listeners":      stats.Listeners,
		"command_usage":  stats.CommandUsage,
		"listener_usage": stats.ListenerUsage,
		"users":          stats.Users,
	})
}

// GET /api/activity/user?name=<username>
func (ws *WebService) handleActivityUser(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.URL.Query().Get("name"))
	if username == "" {
		http.Error(w, "Missing user name", http.StatusBadRequest)
		return
	}

	stats, ok := services.ActivityServiceInstance.UserStats(username)
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
