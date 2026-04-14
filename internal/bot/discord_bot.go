package bot

import (
	"log"
	"pizza-son/internal/models"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type DiscordBot struct {
	session         *discordgo.Session
	registry        *Registry
	allowedChannels map[string]bool
}

type DiscordSender struct {
	session   *discordgo.Session
	channelID string
}

func (d *DiscordSender) Say(channel, message string) {
	d.session.ChannelMessageSend(channel, message)
}

func (d *DiscordSender) Reply(channel, msgID, message string) {
	d.session.ChannelMessageSendReply(channel, message, &discordgo.MessageReference{
		MessageID: msgID,
		ChannelID: channel,
	})
}

func NewDiscordBot(token string, allowedChannelIDs []string, registry *Registry) (*DiscordBot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	allowed := make(map[string]bool, len(allowedChannelIDs))
	for _, id := range allowedChannelIDs {
		allowed[id] = true
	}

	return &DiscordBot{
		session:         session,
		registry:        registry,
		allowedChannels: allowed,
	}, nil
}

func (b *DiscordBot) Start() error {
	b.session.AddHandler(b.onMessage)
	b.session.Identify.Intents = discordgo.IntentGuildMessages |
		discordgo.IntentDirectMessages |
		discordgo.IntentGuildMembers |
		discordgo.IntentMessageContent // for !lobotomize
	if err := b.session.Open(); err != nil {
		return err
	}
	log.Println("[Discord] Connected, listening on", len(b.allowedChannels), "channel(s)")
	return nil
}

func (b *DiscordBot) Stop() {
	b.session.Close()
}

func (b *DiscordBot) onMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}
	if !b.allowedChannels[m.ChannelID] {
		return
	}

	msg := discordMessageToMessage(s, m)
	sender := &DiscordSender{session: s, channelID: m.ChannelID}
	b.registry.Dispatch(sender, msg)
}

func discordMessageToMessage(s *discordgo.Session, m *discordgo.MessageCreate) models.Message {
	author := m.Author
	displayName := author.GlobalName
	if displayName == "" {
		displayName = author.Username
	}

	isMod := false
	isBroadcaster := false

	if m.Member != nil {
		// TODO: Broadcaster check
		// Moderator check
		perms, err := s.UserChannelPermissions(m.Author.ID, m.ChannelID)
		if err == nil {
			isMod = (perms&discordgo.PermissionManageMessages != 0) ||
				(perms&discordgo.PermissionAdministrator != 0)
		}

		// Broadcaster check: (Server owner)
		guild, err := s.State.Guild(m.GuildID)
		if err == nil && guild.OwnerID == author.ID {
			isBroadcaster = true
		}
	}

	msg := models.Message{
		ID:       m.ID,
		Channel:  m.ChannelID,
		Platform: models.PlatformDiscord,
		Text:     m.Content,
		// No first-message on discord
		FirstMessage: false,
		User: models.MessageUser{
			ID:            author.ID,
			Name:          strings.ToLower(author.Username),
			DisplayName:   displayName,
			IsMod:         isMod,
			IsBroadcaster: isBroadcaster,
		},
	}
	if m.ReferencedMessage != nil {
		refAuthor := m.ReferencedMessage.Author
		refName := refAuthor.GlobalName
		if refName == "" {
			refName = refAuthor.Username
		}

		msg.Reply = &models.ParentMessage{
			ParentMsgID:       m.ReferencedMessage.ID,
			ParentMsgBody:     m.ReferencedMessage.Content,
			ParentDisplayName: refName,
		}
	}
	return msg
}

func hasDiscordModPerms(m *discordgo.Member) bool {
	return m.Permissions&discordgo.PermissionAdministrator != 0 ||
		m.Permissions&discordgo.PermissionManageMessages != 0
}

func (b *DiscordBot) SendGlobalMessage(channelID, message string) {
	b.session.ChannelMessageSend(channelID, message)
}
