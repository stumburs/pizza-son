package services

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"pizza-son/internal/config"
	"pizza-son/internal/models"
	"strings"
	"sync"

	"github.com/JexSrs/go-ollama"
)

type chatOptions struct {
	prompt  string
	command string
	model   string
	options *ollama.Options
}

type ChatOption func(*chatOptions)

func WithPrompt(p string) ChatOption {
	return func(o *chatOptions) { o.prompt = p }
}

func WithCommand(name string) ChatOption {
	return func(o *chatOptions) { o.command = name }
}

func WithModel(m string) ChatOption {
	return func(o *chatOptions) { o.model = m }
}

func WithOptions(opts *ollama.Options) ChatOption {
	return func(o *chatOptions) { o.options = opts }
}

type promptContext struct {
	template string           // raw system prompt with {{placeholders}}
	messages []ollama.Message // history for this context
}

type OllamaService struct {
	Client   *ollama.Ollama
	mu       sync.Mutex
	ambient  map[string][]ollama.Message // channel -> chat messages
	contexts map[string]*promptContext   // "channel:command" -> prompt context
}

var OllamaServiceInstance *OllamaService

func NewOllamaService() {
	u, err := url.Parse(config.Get().Ollama.Host)
	if err != nil {
		panic(err)
	}

	OllamaServiceInstance = &OllamaService{
		Client:   ollama.New(*u),
		ambient:  make(map[string][]ollama.Message),
		contexts: make(map[string]*promptContext),
	}
}

func (s *OllamaService) OnPrivateMessage(msg models.Message) {
	role := "user"
	content := fmt.Sprintf("%s chatted: %s", msg.User.DisplayName, msg.Text)
	entry := ollama.Message{Role: &role, Content: &content}

	s.mu.Lock()
	defer s.mu.Unlock()

	ambient := s.ambient[msg.Channel]
	ambient = append(ambient, entry)

	maxAmbient := 40
	if len(ambient) > maxAmbient {
		ambient = ambient[len(ambient)-maxAmbient:]
	}
	s.ambient[msg.Channel] = ambient
}

func (s *OllamaService) GenerateChatResponse(msg models.Message, opts ...ChatOption) (*ollama.ChatResponse, error) {
	cfg := chatOptions{
		model: config.Get().Ollama.Model,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	command := cfg.command
	if command == "" {
		command = ExtractCommand(msg.Text)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// build messages: [system] + [ambient] + [context history] + [current prompt]
	chatMsgs := make([]ollama.Message, 0, 64)

	// system prompt from per-command context
	contextKey := msg.Channel + ":" + command
	pc, exists := s.contexts[contextKey]
	if !exists {
		raw := s.GetPromptByCommand(command)
		pc = &promptContext{template: raw}
		s.contexts[contextKey] = pc
	}
	filled := FillPlaceholders(msg, pc.template)
	sysRole := "system"
	chatMsgs = append(chatMsgs, ollama.Message{Role: &sysRole, Content: &filled})

	// ambient context (general chat messages)
	if ambient, ok := s.ambient[msg.Channel]; ok {
		chatMsgs = append(chatMsgs, ambient...)
	}

	// command history
	chatMsgs = append(chatMsgs, pc.messages...)

	// current user prompt
	userRole := "user"
	promptText := fmt.Sprintf("%s asked: %s", msg.User.DisplayName, cfg.prompt)
	currentMsg := ollama.Message{Role: &userRole, Content: &promptText}
	chatMsgs = append(chatMsgs, currentMsg)

	// build options
	genOpts := ollama.Options{}
	if cfg.options != nil {
		genOpts = *cfg.options
	}
	if genOpts.NumPredict == nil {
		np := config.Get().Ollama.NumPredict
		genOpts.NumPredict = &np
	}

	// pass all messages directly
	builders := make([]func(*ollama.ChatRequestBuilder), 0, len(chatMsgs)+2)
	builders = append(builders, s.Client.Chat.WithModel(cfg.model))
	builders = append(builders, s.Client.Chat.WithOptions(genOpts))
	for i := range chatMsgs {
		m := chatMsgs[i]
		builders = append(builders, s.Client.Chat.WithMessage(m))
	}

	res, err := s.Client.Chat(nil, builders...)
	if err != nil {
		return &ollama.ChatResponse{}, err
	}

	// store the exchange in command context history
	pc.messages = append(pc.messages, currentMsg, res.Message)

	// trim to max history (each entry is a 2 message)
	maxHistory := config.Get().Ollama.MaxHistory
	if maxHistory <= 0 {
		maxHistory = 80
	}
	maxStored := maxHistory * 2
	if len(pc.messages) > maxStored {
		pc.messages = pc.messages[len(pc.messages)-maxStored:]
	}

	return res, nil
}

func (s *OllamaService) GenerateResponse(prompt string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	res, err := s.Client.Generate(
		s.Client.Generate.WithModel(config.Get().Ollama.Model),
		s.Client.Generate.WithPrompt(prompt),
	)
	if err != nil {
		panic(err)
	}
	return res.Response
}

func (s *OllamaService) Lobotomize(channel string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.ambient, channel)

	prefix := channel + ":"
	for key := range s.contexts {
		if strings.HasPrefix(key, prefix) {
			delete(s.contexts, key)
		}
	}
	log.Println("[Ollama] Lobotomized", channel)
}

func (s *OllamaService) GetPromptByCommand(command string) string {
	command = strings.TrimPrefix(command, "!")

	promptFile := filepath.Join("prompts", command+".txt")

	content, err := os.ReadFile(promptFile)
	if err != nil {
		log.Printf("[Ollama] Failed to load prompt for command %s: %v", command, err)
		content, _ = os.ReadFile("prompts/llm.txt")
	}

	return string(content)
}

func ExtractCommand(message string) string {
	parts := strings.Fields(message)
	if len(parts) == 0 {
		return ""
	}

	first := parts[0]

	if first == "!speak" {
		if len(parts) > 1 {
			return parts[1]
		}
		return ""
	}

	if after, ok := strings.CutPrefix(first, "!"); ok {
		return after
	}

	return ""
}
