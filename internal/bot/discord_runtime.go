package bot

import (
	"context"
	"errors"

	"github.com/bwmarrin/discordgo"
)

func handleDiscordMessage(session *discordgo.Session, handler Handler, message *discordgo.MessageCreate, send func(channelID, text string) error) error {
	if message == nil || message.Message == nil {
		return nil
	}
	if message.Author != nil && message.Author.Bot {
		return nil
	}
	reply := handler.Handle(Message{Text: message.Content})
	return send(message.ChannelID, reply)
}

func StartDiscord(ctx context.Context, token string, handler Handler) (*discordgo.Session, error) {
	if token == "" {
		return nil, errors.New("missing discord token")
	}
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}
	session.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		_ = handleDiscordMessage(s, handler, m, func(channelID, text string) error {
			_, err := s.ChannelMessageSend(channelID, text)
			return err
		})
	})
	if err := session.Open(); err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		_ = session.Close()
	}()
	return session, nil
}
