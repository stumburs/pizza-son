package web

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pizza-son/internal/commands"
	"pizza-son/internal/config"
	"pizza-son/internal/services"
	"sort"
	"strings"
	"sync"
	"time"
)

type panelSession struct {
	UserID      string
	Login       string
	DisplayName string
	Channels    []string
	CreatedAt   time.Time
}

type PanelService struct {
	mu       sync.Mutex
	sessions map[string]*panelSession
	sayFn    func(channel, message string)
	clientID string
}

var panelInstance *PanelService

func NewPanelService(sayFn func(channel, message string)) *PanelService {
	p := &PanelService{
		sessions: make(map[string]*panelSession),
		sayFn:    sayFn,
		clientID: config.Get().Twitch.ClientID,
	}
	panelInstance = p
	return p
}

func (p *PanelService) createSession(userID, login, displayName string, channels []string) string {
	b := make([]byte, 32)
	rand.Read(b)
	token := hex.EncodeToString(b)
	p.mu.Lock()
	p.sessions[token] = &panelSession{
		UserID:      userID,
		Login:       login,
		DisplayName: displayName,
		Channels:    channels,
		CreatedAt:   time.Now(),
	}
	p.mu.Unlock()
	return token
}

func (p *PanelService) getSession(token string) *panelSession {
	p.mu.Lock()
	defer p.mu.Unlock()
	s, ok := p.sessions[token]
	if !ok {
		return nil
	}
	if time.Since(s.CreatedAt) > 24*time.Hour {
		delete(p.sessions, token)
		return nil
	}
	return s
}

func (p *PanelService) deleteSession(token string) {
	p.mu.Lock()
	delete(p.sessions, token)
	p.mu.Unlock()
}

func getSessionToken(r *http.Request) string {
	if c, err := r.Cookie("panel_session"); err == nil && c.Value != "" {
		return c.Value
	}
	return r.Header.Get("X-Panel-Session")
}

func requireSession(next func(w http.ResponseWriter, r *http.Request, session *panelSession)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if panelInstance == nil {
			http.Error(w, "Panel not configured", http.StatusInternalServerError)
			return
		}
		token := getSessionToken(r)
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		s := panelInstance.getSession(token)
		if s == nil {
			http.Error(w, "Session expired", http.StatusUnauthorized)
			return
		}
		next(w, r, s)
	}
}

func (ws *WebService) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"client_id": config.Get().Twitch.ClientID,
	})
}

func (ws *WebService) handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token    string `json:"token"`
		Redirect string `json:"redirect"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Token == "" {
		http.Error(w, "Missing token", http.StatusBadRequest)
		return
	}

	userID, login, displayName, err := validateTwitchToken(body.Token)
	if err != nil {
		log.Printf("[Panel] Token validation failed: %v", err)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	chanList := config.Get().Bot.Channels
	configMods := config.Get().Bot.Moderators
	var manageable []string

	// Check if user is a bot moderator (config list) — grants full access
	isConfigMod := false
	for _, m := range configMods {
		if strings.EqualFold(m, login) || strings.EqualFold(m, userID) {
			isConfigMod = true
			break
		}
	}

	if isConfigMod {
		manageable = chanList
	} else {
		// Non-config-mods only get their own broadcaster channel
		for _, ch := range chanList {
			if strings.ToLower(ch) == login {
				manageable = append(manageable, ch)
			}
		}
	}

	if len(manageable) == 0 {
		http.Error(w, "Not a moderator of any bot channel", http.StatusForbidden)
		return
	}

	sort.Strings(manageable)
	unique := make([]string, 0, len(manageable))
	seen := make(map[string]bool)
	for _, ch := range manageable {
		if !seen[ch] {
			seen[ch] = true
			unique = append(unique, ch)
		}
	}

	token := panelInstance.createSession(userID, login, displayName, unique)

	http.SetCookie(w, &http.Cookie{
		Name:     "panel_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":        token,
		"username":     login,
		"display_name": displayName,
		"channels":     unique,
	})
}

func validateTwitchToken(userToken string) (userID, login, displayName string, err error) {
	req, _ := http.NewRequest("GET", "https://api.twitch.tv/helix/users", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Client-Id", config.Get().Twitch.ClientID)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", "", fmt.Errorf("api call failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID          string `json:"id"`
			Login       string `json:"login"`
			DisplayName string `json:"display_name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", "", fmt.Errorf("decode failed: %w", err)
	}
	if len(result.Data) == 0 {
		return "", "", "", fmt.Errorf("no user data")
	}
	u := result.Data[0]
	return u.ID, strings.ToLower(u.Login), u.DisplayName, nil
}

