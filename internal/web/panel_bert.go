package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pizza-son/internal/services"
	"sort"
	"strconv"
	"strings"
	"time"
)

type panelBertUserSummary struct {
	Username    string     `json:"username"`
	Total       int        `json:"total"`
	Zazas       int        `json:"zazas"`
	ZazaLs      int        `json:"zaza_ls"`
	DoubleZazas int        `json:"double_zazas"`
	LastSeen    *time.Time `json:"last_seen,omitempty"`
	LogEntries  int        `json:"log_entries"`
}

type panelBertOverviewResponse struct {
	Channel          string                 `json:"channel"`
	Berts            []string               `json:"berts"`
	TotalActivations int                    `json:"total_activations"`
	UserCount        int                    `json:"user_count"`
	LogEntries       int                    `json:"log_entries"`
	LogStarts        *time.Time             `json:"log_starts,omitempty"`
	LogEnds          *time.Time             `json:"log_ends,omitempty"`
	Users            []panelBertUserSummary `json:"users"`
}

type panelBertRecordItem struct {
	Bert        string    `json:"bert"`
	Count       int       `json:"count"`
	Zazas       int       `json:"zazas"`
	ZazaLs      int       `json:"zaza_ls"`
	DoubleZazas int       `json:"double_zazas"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
}

type panelBertLogItem struct {
	Bert       string    `json:"bert"`
	Timestamp  time.Time `json:"timestamp"`
	IsZaza     bool      `json:"is_zaza"`
	IsZazaL    bool      `json:"is_zaza_l"`
	DoubleZaza bool      `json:"double_zaza"`
}

type panelBertUserResponse struct {
	Username          string                `json:"username"`
	Total             int                   `json:"total"`
	Zazas             int                   `json:"zazas"`
	ZazaLs            int                   `json:"zaza_ls"`
	DoubleZazas       int                   `json:"double_zazas"`
	Records           []panelBertRecordItem `json:"records"`
	Log               []panelBertLogItem    `json:"log"`
	LogEntries        int                   `json:"log_entries"`
	ChannelLogEntries int                   `json:"channel_log_entries"`
	LogStarts         *time.Time            `json:"log_starts,omitempty"`
	LogEnds           *time.Time            `json:"log_ends,omitempty"`
}

type panelBertPurgeRequest struct {
	Channel     string `json:"channel"`
	User        string `json:"user"`
	From        string `json:"from"`
	Until       string `json:"until"`
	Partial     bool   `json:"partial"`
	Approximate bool   `json:"approximate"`
	Count       int    `json:"count"`
}

type panelBertPurgeResponse struct {
	Removed       int            `json:"removed"`
	RemovedByBert map[string]int `json:"removed_by_bert"`
	Partial       bool           `json:"partial"`
	Approximate   bool           `json:"approximate"`
	UserRemoved   bool           `json:"user_removed"`
	Warnings      []string       `json:"warnings"`
	LogStarts     *time.Time     `json:"log_starts,omitempty"`
}

// handlePanelBertOverview returns the channel-level bertcheck data needed by
// the management tab.
func (ws *WebService) handlePanelBertOverview(w http.ResponseWriter, r *http.Request, s *panelSession) {
	channel, ok := panelChannelFromQuery(w, r, s)
	if !ok {
		return
	}

	services.BertServiceInstance.Mu.RLock()
	defer services.BertServiceInstance.Mu.RUnlock()

	ch, ok := services.BertServiceInstance.Data[channel]
	if !ok {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}

	berts := ch.Berts
	if berts == nil {
		berts = []string{}
	}

	response := panelBertOverviewResponse{
		Channel:    channel,
		Berts:      berts,
		UserCount:  len(ch.UserStats),
		LogEntries: len(ch.ActivationLog),
		Users:      make([]panelBertUserSummary, 0, len(ch.UserStats)),
	}

	userLogEntries := make(map[string]int)
	for _, activation := range ch.ActivationLog {
		userLogEntries[activation.Username]++
		response.LogStarts = earlierTime(response.LogStarts, activation.Timestamp)
		response.LogEnds = laterTime(response.LogEnds, activation.Timestamp)
	}

	for username, stats := range ch.UserStats {
		response.TotalActivations += stats.TotalActivations
		response.Users = append(response.Users, panelBertUserSummary{
			Username:    username,
			Total:       stats.TotalActivations,
			Zazas:       stats.TotalZazas,
			ZazaLs:      stats.TotalZazaLs,
			DoubleZazas: stats.TotalDoubleZazas,
			LastSeen:    latestRecordSeen(stats),
			LogEntries:  userLogEntries[username],
		})
	}
	sort.Slice(response.Users, func(i, j int) bool {
		return response.Users[i].Username < response.Users[j].Username
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handlePanelBertUser returns one user's bert records and recent activations.
func (ws *WebService) handlePanelBertUser(w http.ResponseWriter, r *http.Request, s *panelSession) {
	channel, ok := panelChannelFromQuery(w, r, s)
	if !ok {
		return
	}
	user := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("user")))
	if user == "" {
		http.Error(w, "Missing user", http.StatusBadRequest)
		return
	}

	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > 500 {
		limit = 500
	}

	services.BertServiceInstance.Mu.RLock()
	defer services.BertServiceInstance.Mu.RUnlock()

	ch, ok := services.BertServiceInstance.Data[channel]
	if !ok {
		http.Error(w, "Channel not found", http.StatusNotFound)
		return
	}
	stats, ok := ch.UserStats[user]
	if !ok {
		http.Error(w, "User not found in this channel", http.StatusNotFound)
		return
	}

	response := panelBertUserResponse{
		Username:          user,
		Total:             stats.TotalActivations,
		Zazas:             stats.TotalZazas,
		ZazaLs:            stats.TotalZazaLs,
		DoubleZazas:       stats.TotalDoubleZazas,
		Records:           make([]panelBertRecordItem, 0, len(stats.BertRecords)),
		ChannelLogEntries: len(ch.ActivationLog),
	}

	for bert, record := range stats.BertRecords {
		response.Records = append(response.Records, panelBertRecordItem{
			Bert:        bert,
			Count:       record.Count,
			Zazas:       record.ZazaCount,
			ZazaLs:      record.ZazaLCount,
			DoubleZazas: record.DoubleZazaCount,
			FirstSeen:   record.FirstSeen,
			LastSeen:    record.LastSeen,
		})
	}
	sort.Slice(response.Records, func(i, j int) bool {
		if response.Records[i].Count != response.Records[j].Count {
			return response.Records[i].Count > response.Records[j].Count
		}
		return response.Records[i].Bert < response.Records[j].Bert
	})

	var userLog []panelBertLogItem
	var userLogStarts, userLogEnds time.Time
	for _, activation := range ch.ActivationLog {
		if activation.Username != user {
			continue
		}
		response.LogEntries++
		if userLogStarts.IsZero() || activation.Timestamp.Before(userLogStarts) {
			userLogStarts = activation.Timestamp
		}
		if userLogEnds.IsZero() || activation.Timestamp.After(userLogEnds) {
			userLogEnds = activation.Timestamp
		}
		userLog = append(userLog, panelBertLogItem{
			Bert:       activation.Bert,
			Timestamp:  activation.Timestamp,
			IsZaza:     activation.IsZaza,
			IsZazaL:    activation.IsZazaL,
			DoubleZaza: activation.DoubleZaza,
		})
	}
	if !userLogStarts.IsZero() {
		response.LogStarts = &userLogStarts
	}
	if !userLogEnds.IsZero() {
		response.LogEnds = &userLogEnds
	}
	sort.Slice(userLog, func(i, j int) bool {
		return userLog[i].Timestamp.After(userLog[j].Timestamp)
	})
	if len(userLog) > limit {
		userLog = userLog[:limit]
	}
	if userLog == nil {
		userLog = []panelBertLogItem{}
	}
	response.Log = userLog

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handlePanelBertPurge deletes one user's activations in a time range.
func (ws *WebService) handlePanelBertPurge(w http.ResponseWriter, r *http.Request, s *panelSession) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req panelBertPurgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	channel := strings.ToLower(strings.TrimSpace(req.Channel))
	user := strings.ToLower(strings.TrimSpace(req.User))
	if channel == "" || !contains(s.Channels, channel) {
		http.Error(w, "Invalid channel", http.StatusBadRequest)
		return
	}
	if user == "" {
		http.Error(w, "Missing user", http.StatusBadRequest)
		return
	}

	from, err := time.Parse(time.RFC3339, req.From)
	if err != nil {
		http.Error(w, "Invalid start time", http.StatusBadRequest)
		return
	}
	until, err := time.Parse(time.RFC3339, req.Until)
	if err != nil {
		http.Error(w, "Invalid end time", http.StatusBadRequest)
		return
	}
	if !from.Before(until) {
		http.Error(w, "The start of the range must be before the end", http.StatusBadRequest)
		return
	}

	options := services.PurgeOptions{
		User:         user,
		From:         from,
		To:           until,
		AllowPartial: req.Partial,
		Count:        req.Count,
	}
	if options.Count < 0 {
		options.Count = 0
	}

	var result services.PurgeResult
	if req.Approximate {
		result, err = services.BertServiceInstance.PurgeUserActivationsApprox(channel, options)
	} else {
		result, err = services.BertServiceInstance.PurgeUserActivations(channel, options)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[Panel] %s purged %d bertcheck(s) for %s in %s (%s to %s, approximate=%t)",
		s.Login, result.Removed, user, channel,
		from.Format(time.RFC3339), until.Format(time.RFC3339), result.Approximate)
	if result.Removed > 0 {
		message := fmt.Sprintf(
			"🍕 %s removed %d bertcheck(s) for %s via the web panel",
			s.DisplayName, result.Removed, user,
		)
		if result.Approximate {
			message = fmt.Sprintf(
				"🍕 %s approximately removed %d bertcheck(s) for %s via the web panel",
				s.DisplayName, result.Removed, user,
			)
		}
		ws.announcePanelAction(channel, message)
	}

	response := panelBertPurgeResponse{
		Removed:       result.Removed,
		RemovedByBert: result.RemovedByBert,
		Partial:       result.Partial,
		Approximate:   result.Approximate,
		UserRemoved:   result.UserRemoved,
		Warnings:      result.Warnings,
	}
	if !result.LogStarts.IsZero() {
		starts := result.LogStarts
		response.LogStarts = &starts
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handlePanelBertDeleteUser removes a user from a channel completely.
func (ws *WebService) handlePanelBertDeleteUser(w http.ResponseWriter, r *http.Request, s *panelSession) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Channel string `json:"channel"`
		User    string `json:"user"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	channel := strings.ToLower(strings.TrimSpace(req.Channel))
	user := strings.ToLower(strings.TrimSpace(req.User))
	if channel == "" || !contains(s.Channels, channel) {
		http.Error(w, "Invalid channel", http.StatusBadRequest)
		return
	}
	if user == "" {
		http.Error(w, "Missing user", http.StatusBadRequest)
		return
	}

	removed, err := services.BertServiceInstance.RemoveUser(channel, user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[Panel] %s removed user %s from %s (%d logged activations)", s.Login, user, channel, removed)
	ws.announcePanelAction(channel, fmt.Sprintf(
		"🍕 %s removed %s from the bertcheck stats via the web panel",
		s.DisplayName, user,
	))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"removed": removed})
}

