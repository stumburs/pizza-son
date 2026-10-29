package web

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"pizza-son/internal/services"
	"pizza-son/internal/webui"
	"strings"
	"time"

	"github.com/a-h/templ"
)

type WebService struct {
	port string
}

func NewWebService(port string) *WebService {
	return &WebService{port: port}
}

func (ws *WebService) Start() {
	staticFS, err := siteFS()
	if err != nil {
		log.Fatalf("[Web] Failed to load static files: %v", err)
	}
	mux := http.NewServeMux()
	mux.Handle("/", siteHandler(staticFS))

	// API endpoints
	mux.HandleFunc("/api/site/summary", enableCORS(ws.handleSiteSummary))
	mux.HandleFunc("/api/global", enableCORS(ws.handleGlobalStats))
	mux.HandleFunc("/api/channel", enableCORS(ws.handleChannelStats))
	mux.HandleFunc("/api/user", enableCORS(ws.handleUserStats))
	mux.HandleFunc("/api/emotes", enableCORS(ws.handleChannelEmotes))
	mux.HandleFunc("/api/quotes", enableCORS(ws.handleQuotes))
	mux.HandleFunc("/api/docs", enableCORS(ws.handleDocs))
	mux.HandleFunc("/api/activity/global", enableCORS(ws.handleActivityGlobal))
	mux.HandleFunc("/api/activity/channel", enableCORS(ws.handleActivityChannel))
	mux.HandleFunc("/api/activity/user", enableCORS(ws.handleActivityUser))
	mux.HandleFunc("/api/vote/submit", enableCORS(ws.handleVoteSubmit))
	mux.HandleFunc("/api/vote/results", enableCORS(ws.handleVoteResults))
	mux.HandleFunc("/api/live", enableCORS(services.LiveFeedInstance.HandleLiveFeed))

	// Panel API (auth + settings control)
	mux.HandleFunc("/api/auth/login", enableCORS(ws.handleAuthLogin))
	mux.HandleFunc("/api/auth/callback", enableCORS(ws.handleAuthCallback))
	mux.HandleFunc("/api/auth/logout", enableCORS(requireSession(ws.handleAuthLogout)))
	mux.HandleFunc("/api/auth/me", enableCORS(requireSession(ws.handleAuthMe)))
	mux.HandleFunc("/api/panel/commands", enableCORS(requireSession(ws.handlePanelCommands)))
	mux.HandleFunc("/api/panel/listeners", enableCORS(requireSession(ws.handlePanelListeners)))
	mux.HandleFunc("/api/panel/save", enableCORS(requireSession(ws.handlePanelSave)))

	log.Printf("[Web] Starting local dashboard on http://localhost%s/\n", ws.port)
	go func() {
		if err := http.ListenAndServe(ws.port, gzipMiddleware(mux)); err != nil {
			log.Fatalf("[Web] Server failed: %v", err)
		}
	}()
}

// siteFS returns the embedded UI in production. PIZZA_SON_DEV=1 serves the
// files from disk instead so CSS and JS changes do not require a rebuild.
// Templ pages always need a rebuild because they are compiled into the binary.
func siteFS() (fs.FS, error) {
	if os.Getenv("PIZZA_SON_DEV") != "" {
		return os.DirFS("./internal/webui/public"), nil
	}
	return fs.Sub(webui.Files, "public")
}

// siteHandler serves the templ-rendered pages, embedded CSS/JS assets, and the
// 404 page.
func siteHandler(staticFS fs.FS) http.Handler {
	fileServer := http.FileServerFS(staticFS)

	pages := map[string]templ.Component{
		"/":          webui.OverviewPage(),
		"/bertstats": webui.BertsPage(),
		"/bertle":    webui.BertlePage(),
		"/quotes":    webui.QuotesPage(),
		"/docs":      webui.DocsPage(),
		"/vote":      webui.VotePage(),
		"/panel":     webui.PanelPage(),
	}
	notFound := webui.ErrorPage(http.StatusNotFound)
	serverError := webui.ErrorPage(http.StatusInternalServerError)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		pagePath := path
		if pagePath != "/" {
			pagePath = strings.TrimSuffix(pagePath, "/")
		}

		if page, ok := pages[pagePath]; ok {
			renderPage(w, r, page, http.StatusOK, serverError)
			return
		}

		if pagePath == "/berts" {
			http.Redirect(w, r, "/bertstats", http.StatusMovedPermanently)
			return
		}

		if name := strings.TrimPrefix(path, "/"); name != "" {
			if info, err := fs.Stat(staticFS, name); err == nil && !info.IsDir() {
				setAssetCacheHeaders(w, r)
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		renderPage(w, r, notFound, http.StatusNotFound, nil)
	})
}

