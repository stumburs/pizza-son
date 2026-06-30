package commands

import (
	"encoding/json"
	"log"
	"os"
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
	"time"
)

type promptMeta struct {
	Description string               `json:"description"`
	Usage       string               `json:"usage"`
	Permission  string               `json:"permission"`
	Cooldown    string               `json:"cooldown"`
	Category    string               `json:"category"`
	MarkovInput bool                 `json:"markov_input"`
	Hidden      bool                 `json:"hidden"`
	Examples    []bot.CommandExample `json:"examples"`
}

func init() {
	entries, err := os.ReadDir("prompts")
	if err != nil {
		log.Println("[ollama_prompts] Failed to read prompts directory:", err)
		return
	}

	metaData := loadPromptMeta()

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".txt") || entry.Name() == "prompts.json" {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".txt")
		meta, ok := metaData[name]
		if ok && meta.Hidden {
			continue
		}
		if !ok {
			meta = promptMeta{}
		}

		cmd := buildCommand(name, meta)
		Register(cmd)
	}
}

func loadPromptMeta() map[string]promptMeta {
	data, err := os.ReadFile("prompts/prompts.json")
	if err != nil {
		log.Println("[ollama_prompts] No prompts.json found, using defaults")
		return nil
	}
	var meta map[string]promptMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		log.Println("[ollama_prompts] Failed to parse prompts.json:", err)
		return nil
	}
	return meta
}

func buildCommand(name string, meta promptMeta) bot.Command {
	desc := meta.Description
	if desc == "" {
		desc = "Responds as " + name + "."
	}

	usage := meta.Usage
	if usage == "" {
		usage = "!" + name + " <text>"
	}

	perm := parsePermission(meta.Permission)
	cat := parseCategory(meta.Category)

	var cooldown time.Duration
	if meta.Cooldown != "" {
		cooldown, _ = time.ParseDuration(meta.Cooldown)
	}

	isMarkovInput := meta.MarkovInput
	cmdName := name

	return bot.Command{
		Name:        cmdName,
		Description: desc,
		Usage:       usage,
		Permission:  perm,
		Category:    cat,
		Cooldown:    cooldown,
		Examples:    meta.Examples,
		Handler: func(ctx bot.CommandContext) {
			var prompt string
			if isMarkovInput {
				prompt = services.MarkovServiceInstance.Generate(ctx.Message.Channel)
			} else {
				prompt = strings.Join(ctx.Args, " ")
			}

			res, err := services.OllamaServiceInstance.GenerateChatResponse(ctx.Message, services.WithPrompt(prompt), services.WithCommand(cmdName))
			if err != nil {
				log.Println("[Ollama]", cmdName, err)
				return
			}
			content := strings.Join(strings.Fields(*res.Message.Content), " ")
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, content)
		},
	}
}

func parsePermission(s string) bot.Permission {
	switch strings.ToLower(s) {
	case "subscriber":
		return bot.Subscriber
	case "vip":
		return bot.VIP
	case "supporter":
		return bot.Supporter
	case "moderator":
		return bot.Moderator
	case "botmoderator", "bot_moderator", "bot moderator":
		return bot.BotModerator
	case "streamer":
		return bot.Streamer
	default:
		return bot.All
	}
}

func parseCategory(s string) bot.Category {
	switch strings.ToLower(s) {
	case "currency":
		return bot.CategoryCurrency
	case "fun":
		return bot.CategoryFun
	case "games":
		return bot.CategoryGames
	case "moderation":
		return bot.CategoryModeration
	case "quotes":
		return bot.CategoryQuotes
	case "utility":
		return bot.CategoryUtility
	case "uncategorized":
		return bot.CategoryUncategorized
	default:
		return bot.CategoryAI
	}
}
