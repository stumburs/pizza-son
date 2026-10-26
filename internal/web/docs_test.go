package web

import (
	"pizza-son/internal/bot"
	"testing"
	"time"
)

func TestBuildDocsResponse(t *testing.T) {
	now := time.Date(2026, 10, 26, 20, 0, 0, 0, time.UTC)
	registry := bot.NewRegistry("!")

	registry.Register(bot.Command{
		Name:        "zeta",
		Description: "Fun command",
		Usage:       "!zeta",
		Category:    bot.CategoryFun,
		Permission:  bot.All,
	})
	registry.Register(bot.Command{
		Name:        "alpha",
		Description: "AI command",
		Usage:       "!alpha <text>",
		Category:    bot.CategoryAI,
		Permission:  bot.Supporter,
		Cooldown:    5 * time.Second,
		Examples: []bot.CommandExample{
			{Input: "!alpha hi", Output: "hello"},
		},
	})
	registry.RegisterListener(bot.ListenerEntry{
		Name:        "zlistener",
		Description: "Z listener",
		Permission:  bot.All,
	})
	registry.RegisterListener(bot.ListenerEntry{
		Name:        "alistener",
		Description: "A listener",
		Permission:  bot.Moderator,
		Cooldown:    time.Minute,
		Passive:     true,
	})

	response := buildDocsResponse(registry, now)

	if response.Prefix != "!" {
		t.Errorf("prefix = %q, want !", response.Prefix)
	}
	if response.TotalCommands != 2 {
		t.Errorf("total commands = %d, want 2", response.TotalCommands)
	}
	if response.TotalListeners != 2 {
		t.Errorf("total listeners = %d, want 2", response.TotalListeners)
	}
	if len(response.Commands) != 2 {
		t.Fatalf("categories = %d, want 2", len(response.Commands))
	}
	if response.Commands[0].Name != "AI" || response.Commands[1].Name != "Fun" {
		t.Errorf("category order = %s, %s; want AI, Fun", response.Commands[0].Name, response.Commands[1].Name)
	}
	alpha := response.Commands[0].Commands[0]
	if alpha.Name != "alpha" {
		t.Errorf("command = %q, want alpha", alpha.Name)
	}
	if alpha.Permission != "Supporter" {
		t.Errorf("permission = %q, want Supporter", alpha.Permission)
	}
	if alpha.Cooldown != "5s" {
		t.Errorf("cooldown = %q, want 5s", alpha.Cooldown)
	}
	if len(alpha.Examples) != 1 || alpha.Examples[0].Input != "!alpha hi" {
		t.Errorf("examples = %#v", alpha.Examples)
	}

	if response.Listeners[0].Name != "alistener" || response.Listeners[1].Name != "zlistener" {
		t.Errorf("listener order = %s, %s", response.Listeners[0].Name, response.Listeners[1].Name)
	}
	if !response.Listeners[0].Passive {
		t.Error("alistener should be passive")
	}
	if len(response.PermissionLevels) == 0 {
		t.Error("permission levels should not be empty")
	}
}

func TestBuildDocsResponseWithoutRegistry(t *testing.T) {
	response := buildDocsResponse(nil, time.Now())
	if response.TotalCommands != 0 || response.TotalListeners != 0 {
		t.Errorf("totals = %d, %d; want zero", response.TotalCommands, response.TotalListeners)
	}
	if response.Commands == nil || response.Listeners == nil {
		t.Error("commands and listeners should be empty slices, not nil")
	}
}