// renderPage renders into a buffer first so a failed template cannot send a
// half-written page or a 200 with an error body. Normal pages fall back to the
// 500 page; the error pages themselves fall back to plain text.
func renderPage(w http.ResponseWriter, r *http.Request, page templ.Component, status int, fallback templ.Component) {
	w.Header().Set("Cache-Control", "no-cache")

	var buf bytes.Buffer
	if err := page.Render(r.Context(), &buf); err != nil {
		log.Printf("[Web] Failed to render %s: %v", r.URL.Path, err)
		if fallback == nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		buf.Reset()
		status = http.StatusInternalServerError
		if err := fallback.Render(r.Context(), &buf); err != nil {
			log.Printf("[Web] Failed to render 500 page: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

// setAssetCacheHeaders gives versioned asset URLs a long immutable cache. The
// embedded filesystem has no modification times, so unversioned requests have
// to revalidate or they could never be cached safely.
func setAssetCacheHeaders(w http.ResponseWriter, r *http.Request) {
	if webui.Version == "" || webui.Version == "dev" || os.Getenv("PIZZA_SON_DEV") != "" {
		w.Header().Set("Cache-Control", "no-cache")
		return
	}

	if r.URL.Query().Get("v") == webui.Version {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
}

// All channels, all users
func (ws *WebService) handleGlobalStats(w http.ResponseWriter, r *http.Request) {
	services.BertServiceInstance.Mu.RLock()
	defer services.BertServiceInstance.Mu.RUnlock()

	// same source of truth as the milestone counter
	totalActivations := services.BertServiceInstance.GlobalTotalLocked()
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
		if services.IsDiscordChannel(chName) {
			continue
		}
		firstPersonPerBert[chName] = make(map[string]any)

		// compute first person per bert from existing BertRecord data
		for username, stats := range chData.UserStats {
			for bertName, record := range stats.BertRecords {
				if services.IsDiscordBert(bertName) {
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
			totalZazas += stats.TotalZazas
			totalZazaLs += stats.TotalZazaLs
			totalDoubleZazas += stats.TotalDoubleZazas
			globalUsers[username] += stats.TotalActivations
			channelTotals[chName] += stats.TotalActivations

			for bertName, record := range stats.BertRecords {
				if !services.IsDiscordBert(bertName) {
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
		if !services.IsDiscordChannel(ch) {
			channelCount++
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_activations":     totalActivations,
		"total_zazas":           totalZazas,
		"total_zaza_ls":         totalZazaLs,
		"total_double_zazas":    totalDoubleZazas,
		"channel_count":         channelCount,
		"channel_totals":        channelTotals,
		"users":                 globalUsers,
		"berts":                 globalBerts,
		"hourly_timeline":       globalHourly,
		"first_person_per_bert": firstPersonPerBert,
		"goldenzazabert_hall":   allGoldenzazabertEvents,
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
			if services.IsDiscordBert(bert) {
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
		"channel":               channel,
		"total_activations":     totalBerts,
		"total_zazas":           totalZazas,
		"total_zaza_ls":         totalZazaLs,
		"total_double_zazas":    totalDoubleZazas,
		"active_berts":          data.Berts,
		"users":                 userTotals,
		"berts":                 bertTotals,
		"daily_timeline":        data.DailyActivations,
		"hourly_timeline":       data.HourlyActivations,
		"first_person_per_bert": firstPersonPerBert,
		"goldenzazabert_events": data.GoldenzazabertEvents,
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
		if services.IsDiscordChannel(chName) {
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
				if services.IsDiscordBert(b) {
					continue
				}
				if record, exists := stats.BertRecords[b]; exists && record.Count > 0 {
					collected[b] = map[string]any{
						"count":             record.Count,
						"zaza_count":        record.ZazaCount,
						"zaza_l_count":      record.ZazaLCount,
						"double_zaza_count": record.DoubleZazaCount,
						"first_seen":        record.FirstSeen,
						"last_seen":         record.LastSeen,
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
