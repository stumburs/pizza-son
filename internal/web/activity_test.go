package web

import (
	"net/http"
	"net/http/httptest"
	"pizza-son/internal/services"
	"testing"
)

func TestActivityHandlers(t *testing.T) {
	old := services.ActivityServiceInstance
	services.ActivityServiceInstance = &services.ActivityService{}
	t.Cleanup(func() { services.ActivityServiceInstance = old })

	ws := &WebService{}

	var global struct {
		TotalMessages int `json:"total_messages"`
		ChannelCount  int `json:"channel_count"`
	}
	getJSON(t, ws.handleActivityGlobal, "/api/activity/global", &global)
	if global.TotalMessages != 0 || global.ChannelCount != 0 {
		t.Errorf("empty global stats = %#v", global)
	}

	rec := httptest.NewRecorder()
	ws.handleActivityChannel(rec, httptest.NewRequest(http.MethodGet, "/api/activity/channel?name=missing", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("missing channel status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	rec = httptest.NewRecorder()
	ws.handleActivityUser(rec, httptest.NewRequest(http.MethodGet, "/api/activity/user?name=missing", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("missing user status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
