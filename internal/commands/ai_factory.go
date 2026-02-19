package commands

import (
	"pizza-son/internal/services"
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
)

type AIPromptCommand struct {
	PromptName string
	Perm       Permission
}

func (c *AIPromptCommand) Name() string {
	return strings.ToLower(c.PromptName)
}

func (c *AIPromptCommand) Permission() Permission {
	return c.Perm
}

func (c *AIPromptCommand) Execute(ctx *Context, msg twitch.PrivateMessage, args string) {
	if strings.TrimSpace(args) == "" {
		ctx.Reply(msg.Channel, msg.ID, "Please provide a prompt.")
		return
	}

	services.Ollama.AddMessage(msg.Channel, msg.User.DisplayName, args)

	response, err := services.Ollama.GenerateResponse(msg.Channel, args)
	if err != nil {
		ctx.Reply(msg.Channel, msg.ID, "AI service error"+err.Error())
		return
	}

	ctx.Reply(msg.Channel, msg.ID, response)
}

func NewAIPromptCommand(promptName string, perm Permission) Command {

	return &AIPromptCommand{
		PromptName: promptName,
		Perm:       perm,
	}
}
