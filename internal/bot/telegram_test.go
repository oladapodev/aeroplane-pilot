package bot

import "testing"

type echoHandler struct{}

func (echoHandler) Handle(message Message) string {
	return "echo: " + message.Text
}

func TestHandleTelegramUpdateText(t *testing.T) {
	got := HandleTelegramText(echoHandler{}, "/status")
	if got != "echo: /status" {
		t.Fatalf("HandleTelegramText() = %q, want %q", got, "echo: /status")
	}
}
