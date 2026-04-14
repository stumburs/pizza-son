package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pizza-son/internal/bot"
	"pizza-son/internal/config"
	"pizza-son/internal/models"
	"strings"
)

func init() {
	Register(bot.Command{
		Name:        "afterstream",
		Description: "Twitch-only: Sends a message to the Discord after-stream channel.",
		Usage:       "!afterstream <text>",
		Category:    bot.CategoryModeration,
		Permission:  bot.Moderator,
		Handler: func(ctx bot.CommandContext) {
			// Twitch exclusive
			if ctx.Message.Platform != models.PlatformTwitch {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "This is a Twitch exclusive command.")
				return
			}

			if len(ctx.Args) == 0 {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Usage: !afterstream <text>")
				return
			}

			var targetLink config.StreamerLink
			found := false
			for _, link := range config.Get().Discord.Links {
				if strings.EqualFold(link.TwitchChannel, ctx.Message.Channel) {
					targetLink = link
					found = true
					break
				}
			}

			if !found || targetLink.DiscordWebhook == "" {
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "No Discord webhook configured for this channel.")
				return
			}

			content := strings.Join(ctx.Args, " ")
			ping := ""
			if targetLink.DiscordUserID != "" {
				ping = fmt.Sprintf("<@%s> ", targetLink.DiscordUserID)
			}

			payload := map[string]string{
				"content": fmt.Sprintf("%s**After-Stream Note from %s**\n%s", ping, ctx.Message.User.DisplayName, content),
			}
			jsonPayload, _ := json.Marshal(payload)

			// --- 3. SEND WEBHOOK ---
			go func() {
				resp, err := http.Post(targetLink.DiscordWebhook, "application/json", bytes.NewBuffer(jsonPayload))
				if err != nil {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Something went wrong: "+err.Error())
					log.Printf("[AfterStream] Error sending webhook: %v", err)
					return
				}
				defer resp.Body.Close()
			}()

			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "After-stream note sent to Discord! meow")
		},
	})
}
