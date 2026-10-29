package web

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"pizza-son/internal/services"
	"pizza-son/internal/webui"
	"strings"
	"testing"
)

func setupBertService(t *testing.T, data map[string]*services.ChannelData) {
	t.Helper()
	old := services.BertServiceInstance
	services.BertServiceInstance = &services.BertService{Data: data}
	t.Cleanup(func() { services.BertServiceInstance = old })
}

func TestSiteFSContainsEmbeddedUI(t *testing.T) {
	t.Setenv("PIZZA_SON_DEV", "")

	staticFS, err := siteFS()
	if err != nil {
		t.Fatalf("siteFS: %v", err)
	}

	for _, name := range []string{"css/site.css", "js/bertle.js", "js/kofi.js"} {
		file, err := staticFS.Open(name)
		if err != nil {
			t.Errorf("%s is not embedded: %v", name, err)
			continue
		}
		file.Close()
	}
}

func TestSiteHandlerServesPagesAndAssets(t *testing.T) {
	t.Setenv("PIZZA_SON_DEV", "")

	staticFS, err := siteFS()
	if err != nil {
		t.Fatalf("siteFS: %v", err)
	}
	handler := siteHandler(staticFS)

	tests := []struct {
		path        string
		status      int
		contentType string
		contains    string
	}{
		{"/", http.StatusOK, "text/html", "class=\"topbar\""},
		{"/bertstats/", http.StatusOK, "text/html", "Bert stats"},
		{"/bertle", http.StatusOK, "text/html", "Bertle"},
		{"/quotes", http.StatusOK, "text/html", "Quotes"},
		{"/docs", http.StatusOK, "text/html", "Docs"},
		{"/vote", http.StatusOK, "text/html", "Bert vote"},
		{"/panel", http.StatusOK, "text/html", "Control panel"},
		{"/css/site.css", http.StatusOK, "text/css", ".page-berts"},
		{"/js/bertle.js", http.StatusOK, "", "Bertle"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))

			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d", rec.Code, tc.status)
			}
			if tc.contentType != "" {
				if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, tc.contentType) {
					t.Errorf("Content-Type = %q, want prefix %q", got, tc.contentType)
				}
			}

			body, err := io.ReadAll(rec.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			if !strings.Contains(string(body), tc.contains) {
				t.Errorf("body does not contain %q", tc.contains)
			}
		})
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/berts", nil))
	if rec.Code != http.StatusMovedPermanently {
		t.Errorf("/berts status = %d, want %d", rec.Code, http.StatusMovedPermanently)
	}
	if location := rec.Header().Get("Location"); location != "/bertstats" {
		t.Errorf("/berts Location = %q, want %q", location, "/bertstats")
	}
}

