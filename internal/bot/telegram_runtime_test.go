package bot

import (
	"errors"
	"testing"

	"github.com/go-telegram/bot/models"
)

type telegramEchoHandler struct{}

func (telegramEchoHandler) Handle(message Message) string {
	return "echo: " + message.Text
}

func TestHandleTelegramUpdateSendsReply(t *testing.T) {
	var gotChatID int64
	var gotText string

	handler := telegramEchoHandler{}
	update := &models.Update{
		Message: &models.Message{
			Text: "/status",
			Chat: models.Chat{ID: 42},
		},
	}

	err := handleTelegramUpdate(handler, update, func(chatID int64, text string) error {
		gotChatID = chatID
		gotText = text
		return nil
	})
	if err != nil {
		t.Fatalf("handleTelegramUpdate() error = %v, want nil", err)
	}
	if gotChatID != 42 {
		t.Fatalf("chatID = %d, want 42", gotChatID)
	}
	if gotText != "echo: /status" {
		t.Fatalf("text = %q, want %q", gotText, "echo: /status")
	}
}

func TestHandleTelegramUpdateReturnsSendError(t *testing.T) {
	update := &models.Update{
		Message: &models.Message{
			Text: "hello",
			Chat: models.Chat{ID: 7},
		},
	}

	err := handleTelegramUpdate(telegramEchoHandler{}, update, func(int64, string) error {
		return errors.New("send failed")
	})
	if err == nil || err.Error() != "send failed" {
		t.Fatalf("handleTelegramUpdate() error = %v, want send failed", err)
	}
}

func TestHandleTelegramUpdateIgnoresEmptyUpdate(t *testing.T) {
	called := false

	err := handleTelegramUpdate(telegramEchoHandler{}, nil, func(int64, string) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("handleTelegramUpdate() error = %v, want nil", err)
	}
	if called {
		t.Fatal("send function was called for nil update")
	}
}
