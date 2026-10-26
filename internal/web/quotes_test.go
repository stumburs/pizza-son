package web

import (
	"os"
	"path/filepath"
	"pizza-son/internal/services"
	"strings"
	"testing"
	"time"
)

func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestBuildQuotesResponseReadsFilesAndRendersEmotes(t *testing.T) {
	root := t.TempDir()
	quotesDir := filepath.Join(root, "quotes")
	emotesDir := filepath.Join(root, "7tv")

	writeTestFile(t, quotesDir, "beta.json", `[
		{"text":"beta one","added_by":"someone","created_at":"2026-03-01"},
		{"text":"beta two","added_by":"someone else","created_at":"2026-03-02"}
	]`)
	writeTestFile(t, quotesDir, "alpha.json", `[
		{"text":"hello wideVIBE <script>alert(1)</script>","added_by":"pizza_tm","created_at":"2026-03-01"},
		{"text":"xwideVIBEx should stay text","added_by":"pizza_tm","created_at":"2026-03-03"},
		{"text":"wideVIBE","added_by":"pizza_son","created_at":"2026-03-04"}
	]`)
	writeTestFile(t, emotesDir, "alpha.json", `[
		{"name":"wideVIBE","id":"01G1GXCR380004YN3NKDRR9QHD"}
	]`)
	writeTestFile(t, quotesDir, "1543683538598232064.json", `[
		{"text":"should not be public","added_by":"discord","created_at":"2026-03-01"}
	]`)

	response := buildQuotesResponse(quotesDir, emotesDir, time.Now())

	if response.Total != 5 {
		t.Fatalf("total = %d, want 5", response.Total)
	}
	if len(response.Channels) != 2 {
		t.Fatalf("channels = %d, want 2 (discord excluded)", len(response.Channels))
	}
	for _, channel := range response.Channels {
		if services.IsDiscordChannel(channel.Name) {
			t.Errorf("discord channel %q was included", channel.Name)
		}
	}
	if response.Channels[0].Name != "alpha" || response.Channels[1].Name != "beta" {
		t.Fatalf("channel order = %s, %s; want alpha, beta", response.Channels[0].Name, response.Channels[1].Name)
	}
	if response.Channels[0].Count != 3 || response.Channels[1].Count != 2 {
		t.Errorf("counts = %d, %d; want 3, 2", response.Channels[0].Count, response.Channels[1].Count)
	}

	alpha := response.Channels[0]
	if alpha.Quotes[0].Number != 1 || alpha.Quotes[2].Number != 3 {
		t.Errorf("quote numbers = %d, %d; want 1, 3", alpha.Quotes[0].Number, alpha.Quotes[2].Number)
	}

	first := alpha.Quotes[0].HTML
	if !strings.Contains(first, `src="https://cdn.7tv.app/emote/01G1GXCR380004YN3NKDRR9QHD/2x.webp"`) {
		t.Errorf("emote was not rendered: %s", first)
	}
	if strings.Contains(first, "<script>") {
		t.Errorf("quote HTML was not escaped: %s", first)
	}
	if !strings.Contains(first, "&lt;script&gt;") {
		t.Errorf("escaped script missing: %s", first)
	}
	if got := alpha.Quotes[1].HTML; strings.Contains(got, "<img") {
		t.Errorf("partial word was replaced: %s", got)
	}
	if got := alpha.Quotes[2].HTML; !strings.Contains(got, "<img") {
		t.Errorf("standalone emote was not replaced: %s", got)
	}
}

func TestBuildQuotesResponseMissingDirectory(t *testing.T) {
	response := buildQuotesResponse(filepath.Join(t.TempDir(), "nope"), "", time.Now())
	if response.Total != 0 || len(response.Channels) != 0 {
		t.Errorf("response = %#v, want empty", response)
	}
	if response.Channels == nil {
		t.Error("channels should be an empty slice, not nil")
	}
}

func TestBuildQuotesResponseSkipsBrokenFiles(t *testing.T) {
	root := t.TempDir()
	quotesDir := filepath.Join(root, "quotes")
	writeTestFile(t, quotesDir, "good.json", `[{"text":"ok","added_by":"a","created_at":"2026-03-01"}]`)
	writeTestFile(t, quotesDir, "broken.json", `[{`)

	response := buildQuotesResponse(quotesDir, "", time.Now())
	if response.Total != 1 || len(response.Channels) != 1 {
		t.Fatalf("response = %#v, want one good channel", response)
	}
	if response.Channels[0].Name != "good" {
		t.Errorf("channel = %q, want good", response.Channels[0].Name)
	}
}

func TestRenderQuoteHTMLMatchesLongestEmote(t *testing.T) {
	emotes := []services.SevenTVEmote{
		{Name: "BLA", ID: "short"},
		{Name: "BLALOO", ID: "long"},
	}
	got := renderQuoteHTML("BLALOO", emotes)
	if !strings.Contains(got, "/long/") {
		t.Errorf("longest emote was not used: %s", got)
	}
}
