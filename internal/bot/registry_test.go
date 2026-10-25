package bot_test

import (
	"os"
	"pizza-son/internal/bot"
	"pizza-son/internal/config"
	"pizza-son/internal/models"
	"pizza-son/internal/services"
	"testing"
	"time"
)

type testSender struct{}

func (testSender) Say(string, string)           {}
func (testSender) Reply(string, string, string) {}

func TestDispatchLogsMessagesCommandsAndListeners(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.WriteFile("config.toml", []byte("[bot]\nprefix = \"!\"\nchannels = [\"testchannel\"]\n"), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	config.Load("config.toml")

	oldActivity := services.ActivityServiceInstance
	services.NewActivityService()
	t.Cleanup(func() {
		services.ActivityServiceInstance.Close()
		services.ActivityServiceInstance = oldActivity
	})

	oldSettings := services.ChannelSettingsInstance
	services.NewChannelSettingsService()
	t.Cleanup(func() { services.ChannelSettingsInstance = oldSettings })

	registry := bot.NewRegistry("!")

	listenerDone := make(chan struct{}, 1)
	registry.RegisterListener(bot.ListenerEntry{
		Name: "test-listener",
		Matcher: func(ctx bot.CommandContext) bool {
			return true
		},
		Handler: func(ctx bot.CommandContext) bool {
			listenerDone <- struct{}{}
			return true
		},
	})
	registry.RegisterListener(bot.ListenerEntry{
		Name: "test-miss",
		Matcher: func(ctx bot.CommandContext) bool {
			return false
		},
		Handler: func(ctx bot.CommandContext) bool {
			t.Error("listener with a false matcher should not run")
			return false
		},
	})
	registry.RegisterListener(bot.ListenerEntry{
		Name: "test-noop",
		Matcher: func(ctx bot.CommandContext) bool {
			return true
		},
		Handler: func(ctx bot.CommandContext) bool {
			return false
		},
	})
	passiveDone := make(chan struct{}, 1)
	registry.RegisterListener(bot.ListenerEntry{
		Name:    "test-passive",
		Passive: true,
		Handler: func(ctx bot.CommandContext) bool {
			passiveDone <- struct{}{}
			return true
		},
	})

	commandDone := make(chan struct{}, 1)
	registry.Register(bot.Command{
		Name: "testcmd",
		Handler: func(ctx bot.CommandContext) {
			commandDone <- struct{}{}
		},
	})

	msg := models.Message{
		ID:       "msg-1",
		Channel:  "testchannel",
		Platform: models.PlatformTwitch,
		Text:     "!testcmd",
		User: models.MessageUser{
			ID:   "user-1",
			Name: "Alice",
		},
	}
	registry.Dispatch(testSender{}, msg)

	waitForSignal(t, listenerDone, "listener")
	waitForSignal(t, commandDone, "command")
	waitForSignal(t, passiveDone, "passive listener")

	deadline := time.Now().Add(2 * time.Second)
	for {
		channel, ok := services.ActivityServiceInstance.ChannelStats("testchannel")
		if ok && channel.Listeners.Total > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for listener activity")
		}
		time.Sleep(time.Millisecond)
	}

	channel, ok := services.ActivityServiceInstance.ChannelStats("testchannel")
	if !ok {
		t.Fatal("channel activity not recorded")
	}
	if channel.Messages.Total != 1 {
		t.Errorf("messages = %d, want 1", channel.Messages.Total)
	}
	if channel.Commands.Total != 1 || channel.CommandUsage["testcmd"] != 1 {
		t.Errorf("commands = %#v, usage = %#v", channel.Commands, channel.CommandUsage)
	}
	if channel.Listeners.Total != 1 || channel.ListenerUsage["test-listener"] != 1 {
		t.Errorf("listeners = %#v, usage = %#v", channel.Listeners, channel.ListenerUsage)
	}
	if _, ok := channel.ListenerUsage["test-miss"]; ok {
		t.Error("listener that did not match was logged")
	}
	if _, ok := channel.ListenerUsage["test-noop"]; ok {
		t.Error("listener that returned false was logged")
	}
	if _, ok := channel.ListenerUsage["test-passive"]; ok {
		t.Error("passive listener was logged")
	}

	user, ok := services.ActivityServiceInstance.UserStats("alice")
	if !ok {
		t.Fatal("user activity not recorded")
	}
	if user.Messages != 1 || user.Commands != 1 || user.ListenerActivations != 1 {
		t.Errorf("user activity = %#v", user)
	}
}

func waitForSignal(t *testing.T, ch <-chan struct{}, name string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %s", name)
	}
}
