package commands

import (
	"fmt"
	"pizza-son/internal/bot"
	"strings"
	"sync"
)

var (
	ttsVoices    = map[string]string{}
	ttsVoicesMu  sync.RWMutex
	defaultVoice = "bestie"
)

func getTTSVoice(userID string) string {
	ttsVoicesMu.RLock()
	defer ttsVoicesMu.RUnlock()
	if v, ok := ttsVoices[userID]; ok {
		return v
	}
	return defaultVoice
}

type ttsSender struct {
	inner  bot.Sender
	prefix string
}

func (t *ttsSender) Say(channel, message string) {
	t.inner.Say(channel, t.prefix+": "+message)
}

func (t *ttsSender) Reply(channel, msgID, message string) {
	t.Say(channel, message)
}

func init() {
	Register(bot.Command{
		Name:        "speak",
		Description: "Run a command with the result being spoken using TTS.",
		Usage:       "!speak <command> [args]",
		Category:    bot.CategoryFun,
		Permission:  bot.VIP,
		Examples: []bot.CommandExample{
			{Input: "!speak mark", Output: "bestie: cats is thing"},
			{Input: "!speak how are you?", Output: "bestie: I'm wonderful!"},
			{Input: "!speak weather London", Output: "bestie: London: +10°C"},
		},
		Handler: func(ctx bot.CommandContext) {
			if len(ctx.Args) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !speak <command> [args]")
				return
			}
			voice := getTTSVoice(ctx.Message.User.ID)
			ctx.Client = &ttsSender{inner: ctx.Client, prefix: voice}
			ctx.Registry.DispatchCommand(ctx)
		},
	})

	Register(bot.Command{
		Name:        "setvoice",
		Description: "Set your TTS voice prefix.",
		Usage:       "!setvoice [voice]",
		Category:    bot.CategoryFun,
		Permission:  bot.VIP,
		Examples: []bot.CommandExample{
			{Input: "!setvoice chewbacca", Output: "TTS voice set to: chewbacca"},
			{Input: "!setvoice", Output: "Your TTS voice is chewbacca"},
		},
		Handler: func(ctx bot.CommandContext) {
			if len(ctx.Args) == 0 {
				voice := getTTSVoice(ctx.Message.User.ID)
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("Your TTS voice is %s", voice))
				return
			}
			voice := strings.Join(ctx.Args, " ")
			if len(voice) > 32 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Voice prefix must be 32 characters or less.")
				return
			}
			ttsVoicesMu.Lock()
			ttsVoices[ctx.Message.User.ID] = voice
			ttsVoicesMu.Unlock()
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("TTS voice set to: %s", voice))
		},
	})
}
