package bot

func HandleTelegramText(handler Handler, text string) string {
	return handler.Handle(Message{Text: text})
}
