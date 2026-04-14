package commands

import (
	"pizza-son/internal/bot"
	"pizza-son/internal/config"
	"pizza-son/internal/models"
	"pizza-son/internal/services"
	"strings"
	"sync"
	"time"
)

const (
	earnAmount   = 5
	earnCooldown = 5 * time.Minute
)

var (
	lastEarned   = make(map[string]time.Time)
	lastEarnedMu sync.Mutex
)

func init() {
	RegisterListener(bot.ListenerEntry{
		Name:        "currency",
		Description: "Awards slices to users for chatting",
		Handler: func(ctx bot.CommandContext) bool {
			// Exclude Discord
			if ctx.Message.Platform == models.PlatformDiscord {
				return false
			}
			if strings.HasPrefix(ctx.Message.Text, config.Get().Bot.Prefix) {
				return false
			}
			if ctx.Message.User.ID == "" {
				return false
			}

			lastEarnedMu.Lock()
			defer lastEarnedMu.Unlock()

			last, ok := lastEarned[ctx.Message.User.ID]
			if ok && time.Since(last) < earnCooldown {
				return false
			}

			lastEarned[ctx.Message.User.ID] = time.Now()
			services.CurrencyServiceInstance.Add(ctx.Message.User.ID, earnAmount)
			return false
		},
	})
}
