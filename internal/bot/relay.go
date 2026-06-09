package bot

func Relay(handler Handler, message Message) string {
	return handler.Handle(message)
}
