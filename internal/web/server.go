package web

import (
	"encoding/json"
	"log"
	"net/http"
	"pizza-son/internal/services"
	"strings"
)

type WebService struct {
	port string
}

func NewWebService(port string) *WebService {
	return &WebService{port: port}
}

func (ws *WebService) Start() {
	fs := http.FileServer(http.Dir("./web/public"))
	http.Handle("/", fs)

	// API endpoints
	http.HandleFunc("/api/global", ws.handleGlobalStats)
	http.HandleFunc("/api/channel", ws.handleChannelStats)
	http.HandleFunc("/api/user", ws.handleUserStats)
	http.HandleFunc("/api/emotes", ws.handleChannelEmotes)

	log.Printf("[Web] Starting local dashboard on http://localhost%s/\n", ws.port)
	go func() {
		if err := http.ListenAndServe(ws.port, nil); err != nil {
			log.Fatalf("[Web] Server failed: %v", err)
		}
	}()
}

// All channels, all users
func (ws *WebService) handleGlobalStats(w http.ResponseWriter, r *http.Request) {
	services.BertServiceInstance.Mu.RLock()
	defer services.BertServiceInstance.Mu.RUnlock()

	totalActivations := 0
	globalUsers := make(map[string]int)
	globalBerts := make(map[string]int)
	globalHourly := make(map[string]int)

	for _, chData := range services.BertServiceInstance.Data {
		for username, stats := range chData.UserStats {
			totalActivations += stats.TotalActivations
			globalUsers[username] += stats.TotalActivations

			for bertName, record := range stats.BertRecords {
				globalBerts[bertName] += record.Count
			}

			for hour, count := range stats.HourlyActivations {
				globalHourly[hour] += count
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_activations": totalActivations,
		"channel_count":     len(services.BertServiceInstance.Data),
		"users":             globalUsers,
		"berts":             globalBerts,
		"hourly_timeline":   globalHourly,
	})
}

// Specific channel, all users
func (ws *WebService) handleChannelStats(w http.ResponseWriter, r *http.Request) {
	channel := strings.ToLower(r.URL.Query().Get("name"))
	if channel == "" {
		http.Error(w, "Missing channel name", http.StatusBadRequest)
		return
	}

	services.BertServiceInstance.Mu.RLock()
	defer services.BertServiceInstance.Mu.RUnlock()

	data, ok := services.BertServiceInstance.Data[channel]
	if !ok {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	totalBerts := 0
	userTotals := make(map[string]int)
	bertTotals := make(map[string]int)

	for user, uStats := range data.UserStats {
		totalBerts += uStats.TotalActivations
		userTotals[user] = uStats.TotalActivations

		for bert, record := range uStats.BertRecords {
			bertTotals[bert] += record.Count
		}
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"channel":           channel,
		"total_activations": totalBerts,
		"active_berts":      data.Berts,
		"users":             userTotals,
		"berts":             bertTotals,
		"daily_timeline":    data.DailyActivations,
		"hourly_timeline":   data.HourlyActivations,
	})
}

// Specific user, split by channel
func (ws *WebService) handleUserStats(w http.ResponseWriter, r *http.Request) {
	user := strings.ToLower(r.URL.Query().Get("name"))
	if user == "" {
		http.Error(w, "Missing user name", http.StatusBadRequest)
		return
	}

	services.BertServiceInstance.Mu.RLock()
	defer services.BertServiceInstance.Mu.RUnlock()

	globalTotal := 0
	channelBreakdown := make(map[string]any)

	// loop through all channels to see where this user exists
	for chName, chData := range services.BertServiceInstance.Data {
		if stats, hasUser := chData.UserStats[user]; hasUser {
			globalTotal += stats.TotalActivations

			// calculate missing/collected for this specific channel
			collected := make(map[string]any)
			var missing []string
			for _, b := range chData.Berts {
				if record, exists := stats.BertRecords[b]; exists && record.Count > 0 {
					collected[b] = map[string]any{
						"count":      record.Count,
						"zaza_count": record.ZazaCount,
						"first_seen": record.FirstSeen,
						"last_seen":  record.LastSeen,
					}
				} else {
					missing = append(missing, b)
				}
			}

			channelBreakdown[chName] = map[string]any{
				"total":           stats.TotalActivations,
				"total_zazas":     stats.TotalZazas,
				"daily_timeline":  stats.DailyActivations, // feeds the Chart.js timeline
				"hourly_timeline": stats.HourlyActivations,
				"collected":       collected,
				"missing":         missing,
			}
		}
	}

	if globalTotal == 0 {
		http.Error(w, "User not found or has no berts", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"username":     user,
		"global_total": globalTotal,
		"channels":     channelBreakdown,
	})
}

func (ws *WebService) handleChannelEmotes(w http.ResponseWriter, r *http.Request) {
	// Extract the channel name from the URL query parameters (e.g., /api/emotes?channel=pizza_tm)
	channel := r.URL.Query().Get("channel")
	if channel == "" {
		http.Error(w, "Missing channel parameter", http.StatusBadRequest)
		return
	}

	// fetch the emotes from your SevenTVService
	emotes := services.SevenTVServiceInstance.GetEmotes(channel)

	// set the headers so the browser knows it's getting JSON data
	w.Header().Set("Content-Type", "application/json")

	// encode the slice of SevenTVEmote structs directly into the HTTP response stream
	json.NewEncoder(w).Encode(emotes)
}
