package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pizza-son/internal/services"
	"testing"
	"time"
)

func panelBertTestData() map[string]*services.ChannelData {
	base := time.Date(2026, 10, 31, 12, 0, 0, 0, time.Local)
	return map[string]*services.ChannelData{
		"sir_lysergium": {
			Berts: []string{"bertnard", "beert"},
			UserStats: map[string]services.UserStats{
				"alice": {
					TotalActivations: 2,
					TotalZazas:       1,
					BertRecords: map[string]services.BertRecord{
						"bertnard": {Count: 2, ZazaCount: 1, FirstSeen: base, LastSeen: base.Add(90 * time.Minute)},
					},
					DailyActivations: map[string]int{"2026-10-31": 2},
					HourlyActivations: map[string]int{
						"2026-10-31 12:00": 1,
						"2026-10-31 13:00": 1,
					},
				},
				"bob": {
					TotalActivations: 1,
					BertRecords: map[string]services.BertRecord{
						"beert": {Count: 1, FirstSeen: base, LastSeen: base},
					},
					DailyActivations:  map[string]int{"2026-10-31": 1},
					HourlyActivations: map[string]int{"2026-10-31 12:00": 1},
				},
			},
			DailyActivations:  map[string]int{"2026-10-31": 3},
			HourlyActivations: map[string]int{"2026-10-31 12:00": 2, "2026-10-31 13:00": 1},
			ActivationLog: []services.BertActivation{
				{Username: "alice", Bert: "bertnard", Timestamp: base, IsZaza: true},
				{Username: "bob", Bert: "beert", Timestamp: base},
				{Username: "alice", Bert: "bertnard", Timestamp: base.Add(90 * time.Minute)},
			},
		},
	}
}

func panelTestSession(channels ...string) *panelSession {
	return &panelSession{
		UserID:      "1",
		Login:       "mod",
		DisplayName: "Mod",
		Channels:    channels,
		CreatedAt:   time.Now(),
	}
}

func postPanelJSON(t *testing.T, handler func(http.ResponseWriter, *http.Request, *panelSession), session *panelSession, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(payload))
	handler(rec, req, session)
	return rec
}

