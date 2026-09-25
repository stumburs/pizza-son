package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pizza-son/internal/services"
	"testing"
)

func setupBertService(t *testing.T, data map[string]*services.ChannelData) {
	t.Helper()
	old := services.BertServiceInstance
	services.BertServiceInstance = &services.BertService{Data: data}
	t.Cleanup(func() { services.BertServiceInstance = old })
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
