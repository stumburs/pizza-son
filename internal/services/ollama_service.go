package services

import "errors"

var OllamaService = &ollamaService{}

type ollamaService struct{}

func (s *ollamaService) GetLLMResponse(prompt string) (string, error) {
	if prompt == "" {
		return "", errors.New("empty prompt")
	}

	// TODO: implement ollama service calls
	return "AI response to: " + prompt, nil
}
