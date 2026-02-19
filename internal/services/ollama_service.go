package services

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"pizza-son/internal/config"
	"strings"

	"github.com/JexSrs/go-ollama"
	"github.com/gempir/go-twitch-irc/v4"
)

type OllamaService struct {
	Client *ollama.Ollama
}

var OllamaServiceInstance *OllamaService

func NewOllamaService() {
	url, err := url.Parse("http://192.168.0.101:11434")
	if err != nil {
		panic(err)
	}

	OllamaServiceInstance = &OllamaService{
		Client: ollama.New(*url),
	}
}

func (s *OllamaService) GetPromptByCommand(command string) string {
	command = strings.TrimPrefix(command, "!")

	promptFile := filepath.Join("prompts", command+".txt")

	content, err := os.ReadFile(promptFile)
	if err != nil {
		log.Printf("Failed to load prompt for command %s: %v", command, err)
		// Fallback to default prompt
		content, _ = os.ReadFile("prompts/llm.txt")
	}

	return string(content)
}

func ExtractCommand(message string) string {
	parts := strings.Fields(message)
	if len(parts) > 0 && strings.HasPrefix(parts[0], "!") {
		return parts[0]
	}
	return ""
}

func (s *OllamaService) GenerateResponse(prompt string) string {
	res, err := s.Client.Generate(
		s.Client.Generate.WithModel("mistral:latest"),
		s.Client.Generate.WithPrompt(prompt),
	)
	if err != nil {
		panic(err)
	}

	return res.Response
}

func (s *OllamaService) OnPrivateMessage(msg twitch.PrivateMessage) {
	chat := s.Client.GetChat(msg.Channel)
	// Create a new chat instance if it doesn't exist yet on this channel
	if chat == nil {
		newChat := newChat(msg.Channel)
		s.Client.PreloadChat(newChat)
		chat = s.Client.GetChat(msg.Channel)
		log.Println("Preloaded chat in", msg.Channel)
	}

	role := "user"
	content := fmt.Sprintf("%s chatted: %s", msg.User.DisplayName, msg.Message)
	log.Println(content)

	message := ollama.Message{
		Role:    &role,
		Content: &content,
		Images:  nil,
	}
	chat.AddMessage(message)
	log.Println("Added message '", content, "' to channel", msg.Channel)
}

func (s *OllamaService) GenerateChatResponse(msg twitch.PrivateMessage, prompt string) (*ollama.ChatResponse, error) {
	chatID := msg.Channel

	// Get command from original message and load corresponding system prompt FIRST
	command := ExtractCommand(msg.Message)
	systemPrompt := s.GetPromptByCommand(command)

	// Update the system prompt in the chat before generating response
	chat := s.Client.GetChat(chatID)
	if chat == nil {
		newChat := newChat(chatID)
		s.Client.PreloadChat(newChat)
		chat = s.Client.GetChat(chatID)
	}

	// Set system prompt
	readdSystemPrompt(chat, systemPrompt)

	role := "user"
	prompt = fmt.Sprintf("%s asked: %s", msg.User.DisplayName, prompt)

	message := ollama.Message{
		Role:    &role,
		Content: &prompt,
		Images:  nil,
	}

	res, err := s.Client.Chat(
		&chatID,
		s.Client.Chat.WithModel("mistral:latest"),
		s.Client.Chat.WithMessage(message),
		s.Client.Chat.WithOptions(ollama.Options{
			NumPredict: &config.Get().Ollama.NumPredict,
		}),
	)
	if err != nil {
		return &ollama.ChatResponse{}, err
	}

	return res, nil
}

func (s *OllamaService) Lobotomize(channel string) {
	s.Client.DeleteChat(channel)
}

func newChat(channel string) ollama.Chat {
	newChat := ollama.Chat{
		ID:       channel,
		Messages: []ollama.Message{},
	}
	role := "system"
	content := "You are a Twitch chat bot."
	systemPrompt := ollama.Message{
		Role:    &role,
		Content: &content,
		Images:  nil,
	}
	newChat.Messages = append(newChat.Messages, systemPrompt)

	return newChat
}

// Clears and adds new system prompt to the beginning
func readdSystemPrompt(chat *ollama.Chat, prompt string) {
	chat.DeleteMessage(0)
	role := "system"
	prompt = FillPlaceholders(prompt)
	promptMessage := &ollama.Message{
		Role:    &role,
		Content: &prompt,
		Images:  nil,
	}
	chat.AddMessageTo(0, *promptMessage)
}