func panelChannelFromQuery(w http.ResponseWriter, r *http.Request, s *panelSession) (string, bool) {
	channel := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("channel")))
	if channel == "" || !contains(s.Channels, channel) {
		http.Error(w, "Invalid channel", http.StatusBadRequest)
		return "", false
	}
	return channel, true
}

func (ws *WebService) announcePanelAction(channel, message string) {
	if panelInstance != nil && panelInstance.sayFn != nil {
		panelInstance.sayFn(channel, message)
	}
}

func latestRecordSeen(stats services.UserStats) *time.Time {
	var latest time.Time
	for _, record := range stats.BertRecords {
		if record.LastSeen.After(latest) {
			latest = record.LastSeen
		}
	}
	if latest.IsZero() {
		return nil
	}
	return &latest
}

// earlierTime returns a pointer to the earliest of current and candidate.
func earlierTime(current *time.Time, candidate time.Time) *time.Time {
	if candidate.IsZero() {
		return current
	}
	if current == nil || candidate.Before(*current) {
		value := candidate
		return &value
	}
	return current
}

// laterTime returns a pointer to the latest of current and candidate.
func laterTime(current *time.Time, candidate time.Time) *time.Time {
	if candidate.IsZero() {
		return current
	}
	if current == nil || candidate.After(*current) {
		value := candidate
		return &value
	}
	return current
}
