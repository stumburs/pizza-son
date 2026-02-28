package main

import (
	"fmt"
	"html/template"
	"os"
	"pizza-son/internal/bot"
	"pizza-son/internal/commands"
	"sort"
	"strings"
)

var docTemplate = `# Commands

All commands use the {{ .Tick }}!help{{ .Tick }} command to see available commands.

| Command | Description | Permission |
|---------|-------------|------------|
{{ range .Commands -}}
| {{ $.Tick }}!{{ .Name }}{{ $.Tick }} | {{ .Description }} | {{ .Permission }} |
{{ end }}

## Permission Levels

| Level | Description |
|-------|-------------|
| Everyone | All viewers |
| Subscriber | Channel subscribers |
| VIP | Channel VIPs |
| Moderator | Channel moderators |
| Bot Moderator | Bot-specific moderators |
| Streamer | The broadcaster |
`

type CommandDoc struct {
	Name        string
	Description string
	Permission  string
}

type TemplateData struct {
	Commands []CommandDoc
	Tick     string
}

func main() {
	registry := bot.NewRegistry("!")
	commands.SetRegistry(registry)

	cmds := registry.Commands()

	docs := make([]CommandDoc, 0, len(cmds))
	for _, cmd := range cmds {
		docs = append(docs, CommandDoc{
			Name:        cmd.Name,
			Description: cmd.Description,
			Permission:  bot.PermissionName(cmd.Permission),
		})
	}

	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Name < docs[j].Name
	})

	tmpl, err := template.New("docs").Parse(docTemplate)
	if err != nil {
		fmt.Println("Template error:", err)
		os.Exit(1)
	}

	// Write to docs/commands.md
	if err := os.MkdirAll("docs", os.ModePerm); err != nil {
		fmt.Println("Failed to create docs dir:", err)
		os.Exit(1)
	}

	f, err := os.Create("docs/commands.md")
	if err != nil {
		fmt.Println("Failed to create commands.md", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := tmpl.Execute(f, TemplateData{
		Commands: docs,
		Tick:     "`",
	}); err != nil {
		fmt.Println("Failed to execute template:", err)
		os.Exit(1)
	}

	fmt.Printf("Generated docs/commands.md with %d commands\n", len(docs))
	_ = strings.TrimSpace
}
