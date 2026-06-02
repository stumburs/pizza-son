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

	totalBerts := 0
	userTotals := make(map[string]int)
	bertTotals := make(map[string]int)

	for _, chData := range services.BertServiceInstance.Data {
		for user, uStats := range chData.UserStats {
			totalBerts += uStats.TotalActivations
			userTotals[user] += uStats.TotalActivations
			for bert, count := range uStats.BertCounts {
				bertTotals[bert] += count
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]any{
		"total_activations": totalBerts,
		"users":             userTotals,
		"berts":             bertTotals,
		"channel_count":     len(services.BertServiceInstance.Data),
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
		for bert, count := range uStats.BertCounts {
			bertTotals[bert] += count
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"channel":           channel,
		"total_activations": totalBerts,
		"active_berts":      data.Berts,
		"users":             userTotals,
		"berts":             bertTotals,
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
	channelBreakdown := make(map[string]interface{})

	// loop through all channels to see where this user exists
	for chName, chData := range services.BertServiceInstance.Data {
		if stats, hasUser := chData.UserStats[user]; hasUser {
			globalTotal += stats.TotalActivations

			// calculate missing/collected for this specific channel
			collected := make(map[string]int)
			var missing []string
			for _, b := range chData.Berts {
				if count, exists := stats.BertCounts[b]; exists && count > 0 {
					collected[b] = count
				} else {
					missing = append(missing, b)
				}
			}

			channelBreakdown[chName] = map[string]interface{}{
				"total":     stats.TotalActivations,
				"collected": collected,
				"missing":   missing,
			}
		}
	}

	if globalTotal == 0 {
		http.Error(w, "User not found or has no berts", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"username":     user,
		"global_total": globalTotal,
		"channels":     channelBreakdown,
	})
}
