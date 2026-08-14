package twitch

import (
	"errors"
	"fmt"
	"pizza-son/internal/config"

	"github.com/gempir/go-twitch-irc/v4"
)

type TwitchClient struct {
	client   *twitch.Client
	channels []string
	clientID string
	secret   string
	tokens   tokens
}

func New(config *config.Config) (*TwitchClient, error) {
	tok, err := loadTokens(tokensPath)
	if errors.Is(err, errTokensMissing) {
		return nil, fmt.Errorf("twitch: created %s - paste your access_token and refresh_token into it and rerun", tokensPath)
	}
	if err != nil {
		return nil, fmt.Errorf("twitch: load tokens: %w", err)
	}

	tc := &TwitchClient{
		client:   twitch.NewClient(config.Auth.Twitch.Username, "oauth:"+tok.AccessToken),
		channels: []string{"pizza_tm"},
		clientID: config.Auth.Twitch.ClientID,
		secret:   config.Auth.Twitch.ClientSecret,
		tokens:   tok,
	}

	tc.client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		fmt.Println("Twitch: " + message.Message)
	})

	return tc, nil
}

func (tc *TwitchClient) Run() error {
	tc.client.Join(tc.channels...)
	return tc.connect()
}

func (t *TwitchClient) connect() error {
	err := t.client.Connect()
	if !errors.Is(err, twitch.ErrLoginAuthenticationFailed) {
		return err
	}

	newTokens, refreshErr := refreshTokens(t.clientID, t.secret, t.tokens.RefreshToken)
	if refreshErr != nil {
		return fmt.Errorf("twitch: %w; token refresh failed: %w", err, refreshErr)
	}

	// save new tokens
	if err := saveTokens(tokensPath, newTokens); err != nil {
		return fmt.Errorf("twitch: token refresh ok, but saving tokens failed: %w", err)
	}
	t.tokens = newTokens
	t.client.SetIRCToken("oauth:" + newTokens.AccessToken)

	return t.client.Connect()
}
