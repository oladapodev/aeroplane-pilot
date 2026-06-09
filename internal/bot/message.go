package bot

type Message struct {
	Text string
}

type Sender interface {
	Send(text string) error
}

type Handler interface {
	Handle(message Message) string
}
