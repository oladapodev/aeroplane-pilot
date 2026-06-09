package bot

import "testing"

type fakeHandler struct {
	last Message
	resp string
}

func (f *fakeHandler) Handle(message Message) string {
	f.last = message
	return f.resp
}

func TestRelayMessage(t *testing.T) {
	handler := &fakeHandler{resp: "ok"}
	got := Relay(handler, Message{Text: "/status"})
	if got != "ok" {
		t.Fatalf("Relay() = %q, want ok", got)
	}
	if handler.last.Text != "/status" {
		t.Fatalf("handler.last.Text = %q, want /status", handler.last.Text)
	}
}