func TestPanelBertOverview(t *testing.T) {
	setupBertService(t, panelBertTestData())
	ws := &WebService{}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/panel/bert/overview?channel=sir_lysergium", nil)
	ws.handlePanelBertOverview(rec, req, panelTestSession("sir_lysergium"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}

	var response panelBertOverviewResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.TotalActivations != 3 {
		t.Errorf("total activations = %d, want 3", response.TotalActivations)
	}
	if response.UserCount != 2 || len(response.Users) != 2 {
		t.Fatalf("users = %d, want 2", len(response.Users))
	}
	if response.Users[0].Username != "alice" || response.Users[0].LogEntries != 2 {
		t.Errorf("first user = %+v, want alice with 2 log entries", response.Users[0])
	}
	if response.LogEntries != 3 {
		t.Errorf("log entries = %d, want 3", response.LogEntries)
	}
	if response.LogStarts == nil || response.LogEnds == nil {
		t.Error("log range should be set")
	}
}

func TestPanelBertOverviewRejectsForeignChannel(t *testing.T) {
	setupBertService(t, panelBertTestData())
	ws := &WebService{}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/panel/bert/overview?channel=sir_lysergium", nil)
	ws.handlePanelBertOverview(rec, req, panelTestSession("other_channel"))

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestPanelBertUser(t *testing.T) {
	setupBertService(t, panelBertTestData())
	ws := &WebService{}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/panel/bert/user?channel=sir_lysergium&user=alice", nil)
	ws.handlePanelBertUser(rec, req, panelTestSession("sir_lysergium"))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}

	var response panelBertUserResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Total != 2 || response.LogEntries != 2 {
		t.Errorf("total/log = %d/%d, want 2/2", response.Total, response.LogEntries)
	}
	if len(response.Records) != 1 || response.Records[0].Bert != "bertnard" {
		t.Errorf("records = %+v, want bertnard", response.Records)
	}
	if len(response.Log) != 2 || response.Log[0].Timestamp.Before(response.Log[1].Timestamp) {
		t.Errorf("log is not newest first: %+v", response.Log)
	}
}

func TestPanelBertPurge(t *testing.T) {
	data := panelBertTestData()
	setupBertService(t, data)
	ws := &WebService{}
	base := time.Date(2026, 10, 31, 12, 0, 0, 0, time.Local)

	rec := postPanelJSON(t, ws.handlePanelBertPurge, panelTestSession("sir_lysergium"), "/api/panel/bert/purge", map[string]any{
		"channel": "sir_lysergium",
		"user":    "alice",
		"from":    base.Add(30 * time.Minute).Format(time.RFC3339),
		"until":   base.Add(2 * time.Hour).Format(time.RFC3339),
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}

	var response panelBertPurgeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Removed != 1 {
		t.Errorf("removed = %d, want 1", response.Removed)
	}

	ch := services.BertServiceInstance.Data["sir_lysergium"]
	alice := ch.UserStats["alice"]
	if alice.TotalActivations != 1 {
		t.Errorf("alice total = %d, want 1", alice.TotalActivations)
	}
	if got := ch.UserStats["alice"].BertRecords["bertnard"]; got.Count != 1 || !got.LastSeen.Equal(base) {
		t.Errorf("bertnard record = %+v, want count 1 and last seen %s", got, base)
	}
	if got := ch.DailyActivations["2026-10-31"]; got != 2 {
		t.Errorf("channel daily = %d, want 2", got)
	}
}

func TestPanelBertPurgeRejectsUnloggedRange(t *testing.T) {
	data := panelBertTestData()
	setupBertService(t, data)
	ws := &WebService{}
	base := time.Date(2026, 10, 31, 12, 0, 0, 0, time.Local)

	rec := postPanelJSON(t, ws.handlePanelBertPurge, panelTestSession("sir_lysergium"), "/api/panel/bert/purge", map[string]any{
		"channel": "sir_lysergium",
		"user":    "alice",
		"from":    base.Add(-time.Hour).Format(time.RFC3339),
		"until":   base.Add(10 * time.Minute).Format(time.RFC3339),
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if got := services.BertServiceInstance.Data["sir_lysergium"].UserStats["alice"].TotalActivations; got != 2 {
		t.Errorf("alice total = %d, want 2 (unchanged)", got)
	}
}

func TestPanelBertDeleteUser(t *testing.T) {
	data := panelBertTestData()
	setupBertService(t, data)
	ws := &WebService{}

	rec := postPanelJSON(t, ws.handlePanelBertDeleteUser, panelTestSession("sir_lysergium"), "/api/panel/bert/delete-user", map[string]any{
		"channel": "sir_lysergium",
		"user":    "alice",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}

	var response map[string]int
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response["removed"] != 2 {
		t.Errorf("removed = %d, want 2", response["removed"])
	}
	ch := services.BertServiceInstance.Data["sir_lysergium"]
	if _, exists := ch.UserStats["alice"]; exists {
		t.Error("alice still exists")
	}
	if _, exists := ch.UserStats["bob"]; !exists {
		t.Error("bob was removed too")
	}
}

func TestPanelBertPurgeApproximateCount(t *testing.T) {
	base := time.Date(2026, 10, 31, 12, 0, 0, 0, time.Local)
	setupBertService(t, map[string]*services.ChannelData{
		"sir_lysergium": {
			Berts: []string{"beert"},
			UserStats: map[string]services.UserStats{
				"alice": {
					TotalActivations: 5,
					BertRecords: map[string]services.BertRecord{
						"beert": {Count: 5, FirstSeen: base, LastSeen: base.Add(50 * time.Minute)},
					},
					DailyActivations:  map[string]int{"2026-10-31": 5},
					HourlyActivations: map[string]int{"2026-10-31 12:00": 5},
				},
			},
			DailyActivations:  map[string]int{"2026-10-31": 5},
			HourlyActivations: map[string]int{"2026-10-31 12:00": 5},
		},
	})
	ws := &WebService{}

	rec := postPanelJSON(t, ws.handlePanelBertPurge, panelTestSession("sir_lysergium"), "/api/panel/bert/purge", map[string]any{
		"channel":     "sir_lysergium",
		"user":        "alice",
		"from":        base.Format(time.RFC3339),
		"until":       base.Add(time.Hour).Format(time.RFC3339),
		"approximate": true,
		"count":       2,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}

	var response panelBertPurgeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if !response.Approximate || response.Removed != 2 {
		t.Errorf("response = %+v, want 2 approximate removals", response)
	}
	alice := services.BertServiceInstance.Data["sir_lysergium"].UserStats["alice"]
	if alice.TotalActivations != 3 || alice.BertRecords["beert"].Count != 3 {
		t.Errorf("alice total/record = %d/%d, want 3/3",
			alice.TotalActivations, alice.BertRecords["beert"].Count)
	}
}

func TestPanelBertPurgeRequiresPost(t *testing.T) {
	setupBertService(t, panelBertTestData())
	ws := &WebService{}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/panel/bert/purge", nil)
	ws.handlePanelBertPurge(rec, req, panelTestSession("sir_lysergium"))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestPanelBertPurgeApproximate(t *testing.T) {
	base := time.Date(2026, 10, 31, 12, 0, 0, 0, time.Local)
	setupBertService(t, map[string]*services.ChannelData{
		"sir_lysergium": {
			Berts: []string{"beert"},
			UserStats: map[string]services.UserStats{
				"alice": {
					TotalActivations: 2,
					BertRecords: map[string]services.BertRecord{
						"beert": {Count: 2, FirstSeen: base, LastSeen: base.Add(30 * time.Minute)},
					},
					DailyActivations:  map[string]int{"2026-10-31": 2},
					HourlyActivations: map[string]int{"2026-10-31 12:00": 2},
				},
			},
			DailyActivations:  map[string]int{"2026-10-31": 2},
			HourlyActivations: map[string]int{"2026-10-31 12:00": 2},
		},
	})
	ws := &WebService{}

	rec := postPanelJSON(t, ws.handlePanelBertPurge, panelTestSession("sir_lysergium"), "/api/panel/bert/purge", map[string]any{
		"channel":     "sir_lysergium",
		"user":        "alice",
		"from":        base.Format(time.RFC3339),
		"until":       base.Add(time.Hour).Format(time.RFC3339),
		"approximate": true,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
	}

	var response panelBertPurgeResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if !response.Approximate {
		t.Error("response should be marked approximate")
	}
	if response.Removed != 2 || !response.UserRemoved {
		t.Errorf("response = %+v, want 2 removals and the user removed", response)
	}
	if _, exists := services.BertServiceInstance.Data["sir_lysergium"].UserStats["alice"]; exists {
		t.Error("alice should be gone from the channel")
	}
}
