package web

import (
	"encoding/json"
	"log"
	"net/http"
	"pizza-son/internal/services"
	"regexp"
	"strings"
	"time"
)

var discordSnowflakeRe = regexp.MustCompile(`^\d{17,20}$`)
var discordEmojiRe = regexp.MustCompile(`^<a?:\w+:\d+>$`)

func isDiscordChannel(name string) bool {
	return discordSnowflakeRe.MatchString(name)
}

func isDiscordBert(name string) bool {
	return discordEmojiRe.MatchString(name)
}

type WebService struct {
	port string
}

func NewWebService(port string) *WebService {
	return &WebService{port: port}
}

func (ws *WebService) Start() {
	fs := http.FileServer(http.Dir("./web/public"))
	http.Handle("/", fs)

	// Clean URLs for panel
	http.HandleFunc("/panel", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/public/panel.html")
	})

	// API endpoints
	http.HandleFunc("/api/global", enableCORS(ws.handleGlobalStats))
	http.HandleFunc("/api/channel", enableCORS(ws.handleChannelStats))
	http.HandleFunc("/api/user", enableCORS(ws.handleUserStats))
	http.HandleFunc("/api/emotes", enableCORS(ws.handleChannelEmotes))
	http.HandleFunc("/api/live", enableCORS(services.LiveFeedInstance.HandleLiveFeed))

	// Panel API (auth + settings control)
	http.HandleFunc("/api/auth/login", enableCORS(ws.handleAuthLogin))
	http.HandleFunc("/api/auth/callback", enableCORS(ws.handleAuthCallback))
	http.HandleFunc("/api/auth/logout", enableCORS(requireSession(ws.handleAuthLogout)))
	http.HandleFunc("/api/auth/me", enableCORS(requireSession(ws.handleAuthMe)))
	http.HandleFunc("/api/panel/commands", enableCORS(requireSession(ws.handlePanelCommands)))
	http.HandleFunc("/api/panel/listeners", enableCORS(requireSession(ws.handlePanelListeners)))
	http.HandleFunc("/api/panel/save", enableCORS(requireSession(ws.handlePanelSave)))

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
	totalZazas := 0
	totalZazaLs := 0
	totalDoubleZazas := 0
	globalUsers := make(map[string]int)
	globalBerts := make(map[string]int)
	globalHourly := make(map[string]int)
	channelTotals := make(map[string]int)
	firstPersonPerBert := make(map[string]map[string]any) // channel -> bert -> {user, timestamp}
	var allGoldenzazabertEvents []map[string]any

	for chName, chData := range services.BertServiceInstance.Data {
		if isDiscordChannel(chName) {
			continue
		}
		firstPersonPerBert[chName] = make(map[string]any)

		// compute first person per bert from existing BertRecord data
		for username, stats := range chData.UserStats {
			for bertName, record := range stats.BertRecords {
				if isDiscordBert(bertName) {
					continue
				}
				existing, exists := firstPersonPerBert[chName][bertName]
				if !exists {
					firstPersonPerBert[chName][bertName] = map[string]any{
						"user":      username,
						"timestamp": record.FirstSeen,
					}
				} else if existingMap, ok := existing.(map[string]any); ok {
					if record.FirstSeen.Before(existingMap["timestamp"].(time.Time)) {
						firstPersonPerBert[chName][bertName] = map[string]any{
							"user":      username,
							"timestamp": record.FirstSeen,
						}
					}
				}
			}
		}

		for username, stats := range chData.UserStats {
			totalActivations += stats.TotalActivations
			totalZazas += stats.TotalZazas
			totalZazaLs += stats.TotalZazaLs
			totalDoubleZazas += stats.TotalDoubleZazas
			globalUsers[username] += stats.TotalActivations
			channelTotals[chName] += stats.TotalActivations

			for bertName, record := range stats.BertRecords {
				if !isDiscordBert(bertName) {
					globalBerts[bertName] += record.Count
				}
			}

			for hour, count := range stats.HourlyActivations {
				globalHourly[hour] += count
			}
		}

		for _, event := range chData.GoldenzazabertEvents {
			allGoldenzazabertEvents = append(allGoldenzazabertEvents, map[string]any{
				"username":  event.Username,
				"channel":   chName,
				"timestamp": event.Timestamp,
			})
		}
	}

	channelCount := 0
	for ch := range services.BertServiceInstance.Data {
		if !isDiscordChannel(ch) {
			channelCount++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_activations":        totalActivations,
		"total_zazas":              totalZazas,
		"total_zaza_ls":            totalZazaLs,
		"total_double_zazas":       totalDoubleZazas,
		"channel_count":            channelCount,
		"channel_totals":           channelTotals,
		"users":                    globalUsers,
		"berts":                    globalBerts,
		"hourly_timeline":          globalHourly,
		"first_person_per_bert":    firstPersonPerBert,
		"goldenzazabert_hall":      allGoldenzazabertEvents,
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
	totalZazas := 0
	totalZazaLs := 0
	totalDoubleZazas := 0
	userTotals := make(map[string]int)
	bertTotals := make(map[string]int)
	firstPersonPerBert := make(map[string]any)

	for user, uStats := range data.UserStats {
		totalBerts += uStats.TotalActivations
		totalZazas += uStats.TotalZazas
		totalZazaLs += uStats.TotalZazaLs
		totalDoubleZazas += uStats.TotalDoubleZazas
		userTotals[user] = uStats.TotalActivations

		for bert, record := range uStats.BertRecords {
			if isDiscordBert(bert) {
				continue
			}
			bertTotals[bert] += record.Count

			// track first person per bert
			existing, exists := firstPersonPerBert[bert]
			if !exists {
				firstPersonPerBert[bert] = map[string]any{
					"user":      user,
					"timestamp": record.FirstSeen,
				}
			} else if existingMap, ok := existing.(map[string]any); ok {
				if record.FirstSeen.Before(existingMap["timestamp"].(time.Time)) {
					firstPersonPerBert[bert] = map[string]any{
						"user":      user,
						"timestamp": record.FirstSeen,
					}
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"channel":                  channel,
		"total_activations":        totalBerts,
		"total_zazas":              totalZazas,
		"total_zaza_ls":            totalZazaLs,
		"total_double_zazas":       totalDoubleZazas,
		"active_berts":             data.Berts,
		"users":                    userTotals,
		"berts":                    bertTotals,
		"daily_timeline":           data.DailyActivations,
		"hourly_timeline":          data.HourlyActivations,
		"first_person_per_bert":    firstPersonPerBert,
		"goldenzazabert_events":    data.GoldenzazabertEvents,
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
	globalZazas := 0
	globalZazaLs := 0
	globalDoubleZazas := 0
	channelBreakdown := make(map[string]any)

	// loop through all channels to see where this user exists
	for chName, chData := range services.BertServiceInstance.Data {
		if isDiscordChannel(chName) {
			continue
		}
		if stats, hasUser := chData.UserStats[user]; hasUser {
			globalTotal += stats.TotalActivations
			globalZazas += stats.TotalZazas
			globalZazaLs += stats.TotalZazaLs
			globalDoubleZazas += stats.TotalDoubleZazas

			// calculate missing/collected for this specific channel
			collected := make(map[string]any)
			var missing []string
			for _, b := range chData.Berts {
				if isDiscordBert(b) {
					continue
				}
				if record, exists := stats.BertRecords[b]; exists && record.Count > 0 {
					collected[b] = map[string]any{
						"count":            record.Count,
						"zaza_count":       record.ZazaCount,
						"zaza_l_count":     record.ZazaLCount,
						"double_zaza_count": record.DoubleZazaCount,
						"first_seen":       record.FirstSeen,
						"last_seen":        record.LastSeen,
					}
				} else {
					missing = append(missing, b)
				}
			}

			channelBreakdown[chName] = map[string]any{
				"total":              stats.TotalActivations,
				"total_zazas":        stats.TotalZazas,
				"total_zaza_ls":      stats.TotalZazaLs,
				"total_double_zazas": stats.TotalDoubleZazas,
				"daily_timeline":     stats.DailyActivations,
				"hourly_timeline":    stats.HourlyActivations,
				"collected":          collected,
				"missing":            missing,
			}
		}
	}

	if globalTotal == 0 {
		http.Error(w, "User not found or has no berts", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"username":            user,
		"global_total":        globalTotal,
		"global_zazas":        globalZazas,
		"global_zaza_ls":      globalZazaLs,
		"global_double_zazas": globalDoubleZazas,
		"channels":            channelBreakdown,
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

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}
