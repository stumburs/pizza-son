package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

var quoteTempleate = `# Quotes — {{ .Channel }}

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

		docs := make([]QuoteDoc, len(raw))
		for i, q := range raw {
			docs[i] = QuoteDoc{
				Number:    i + 1,
				Text:      q.Text,
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
