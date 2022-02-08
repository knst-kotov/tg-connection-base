package pkg

import (
	"log"

	"github.com/pkg/errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func StartBot(token string) (*tgbotapi.BotAPI, tgbotapi.UpdatesChannel, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return &tgbotapi.BotAPI{}, nil, errors.Wrap(err, "StartBot")
	}
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 600
	updates := bot.GetUpdatesChan(u)
	return bot, updates, nil
}
