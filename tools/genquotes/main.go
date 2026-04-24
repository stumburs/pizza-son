package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
)

type QuoteEntry struct {
	Text      string `json:"text"`
	AddedBy   string `json:"added_by"`
	CreatedAt string `json:"created_at"`
}

type QuoteDoc struct {
	Number    int
	Text      string
	AddedBy   string
	CreatedAt string
}

type ChannelQuotes struct {
	Channel string
	Quotes  []QuoteDoc
}

type SevenTVEmote struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

func loadEmotes(channel string) []SevenTVEmote {
	path := filepath.Join("data/7tv", channel+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var emotes []SevenTVEmote
	if err := json.Unmarshal(data, &emotes); err != nil {
		return nil
	}
	return emotes
}

func replaceEmotes(text string, emotes []SevenTVEmote) string {
	for _, e := range emotes {
		imgURL := fmt.Sprintf("https://cdn.7tv.app/emote/%s/1x.webp", e.ID)
		img := fmt.Sprintf(`<img src="%s" alt="%s" title="%s" style="height:1.5em;vertical-align:middle;">`, imgURL, e.Name, e.Name)
		// Only replace whole words
		text = regexp.MustCompile(`\b`+regexp.QuoteMeta(e.Name)+`\b`).ReplaceAllString(text, img)
	}
	return text
}

var quoteTempleate = `# Quotes - {{ .Channel }}

| # | Quote | Added By | Date |
|---|-------|----------|------|
{{ range .Quotes }}| {{ .Number }} | {{ .Text }} | {{ .AddedBy }} | {{ .CreatedAt }} |
{{ end }}
`

func main() {
	quotesDir := "data/quotes"
	outputDir := "docs/quotes"

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		fmt.Println("Failed to create output dir:", err)
		os.Exit(1)
	}

	entries, err := os.ReadDir(quotesDir)
	if err != nil {
		fmt.Println("Failed to read quotes dir:", err)
		os.Exit(1)
	}

	channels := make([]ChannelQuotes, 0)

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		channel := strings.TrimSuffix(entry.Name(), ".json")

		data, err := os.ReadFile(filepath.Join(quotesDir, entry.Name()))
		if err != nil {
			fmt.Println("Failed to read:", entry.Name(), err)
			continue
		}

		var raw []QuoteEntry
		if err := json.Unmarshal(data, &raw); err != nil {
			fmt.Println("Failed to parse:", entry.Name(), err)
			continue
		}

		emotes := loadEmotes(channel)

		docs := make([]QuoteDoc, len(raw))
		for i, q := range raw {
			text := q.Text
			if len(emotes) > 0 {
				text = replaceEmotes(text, emotes)
			}
			docs[i] = QuoteDoc{
				Number:    i + 1,
				Text:      text,
				AddedBy:   q.AddedBy,
				CreatedAt: q.CreatedAt,
			}
		}

		channels = append(channels, ChannelQuotes{
			Channel: channel,
			Quotes:  docs,
		})
	}

	sort.Slice(channels, func(i, j int) bool {
		return channels[i].Channel < channels[j].Channel
	})

	tmpl, err := template.New("quotes").Parse(quoteTempleate)
	if err != nil {
		fmt.Println("Template error:", err)
		os.Exit(1)
	}

	for _, cq := range channels {
		path := filepath.Join(outputDir, cq.Channel+".md")
		f, err := os.Create(path)
		if err != nil {
			fmt.Println("Failed to create:", path, err)
			continue
		}

		if err := tmpl.Execute(f, cq); err != nil {
			fmt.Println("Failed to execute template for:", cq.Channel, err)
		}
		f.Close()
		fmt.Printf("Generated %s (%d quotes)\n", path, len(cq.Quotes))
	}
}
