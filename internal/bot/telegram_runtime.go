package bot

import (
	"context"

	telegram "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

func StartTelegram(ctx context.Context, token string, handler Handler) (*telegram.Bot, error) {
	b, err := telegram.New(token, telegram.WithDefaultHandler(func(ctx context.Context, b *telegram.Bot, update *models.Update) {
		if update == nil || update.Message == nil {
			return
		}
		reply := HandleTelegramText(handler, update.Message.Text)
		_, _ = b.SendMessage(ctx, &telegram.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   reply,
		})
	}))
	if err != nil {
		return nil, err
	}
	go b.Start(ctx)
	return b, nil
}
