package main

import (
	"fmt"
	"os"
	"pizza-son/internal/bot"
	"pizza-son/internal/commands"
	"sort"
	"strings"
	"text/template"
)

var categoryOrder = []string{"AI", "Currency", "Fun", "Games", "Moderation", "Quotes", "Utility", "Uncategorized"}

var categoryIcons = map[string]string{
	"AI":            ":material-robot:",
	"Currency":      ":material-pizza:",
	"Fun":           ":material-emoticon:",
	"Games":         ":material-gamepad-variant:",
	"Moderation":    ":material-shield:",
	"Quotes":        ":material-format-quote-close:",
	"Utility":       ":material-wrench:",
	"Uncategorized": ":material-help:",
}

var docTemplate = `# Commands

Use the search bar above to find a specific command.

!!! info "Usage notation"
    - {{ $.Tick }}<argument>{{ $.Tick }} — required
    - {{ $.Tick }}[argument]{{ $.Tick }} — optional


<div class="grid cards" markdown>

{{ range .Categories -}}
-   {{ .Icon }} __{{ .Name }}__

    ---

	{{ len .Commands }} commands

	[:octicons-arrow-right-24: {{ .Name }}](#{{ .Name | lower }})

{{ end }}
</div>

---

{{ range .Categories }}
## {{ .Name }}

{{ range .Commands }}
### !{{ .Name }}

{{ .Description }}

{{ if .Usage -}}
**Usage:** {{ $.Tick }}{{ .Usage }}{{ $.Tick }}
{{- end }}

**Permission:** {{ .Permission }}

{{ if .Examples -}}
<details markdown>
<summary>Examples</summary>

| Input | Output |
|-------|--------|
{{ range .Examples -}}
| {{ $.Tick }}{{ .Input }}{{ $.Tick }} | {{ $.Tick }}{{ .Output }}{{ $.Tick }} |
{{ end }}

</details>
{{- end }}

---
{{ end }}
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

*[Everyone]: All viewers can use this command
*[Subscriber]: Channel subscribers and above
*[VIP]: Channel VIPs and above
*[Moderator]: Channel moderators and above
*[Bot Moderator]: Bot-specific moderators and above
*[Streamer]: The broadcaster only
`

type ExampleDoc struct {
	Input  string
	Output string
}

type CommandDoc struct {
	Name        string
	Description string
	Usage       string
	Permission  string
	Examples    []ExampleDoc
}

type CategoryDoc struct {
	Name     string
	Icon     string
	Commands []CommandDoc
}

type TemplateData struct {
	Categories []CategoryDoc
	Tick       string
}

var listenerTemplate = `# Listeners

Listeners run on every message passively, without requiring a command prefix.

---

{{ range .Listeners }}
## {{ .Name }}

{{ .Description }}

**Permission:** {{ .Permission }}

---
{{ end }}
`

type ListenerDoc struct {
	Name        string
	Description string
	Permission  string
}

type ListenerTemplateData struct {
	Listeners []ListenerDoc
}

func main() {
	registry := bot.NewRegistry("!")
	commands.SetRegistry(registry)

	cmds := registry.Commands()

	// Group by category
	categoryMap := make(map[string][]CommandDoc)
	for _, cmd := range cmds {
		category := cmd.Category.String()
		if category == "" {
			category = "Uncategorized"
		}
		examples := make([]ExampleDoc, 0, len(cmd.Examples))
		for _, ex := range cmd.Examples {
			examples = append(examples, ExampleDoc{
				Input:  ex.Input,
				Output: ex.Output,
			})
		}
		categoryMap[category] = append(categoryMap[category], CommandDoc{
			Name:        cmd.Name,
			Description: cmd.Description,
			Usage:       cmd.Usage,
			Permission:  bot.PermissionName(cmd.Permission),
			Examples:    examples,
		})
	}

	// Build ordered categories
	categories := make([]CategoryDoc, 0)
	seen := make(map[string]bool)
	for _, name := range categoryOrder {
		if cmds, ok := categoryMap[name]; ok {
			sort.Slice(cmds, func(i, j int) bool {
				return cmds[i].Name < cmds[j].Name
			})
			categories = append(categories, CategoryDoc{
				Name:     name,
				Icon:     categoryIcons[name],
				Commands: cmds,
			})
			seen[name] = true
		}
	}

	// Uncategorized
	for name, cmds := range categoryMap {
		if seen[name] {
			continue
		}
		sort.Slice(cmds, func(i, j int) bool {
			return cmds[i].Name < cmds[j].Name
		})
		categories = append(categories, CategoryDoc{Name: name, Commands: cmds})
	}

	tmpl, err := template.New("docs").Funcs(template.FuncMap{
		"lower": strings.ToLower,
	}).Parse(docTemplate)
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
		Categories: categories,
		Tick:       "`",
	}); err != nil {
		fmt.Println("Failed to execute template:", err)
		os.Exit(1)
	}

	fmt.Printf("Generated docs/commands.md with %d categories\n", len(categories))

	// Listeners
	listenerDocs := make([]ListenerDoc, 0)
	for _, l := range registry.Listeners() {
		listenerDocs = append(listenerDocs, ListenerDoc{
			Name:        l.Name,
			Description: l.Description,
			Permission:  bot.PermissionName(l.Permission),
		})
	}
	sort.Slice(listenerDocs, func(i, j int) bool {
		return listenerDocs[i].Name < listenerDocs[j].Name
	})

	listenerTmpl, err := template.New("listeners").Parse(listenerTemplate)
	if err != nil {
		fmt.Println("Listener template error:", err)
		os.Exit(1)
	}

	lf, err := os.Create("docs/listeners.md")
	if err != nil {
		fmt.Println("Failed to create listeners.md", err)
		os.Exit(1)
	}
	defer lf.Close()

	if err := listenerTmpl.Execute(lf, ListenerTemplateData{Listeners: listenerDocs}); err != nil {
		fmt.Println("Failed to execute listener template:", err)
		os.Exit(1)
	}

	fmt.Printf("Generated docs/listeners.md with %d listeners\n", len(listenerDocs))

	_ = strings.TrimSpace
}
