package bot

import (
	"context"

	telegram "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func handleTelegramUpdate(handler Handler, update *models.Update, send func(chatID int64, text string) error) error {
	if update == nil || update.Message == nil {
		return nil
	}
	reply := HandleTelegramText(handler, update.Message.Text)
	return send(update.Message.Chat.ID, reply)
}

func StartTelegram(ctx context.Context, token string, handler Handler) (*telegram.Bot, error) {
	b, err := telegram.New(token, telegram.WithDefaultHandler(func(ctx context.Context, b *telegram.Bot, update *models.Update) {
		_ = handleTelegramUpdate(handler, update, func(chatID int64, text string) error {
			_, err := b.SendMessage(ctx, &telegram.SendMessageParams{
				ChatID: chatID,
				Text:   text,
			})
			return err
		})
	}))
	if err != nil {
		return nil, err
	}
	go b.Start(ctx)
	return b, nil
}
