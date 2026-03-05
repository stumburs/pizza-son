package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"pizza-son/internal/bot"
	"pizza-son/internal/services"
	"strings"
	"time"
)

const signupFile = "data/signups.json"

type SignupEntry struct {
	UserID              string `json:"user_id"`
	Username            string `json:"username"`
	DisplayName         string `json:"display_name"`
	TwitchURL           string `json:"twitch_url"`
	SignedUpAt          string `json:"signed_up_at"`
	SignedUpFrom        string `json:"signed_up_from"`
	SignedUpFromViewers int    `json:"signed_up_from_viewers"`
	FirstMessage        bool   `json:"first_message"`
}

func loadSignups() ([]SignupEntry, error) {
	data, err := os.ReadFile(signupFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []SignupEntry{}, nil
		}
		return nil, err
	}
	var entries []SignupEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func saveSignups(entries []SignupEntry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(signupFile, data, 0644)
}

func init() {
	Register(bot.Command{
		Name:        "signup",
		Description: "Sign up to have the bot added to your channel. Your channel will be manually reviewed to determine whether it meets certain criteria. Having pizza_son added to your channel might take a day or two after passing the criteria.",
		Usage:       "!signup",
		Examples: []bot.CommandExample{
			{Input: "!signup", Output: "You have been added to the signup list, fantastic_streamer4235! Please wait for manual review."},
		},
		Handler: func(ctx bot.CommandContext) {
			entries, err := loadSignups()
			if err != nil {
				log.Println("[Signup] Failed to load signups:", err)
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Something went wrong, please try again later.")
				return
			}

			// Check for duplicates
			for _, e := range entries {
				if strings.EqualFold(e.UserID, ctx.Message.User.ID) {
					ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "You've already signed up! Please wait for manual review.")
					return
				}
			}

			info := services.TwitchServiceInstance.GetStreamInfo(ctx.Message.Channel)

			entry := SignupEntry{
				UserID:              ctx.Message.User.ID,
				Username:            ctx.Message.User.Name,
				DisplayName:         ctx.Message.User.DisplayName,
				TwitchURL:           "https://twitch.tv/" + ctx.Message.User.Name,
				SignedUpAt:          time.Now().Format(time.RFC3339),
				SignedUpFrom:        ctx.Message.Channel,
				SignedUpFromViewers: info.ViewerCount,
				FirstMessage:        ctx.Message.FirstMessage,
			}

			entries = append(entries, entry)
			if err := saveSignups(entries); err != nil {
				log.Println("[Signup] Failed to save signups:", err)
				ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, "Something went wrong, please try again later.")
				return
			}

			log.Printf("[Signup] New signup: %s (%s)", entry.DisplayName, entry.TwitchURL)
			ctx.Client.Reply(ctx.Message.Channel, ctx.Message.ID, fmt.Sprintf("You have been added to the signup list, %s! Please wait for manual review.", ctx.Message.User.DisplayName))
		},
	})
}
