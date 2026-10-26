package web

import (
	"encoding/json"
	"html"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"pizza-son/internal/services"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	quotesPublicDir = "data/quotes"
	emotesPublicDir = "data/7tv"
)

type QuoteItem struct {
	Number    int    `json:"number"`
	Text      string `json:"text"`
	HTML      string `json:"html"`
	AddedBy   string `json:"added_by"`
	CreatedAt string `json:"created_at"`
}

type QuoteChannel struct {
	Name   string      `json:"name"`
	Count  int         `json:"count"`
	Quotes []QuoteItem `json:"quotes"`
}

type QuotesResponse struct {
	GeneratedAt time.Time      `json:"generated_at"`
	Total       int            `json:"total"`
	Channels    []QuoteChannel `json:"channels"`
}

// GET /api/quotes
//
// Quotes and emotes are read from disk on every request. The quote commands
// append to data/quotes, so the public page picks up new entries as soon as it
// refreshes without running a generator or redeploying anything.
func (ws *WebService) handleQuotes(w http.ResponseWriter, r *http.Request) {
	response := buildQuotesResponse(quotesPublicDir, emotesPublicDir, time.Now())

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(response)
}

func buildQuotesResponse(quotesDir, emotesDir string, now time.Time) QuotesResponse {
	response := QuotesResponse{
		GeneratedAt: now,
		Channels:    []QuoteChannel{},
	}

	entries, err := os.ReadDir(quotesDir)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Println("[Quotes] Failed to read quotes directory:", err)
		}
		return response
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		channel := strings.ToLower(strings.TrimSuffix(entry.Name(), ".json"))
		if channel == "" || services.IsDiscordChannel(channel) {
			continue
		}

		raw, err := os.ReadFile(filepath.Join(quotesDir, entry.Name()))
		if err != nil {
			log.Printf("[Quotes] Failed to read %s: %v", entry.Name(), err)
			continue
		}

		var quotes []services.Quote
		if err := json.Unmarshal(raw, &quotes); err != nil {
			log.Printf("[Quotes] Failed to parse %s: %v", entry.Name(), err)
			continue
		}

		emotes := loadQuoteEmotes(emotesDir, channel)
		items := make([]QuoteItem, 0, len(quotes))
		for i, quote := range quotes {
			items = append(items, QuoteItem{
				Number:    i + 1,
				Text:      quote.Text,
				HTML:      renderQuoteHTML(quote.Text, emotes),
				AddedBy:   quote.AddedBy,
				CreatedAt: quote.CreatedAt,
			})
		}

		response.Channels = append(response.Channels, QuoteChannel{
			Name:   channel,
			Count:  len(items),
			Quotes: items,
		})
		response.Total += len(items)
	}

	sort.Slice(response.Channels, func(i, j int) bool {
		if response.Channels[i].Count != response.Channels[j].Count {
			return response.Channels[i].Count > response.Channels[j].Count
		}
		return response.Channels[i].Name < response.Channels[j].Name
	})

	return response
}

func loadQuoteEmotes(dir, channel string) []services.SevenTVEmote {
	raw, err := os.ReadFile(filepath.Join(dir, channel+".json"))
	if err != nil {
		return nil
	}

	var emotes []services.SevenTVEmote
	if err := json.Unmarshal(raw, &emotes); err != nil {
		return nil
	}
	return emotes
}

// renderQuoteHTML escapes the quote and swaps whole-word 7TV emote names for
// images. The server does this so quote text is never injected as raw HTML.
func renderQuoteHTML(text string, emotes []services.SevenTVEmote) string {
	if text == "" {
		return ""
	}

	unique := make(map[string]string, len(emotes))
	for _, e := range emotes {
		if e.Name == "" || e.ID == "" {
			continue
		}
		if _, exists := unique[e.Name]; exists {
			continue
		}
		unique[e.Name] = e.ID
	}

	if len(unique) == 0 {
		return html.EscapeString(text)
	}

	names := make([]string, 0, len(unique))
	for name := range unique {
		names = append(names, name)
	}
	// Longest name first so overlapping names (BLA / BLALOO) resolve to the
	// more specific emote.
	sort.Slice(names, func(i, j int) bool {
		if len(names[i]) != len(names[j]) {
			return len(names[i]) > len(names[j])
		}
		return names[i] < names[j]
	})

	var b strings.Builder
	b.Grow(len(text) + 64)

	for i := 0; i < len(text); {
		matched := false
		for _, name := range names {
			if !strings.HasPrefix(text[i:], name) {
				continue
			}
			end := i + len(name)
			if !isWordBoundary(text, i, end) {
				continue
			}

			escaped := html.EscapeString(name)
			b.WriteString(`<img class="emote" src="https://cdn.7tv.app/emote/`)
			b.WriteString(html.EscapeString(unique[name]))
			b.WriteString(`/2x.webp" alt="`)
			b.WriteString(escaped)
			b.WriteString(`" title="`)
			b.WriteString(escaped)
			b.WriteString(`" loading="lazy">`)

			i = end
			matched = true
			break
		}
		if matched {
			continue
		}

		r, size := utf8.DecodeRuneInString(text[i:])
		b.WriteString(html.EscapeString(string(r)))
		i += size
	}

	return b.String()
}

// isWordBoundary mirrors the \b behaviour used by the old markdown generator:
// an emote only matches when it is not glued to an ASCII word character.
func isWordBoundary(text string, start, end int) bool {
	if start > 0 && isWordChar(text[start-1]) {
		return false
	}
	if end < len(text) && isWordChar(text[end]) {
		return false
	}
	return true
}

func isWordChar(c byte) bool {
	return c == '_' ||
		(c >= '0' && c <= '9') ||
		(c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z')
}
