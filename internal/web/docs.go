package web

import (
	"encoding/json"
	"net/http"
	"pizza-son/internal/bot"
	"pizza-son/internal/commands"
	"sort"
	"time"
)

var docsCategoryOrder = []string{
	"AI",
	"Currency",
	"Fun",
	"Games",
	"Moderation",
	"Quotes",
	"Utility",
	"Uncategorized",
}

type DocsExample struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type DocsCommand struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Usage       string        `json:"usage"`
	Permission  string        `json:"permission"`
	Cooldown    string        `json:"cooldown,omitempty"`
	Examples    []DocsExample `json:"examples,omitempty"`
}

type DocsCategory struct {
	Name     string        `json:"name"`
	Commands []DocsCommand `json:"commands"`
}

type DocsListener struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Permission  string `json:"permission"`
	Cooldown    string `json:"cooldown,omitempty"`
	Passive     bool   `json:"passive,omitempty"`
}

type DocsPermission struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type DocsResponse struct {
	GeneratedAt      time.Time        `json:"generated_at"`
	Prefix           string           `json:"prefix"`
	TotalCommands    int              `json:"total_commands"`
	TotalListeners   int              `json:"total_listeners"`
	Commands         []DocsCategory   `json:"commands"`
	Listeners        []DocsListener   `json:"listeners"`
	PermissionLevels []DocsPermission `json:"permission_levels"`
}

// GET /api/docs
//
// Documentation is built from the same registry the bot dispatches from, so
// command and listener changes are reflected without regenerating markdown.
func (ws *WebService) handleDocs(w http.ResponseWriter, r *http.Request) {
	response := buildDocsResponse(commands.GetRegistry(), time.Now())

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(response)
}

func buildDocsResponse(registry *bot.Registry, now time.Time) DocsResponse {
	response := DocsResponse{
		GeneratedAt:      now,
		Commands:         []DocsCategory{},
		Listeners:        []DocsListener{},
		PermissionLevels: docsPermissionLevels(),
	}
	if registry == nil {
		return response
	}

	response.Prefix = registry.Prefix

	grouped := make(map[string][]DocsCommand)
	for _, cmd := range registry.Commands() {
		category := cmd.Category.String()
		if category == "" {
			category = "Uncategorized"
		}

		examples := make([]DocsExample, 0, len(cmd.Examples))
		for _, example := range cmd.Examples {
			examples = append(examples, DocsExample{
				Input:  example.Input,
				Output: example.Output,
			})
		}

		grouped[category] = append(grouped[category], DocsCommand{
			Name:        cmd.Name,
			Description: cmd.Description,
			Usage:       cmd.Usage,
			Permission:  bot.PermissionName(cmd.Permission),
			Cooldown:    formatDocsCooldown(cmd.Cooldown),
			Examples:    examples,
		})
	}

	for _, category := range docsCategoryOrder {
		commands := grouped[category]
		if len(commands) == 0 {
			continue
		}
		sort.Slice(commands, func(i, j int) bool {
			return commands[i].Name < commands[j].Name
		})
		response.Commands = append(response.Commands, DocsCategory{
			Name:     category,
			Commands: commands,
		})
		response.TotalCommands += len(commands)
	}

	for _, listener := range registry.Listeners() {
		response.Listeners = append(response.Listeners, DocsListener{
			Name:        listener.Name,
			Description: listener.Description,
			Permission:  bot.PermissionName(listener.Permission),
			Cooldown:    formatDocsCooldown(listener.Cooldown),
			Passive:     listener.Passive,
		})
	}
	sort.Slice(response.Listeners, func(i, j int) bool {
		return response.Listeners[i].Name < response.Listeners[j].Name
	})
	response.TotalListeners = len(response.Listeners)

	return response
}

func formatDocsCooldown(cooldown time.Duration) string {
	if cooldown <= 0 {
		return ""
	}
	return cooldown.String()
}

func docsPermissionLevels() []DocsPermission {
	return []DocsPermission{
		{Name: "Everyone", Description: "All viewers"},
		{Name: "Subscriber", Description: "Channel subscribers and above"},
		{Name: "VIP", Description: "Channel VIPs and above"},
		{Name: "Supporter", Description: "Ko-fi supporters"},
		{Name: "Moderator", Description: "Channel moderators and above"},
		{Name: "Bot Moderator", Description: "Bot-specific moderators and above"},
		{Name: "Streamer", Description: "The broadcaster only"},
	}
}
