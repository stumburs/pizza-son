package services

import (
	"fmt"
	"log"
	"net/url"
	"pizza-son/internal/config"

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

	// Remove prompt and re-append to end
	chat := s.Client.GetChat(chatID)
	if chat == nil {
		log.Println("Failed to get current chat to re-append system prompt")
	} else {
		readdSystemPrompt(chat, "TODO")
	}

	fmt.Println(chat)

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

// Clears and adds new system prompt to the end of the chat
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
