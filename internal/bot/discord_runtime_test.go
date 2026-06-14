package bot

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
)

type discordEchoHandler struct{}

func (discordEchoHandler) Handle(message Message) string {
	return "echo: " + message.Text
}

func TestHandleDiscordMessageSendsReply(t *testing.T) {
	var gotChannelID string
	var gotText string

	session := &discordgo.Session{}
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Content: "/status",
			ChannelID: "chan-1",
		},
	}

	err := handleDiscordMessage(session, discordEchoHandler{}, message, func(channelID, text string) error {
		gotChannelID = channelID
		gotText = text
		return nil
	})
	if err != nil {
		t.Fatalf("handleDiscordMessage() error = %v, want nil", err)
	}
	if gotChannelID != "chan-1" {
		t.Fatalf("channelID = %q, want chan-1", gotChannelID)
	}
	if gotText != "echo: /status" {
		t.Fatalf("text = %q, want %q", gotText, "echo: /status")
	}
}

func TestHandleDiscordMessageIgnoresBotAuthor(t *testing.T) {
	called := false

	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Content:    "ping",
			ChannelID: "chan-2",
			Author: &discordgo.User{
				ID:  "user-10",
				Bot: true,
			},
		},
	}

	err := handleDiscordMessage(&discordgo.Session{}, discordEchoHandler{}, message, func(string, string) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("handleDiscordMessage() error = %v, want nil", err)
	}
	if called {
		t.Fatal("send function was called for bot authored message")
	}
}

func TestHandleDiscordMessageReturnsSendError(t *testing.T) {
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{Content: "hello", ChannelID: "chan-2"},
	}

	err := handleDiscordMessage(&discordgo.Session{}, discordEchoHandler{}, message, func(string, string) error {
		return errors.New("send failed")
	})
	if err == nil || err.Error() != "send failed" {
		t.Fatalf("handleDiscordMessage() error = %v, want send failed", err)
	}
}

func TestHandleDiscordMessageIgnoresNilMessage(t *testing.T) {
	called := false

	err := handleDiscordMessage(&discordgo.Session{}, discordEchoHandler{}, nil, func(string, string) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("handleDiscordMessage() error = %v, want nil", err)
	}
	if called {
		t.Fatal("send function was called for nil message")
	}
}