func TestSiteHandlerServes404Page(t *testing.T) {
	t.Setenv("PIZZA_SON_DEV", "")

	staticFS, err := siteFS()
	if err != nil {
		t.Fatalf("siteFS: %v", err)
	}
	handler := siteHandler(staticFS)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/does-not-exist", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`class="site-main page-error"`,
		`class="error-emote"`,
		"404 — page not found",
		`class="error-home"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("body does not contain %q", want)
		}
	}
	if strings.Contains(body, "/js/error.js") {
		t.Error("404 page should not load a page script")
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
		t.Errorf("Cache-Control = %q, want no-cache", got)
	}
}

func TestAssetCacheHeaders(t *testing.T) {
	t.Setenv("PIZZA_SON_DEV", "")

	staticFS, err := siteFS()
	if err != nil {
		t.Fatalf("siteFS: %v", err)
	}
	handler := siteHandler(staticFS)

	old := webui.Version
	webui.Version = "test123"
	t.Cleanup(func() { webui.Version = old })

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/css/site.css?v=test123", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("versioned asset status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Errorf("versioned Cache-Control = %q", got)
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/css/site.css", nil))
	if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
		t.Errorf("unversioned Cache-Control = %q, want no-cache", got)
	}
}

type failingComponent struct{}

func (failingComponent) Render(context.Context, io.Writer) error {
	return errors.New("boom")
}

func TestRenderPageFallsBackTo500(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	renderPage(rec, req, failingComponent{}, http.StatusOK, webui.ErrorPage(http.StatusInternalServerError))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if got := rec.Header().Get("Content-Type"); !strings.HasPrefix(got, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", got)
	}
	if !strings.Contains(rec.Body.String(), "500 — something broke") {
		t.Error("body does not contain the 500 page")
	}
}

func getJSON(t *testing.T, handler http.HandlerFunc, url string, out any) {
	t.Helper()
	rec := httptest.NewRecorder()
	handler(rec, httptest.NewRequest(http.MethodGet, url, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: status %d: %s", url, rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
}

// The numbers the site reports and the numbers !bertstats reports have to
// come from the same data, otherwise they drift apart.
func TestSiteTotalsMatchBertstatsCommand(t *testing.T) {
	setupBertService(t, map[string]*services.ChannelData{
		"sir_lysergium": {
			Berts: []string{"bertnard", "beert"},
			UserStats: map[string]services.UserStats{
				"alice": {
					TotalActivations: 10,
					BertRecords: map[string]services.BertRecord{
						"bertnard": {Count: 8},
						"gonebert": {Count: 2},
					},
				},
				"bob": {
					TotalActivations: 5,
					BertRecords: map[string]services.BertRecord{
						"beert": {Count: 5},
					},
				},
			},
		},
		"123456789012345678": {
			Berts: []string{"<a:bert:1>"},
			UserStats: map[string]services.UserStats{
				"alice": {
					TotalActivations: 7,
					BertRecords: map[string]services.BertRecord{
						"<a:bert:1>": {Count: 7},
					},
				},
			},
		},
	})
	ws := &WebService{}

	var global struct {
		TotalActivations int            `json:"total_activations"`
		ChannelTotals    map[string]int `json:"channel_totals"`
		Users            map[string]int `json:"users"`
	}
	getJSON(t, ws.handleGlobalStats, "/api/global", &global)

	var channel struct {
		TotalActivations int            `json:"total_activations"`
		Users            map[string]int `json:"users"`
	}
	getJSON(t, ws.handleChannelStats, "/api/channel?name=sir_lysergium", &channel)

	var user struct {
		GlobalTotal int `json:"global_total"`
		Channels    map[string]struct {
			Total int `json:"total"`
		} `json:"channels"`
	}
	getJSON(t, ws.handleUserStats, "/api/user?name=alice", &user)

	cmd := services.BertServiceInstance.GetUserStats("sir_lysergium", "alice")

	if global.TotalActivations != 15 {
		t.Errorf("global total = %d, want 15 (Discord channels excluded)", global.TotalActivations)
	}
	if _, ok := global.ChannelTotals["123456789012345678"]; ok {
		t.Error("discord channel leaked into channel_totals")
	}

	// !bertstats "berts by all chatters" vs site channel total
	if cmd.ChannelTotalBertchecks != channel.TotalActivations || channel.TotalActivations != 15 {
		t.Errorf("channel total: command %d, site %d, want 15", cmd.ChannelTotalBertchecks, channel.TotalActivations)
	}
	if global.ChannelTotals["sir_lysergium"] != channel.TotalActivations {
		t.Errorf("global channel_totals: %d, channel endpoint: %d", global.ChannelTotals["sir_lysergium"], channel.TotalActivations)
	}

	// !bertstats "total berts" vs site user totals
	if cmd.TotalBertchecks != channel.Users["alice"] || channel.Users["alice"] != 10 {
		t.Errorf("user channel total: command %d, site %d, want 10", cmd.TotalBertchecks, channel.Users["alice"])
	}
	if user.Channels["sir_lysergium"].Total != channel.Users["alice"] {
		t.Errorf("user endpoint channel total: %d, channel endpoint: %d", user.Channels["sir_lysergium"].Total, channel.Users["alice"])
	}
	if user.GlobalTotal != global.Users["alice"] || global.Users["alice"] != 10 {
		t.Errorf("user global: endpoint %d, global endpoint %d, want 10 (discord excluded)", user.GlobalTotal, global.Users["alice"])
	}

	// the milestone counter uses the same number as the site
	if milestoneTotal := services.BertServiceInstance.GlobalTotalLocked(); milestoneTotal != global.TotalActivations {
		t.Errorf("milestone global total %d != site global total %d", milestoneTotal, global.TotalActivations)
	}
}
