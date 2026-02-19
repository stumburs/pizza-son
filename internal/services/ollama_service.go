package services

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/JexSrs/go-ollama"
)

type OllamaService struct {
	client     *ollama.Ollama
	model      string
	numPredict int
	mutex      sync.RWMutex
}

var Ollama *OllamaService

func InitOllamaService(host url.URL, model string, numPredict int) {
	Ollama = &OllamaService{
		client:     ollama.New(host),
		model:      model,
		numPredict: numPredict,
	}
}

// helper to get pointer of a string
func strPtr(s string) *string {
	return &s
}

func (s *OllamaService) AddMessage(channel, user, message string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	chat := s.client.GetChat(channel)
	if chat == nil {
		chat = &ollama.Chat{
			ID:       channel,
			Messages: []ollama.Message{},
		}
		s.client.PreloadChat(*chat)
		chat = s.client.GetChat(channel)
	}

	content := fmt.Sprintf("%s: %s", user, message)
	chat.AddMessage(ollama.Message{
		Role:    strPtr("user"),
		Content: strPtr(content),
		Images:  nil,
	})
}

func (s *OllamaService) GenerateResponse(channel, userPrompt string) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	chat := s.client.GetChat(channel)
	if chat == nil {
		newChat := ollama.Chat{
			ID:       channel,
			Messages: []ollama.Message{},
		}
		s.client.PreloadChat(newChat)
		chat = s.client.GetChat(channel)
	}

	chat.AddMessage(ollama.Message{
		Role:    strPtr("user"),
		Content: strPtr(userPrompt),
		Images:  nil,
	})

	chatId := channel
	res, err := s.client.Chat(strPtr(chatId),
		s.client.Chat.WithModel(s.model),
	)
	if err != nil {
		return "", fmt.Errorf("Ollama error: %w", err)
	}

	return *res.Message.Content, nil
}

func (s *OllamaService) ClearChannelHistory(channel string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.client.DeleteChat(channel)
}