func (ws *WebService) handleAuthMe(w http.ResponseWriter, r *http.Request, s *panelSession) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"username":     s.Login,
		"display_name": s.DisplayName,
		"channels":     s.Channels,
	})
}

func (ws *WebService) handleAuthLogout(w http.ResponseWriter, r *http.Request, s *panelSession) {
	token := getSessionToken(r)
	if token != "" {
		panelInstance.deleteSession(token)
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "panel_session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	w.WriteHeader(http.StatusOK)
}

type panelCommandItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	CanDisable  bool   `json:"can_disable"`
}

type panelListenerItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	CanDisable  bool   `json:"can_disable"`
}

func (ws *WebService) handlePanelCommands(w http.ResponseWriter, r *http.Request, s *panelSession) {
	channel := strings.ToLower(r.URL.Query().Get("channel"))
	if channel == "" || !contains(s.Channels, channel) {
		http.Error(w, "Invalid channel", http.StatusBadRequest)
		return
	}

	var items []panelCommandItem
	registry := commands.GetRegistry()
	if registry == nil {
		http.Error(w, "Registry not available", http.StatusInternalServerError)
		return
	}

	cmds := registry.Commands()
	names := make([]string, 0, len(cmds))
	for n := range cmds {
		names = append(names, n)
	}
	sort.Strings(names)

	for _, name := range names {
		cmd := cmds[name]
		enabled := services.ChannelSettingsInstance.IsCommandEnabled(channel, name)
		canDisable := !commands.ProtectedCommands[name]
		items = append(items, panelCommandItem{
			Name:        name,
			Description: cmd.Description,
			Enabled:     enabled,
			CanDisable:  canDisable,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (ws *WebService) handlePanelListeners(w http.ResponseWriter, r *http.Request, s *panelSession) {
	channel := strings.ToLower(r.URL.Query().Get("channel"))
	if channel == "" || !contains(s.Channels, channel) {
		http.Error(w, "Invalid channel", http.StatusBadRequest)
		return
	}

	var items []panelListenerItem
	registry := commands.GetRegistry()
	if registry == nil {
		http.Error(w, "Registry not available", http.StatusInternalServerError)
		return
	}

	listeners := registry.Listeners()
	for _, l := range listeners {
		enabled := services.ChannelSettingsInstance.IsListenerEnabled(channel, l.Name)
		canDisable := !commands.ProtectedListeners[l.Name]
		items = append(items, panelListenerItem{
			Name:        l.Name,
			Description: l.Description,
			Enabled:     enabled,
			CanDisable:  canDisable,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

type panelChange struct {
	Type     string `json:"type"` // "command" or "listener"
	Name     string `json:"name"`
	Enable   bool   `json:"enable"`
}

type panelSaveRequest struct {
	Channel string        `json:"channel"`
	Changes []panelChange `json:"changes"`
}

func (ws *WebService) handlePanelSave(w http.ResponseWriter, r *http.Request, s *panelSession) {
	var req panelSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	req.Channel = strings.ToLower(req.Channel)
	if req.Channel == "" || !contains(s.Channels, req.Channel) {
		http.Error(w, "Invalid channel", http.StatusBadRequest)
		return
	}

	cmdCount := 0
	listenerCount := 0

	for _, ch := range req.Changes {
		switch ch.Type {
		case "command":
			if ch.Enable {
				services.ChannelSettingsInstance.EnableCommand(req.Channel, ch.Name)
			} else {
				if commands.ProtectedCommands[ch.Name] {
					continue
				}
				services.ChannelSettingsInstance.DisableCommand(req.Channel, ch.Name)
			}
			cmdCount++
		case "listener":
			if ch.Enable {
				services.ChannelSettingsInstance.EnableListener(req.Channel, ch.Name)
			} else {
				if commands.ProtectedListeners[ch.Name] {
					continue
				}
				services.ChannelSettingsInstance.DisableListener(req.Channel, ch.Name)
			}
			listenerCount++
		}
	}

	if panelInstance.sayFn != nil && (cmdCount > 0 || listenerCount > 0) {
		var parts []string
		if cmdCount > 0 {
			parts = append(parts, fmt.Sprintf("%d command(s)", cmdCount))
		}
		if listenerCount > 0 {
			parts = append(parts, fmt.Sprintf("%d listener(s)", listenerCount))
		}
		msg := fmt.Sprintf("🍕 %s updated %s via the web panel", s.DisplayName, strings.Join(parts, " and "))
		panelInstance.sayFn(req.Channel, msg)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"commands_updated":  cmdCount,
		"listeners_updated": listenerCount,
	})
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
