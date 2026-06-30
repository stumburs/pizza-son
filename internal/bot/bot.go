package bot

import (
	"log"
	"pizza-son/internal/config"
	"pizza-son/internal/models"
	"pizza-son/internal/services"
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
)

type Bot struct {
	client   *twitch.Client
	registry *Registry
	channels []string
}

type TwitchSender struct {
	bot *Bot
}

func (t *TwitchSender) Say(channel, message string) {
	t.bot.client.Say(channel, message)
}

func (t *TwitchSender) Reply(channel, msgID, message string) {
	t.bot.client.Reply(channel, msgID, message)
}

func New(username string, channels []string, registry *Registry) *Bot {
	return &Bot{
		registry: registry,
		channels: channels,
	}
}

func (b *Bot) Start() error {
	token := "oauth:" + services.TwitchServiceInstance.GetAccessToken()
	b.client = twitch.NewClient(config.Get().Twitch.User, token)
	b.setupHandlers()
	log.Println("[Bot] Bot connecting to:", b.channels)
	return b.client.Connect()
}

func (b *Bot) Reconnect(newToken string) {
	log.Println("[Bot] Token refreshed, disconnecting...")
	b.client.Disconnect()
	// Start() loop will reconnect automatically
}

func (b *Bot) setupHandlers() {
	b.client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		msg := twitchMessageToMessage(message)
		b.registry.Dispatch(&TwitchSender{bot: b}, msg)
	})

	// Auto raid shoutout (disabled for now to not take human jobs)
	// b.client.OnUserNoticeMessage(func(message twitch.UserNoticeMessage) {
	// 	if message.MsgID == "raid" {
	// 		raiderUsername := message.MsgParams["msg-param-login"]
	// 		if raiderUsername == "" {
	// 			raiderUsername = message.User.Name // fallback
	// 		}

	// 		if raiderUsername != "" {
	// 			log.Printf("[Bot] Channel #%s was raided by %s! Shouting out...", message.Channel, raiderUsername)

	// 			go services.TwitchServiceInstance.Shoutout(message.Channel, raiderUsername)
	// 		}
	// 	}
	// })

	b.client.OnConnect(func() {
		log.Println("[Bot] Connected")
		for _, ch := range b.channels {
			b.client.Join(ch)
		}
	})
}

func twitchMessageToMessage(m twitch.PrivateMessage) models.Message {
	msg := models.Message{
		ID:       m.ID,
		Channel:  m.Channel,
		Platform: models.PlatformTwitch,
		Text:     m.Message,
		User: models.MessageUser{
			ID:            m.User.ID,
			Name:          m.User.Name,
			DisplayName:   m.User.DisplayName,
			IsBroadcaster: m.User.IsBroadcaster,
			IsMod:         m.User.IsMod,
			IsSubscriber:  m.User.Badges["subscriber"] > 0,
		},
		FirstMessage: m.FirstMessage,
	}
	if m.Reply != nil {
		cleanBody := strings.ReplaceAll(m.Reply.ParentMsgBody, "\\s", " ")

		msg.Reply = &models.ParentMessage{
			ParentMsgID:       m.Reply.ParentMsgID,
			ParentMsgBody:     cleanBody,
			ParentDisplayName: m.Reply.ParentDisplayName,
		}
	}

	return msg
}
