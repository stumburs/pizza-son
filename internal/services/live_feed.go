package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type BertUnlockEvent struct {
	Username string `json:"username"`
	Channel  string `json:"channel"`
	Bert     string `json:"bert"`
	EmoteID  string `json:"emote_id,omitempty"` // optional
}

type LiveFeedHub struct {
	mu          sync.Mutex
	clients     map[chan BertUnlockEvent]bool
	broadcastCh chan BertUnlockEvent
}

var LiveFeedInstance = &LiveFeedHub{
	clients:     make(map[chan BertUnlockEvent]bool),
	broadcastCh: make(chan BertUnlockEvent),
}

func init() {
	go LiveFeedInstance.run()
}

func (h *LiveFeedHub) run() {
	for event := range h.broadcastCh {
		h.mu.Lock()
		for clientCh := range h.clients {
			select {
			case clientCh <- event:
			default:
			}
		}
		h.mu.Unlock()
	}
}

func (h *LiveFeedHub) Broadcast(username, channel, bert, emoteID string) {
	h.broadcastCh <- BertUnlockEvent{
		Username: username,
		Channel:  channel,
		Bert:     bert,
		EmoteID:  emoteID,
	}
}

func (h *LiveFeedHub) HandleLiveFeed(w http.ResponseWriter, r *http.Request) {
	// sse headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no") // prevent nginx from buffering your feed

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// force immediate flush
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	// register client channel
	clientCh := make(chan BertUnlockEvent, 10)
	h.mu.Lock()
	h.clients[clientCh] = true
	h.mu.Unlock()

	// clean up client on close
	defer func() {
		h.mu.Lock()
		delete(h.clients, clientCh)
		h.mu.Unlock()
		close(clientCh)
	}()

	// keep alive ticker
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return

		case <-ticker.C:
			// keep live
			fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()

		case event := <-clientCh:
			jsonData, err := json.Marshal(event)
			if err != nil {
				continue
			}
			// write SSE payload
			fmt.Fprintf(w, "data: %s\n\n", jsonData)
			flusher.Flush()
		}
	}
}
