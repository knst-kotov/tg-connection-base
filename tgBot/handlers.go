package tgBot

import (
	"context"

	"github.com/CookieNyanCloud/tg-connection-base/repo"
	"github.com/CookieNyanCloud/tg-connection-base/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

const (
	welcome     = "обратная связь: тест: кнопка в меню"
	feedbackTxt = "обратая связь"
	unknownTxt  = "неизвестная команда"
	saveTxt     = "обращение передано"
)

type Handler struct {
	Ctx    context.Context
	Chat   int64
	Cache  repo.ICache
	Sheets service.ISheets
	Bot    *tgbotapi.BotAPI
}

func NewHandler(
	ctx context.Context,
	chat int64,
	cache repo.ICache,
	sheets service.ISheets,
	bot *tgbotapi.BotAPI) *Handler {
	return &Handler{
		Ctx:    ctx,
		Chat:   chat,
		Cache:  cache,
		Sheets: sheets,
		Bot:    bot,
	}
}

type IHandler interface {
	Starting(id int64, name, nick string) error
	Unknown(id int64) error
	Feedback(id int64, msgId int) error
	Find(id int64) error
	//ReplyToMsg(chatId int, txt string) error
	//SendMsg(chatId int64, msgId int) error
	//SendAll(txt string) error
	//SaveMsg() error

}

func (h Handler) Starting(id int64, name, nick string) error {
	msg := tgbotapi.NewMessage(id, welcome)
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(feedbackTxt),
		),
	)
	_, err := h.Bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "Send")
	}
	err = h.Sheets.SaveContact(id, name, nick)
	if err != nil {
		return errors.Wrap(err, "SaveContact")
	}
	return nil
}

func (h Handler) Unknown(id int64) error {
	msg := tgbotapi.NewMessage(id, unknownTxt)
	//msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
	//	tgbotapi.NewKeyboardButtonRow(
	//		tgbotapi.NewKeyboardButton(feedbackTxt),
	//	),
	//)
	_, err := h.Bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "Send")
	}
	return nil
}

//надо ли хранить текст в таблице, если бот подтянет сообщение по айди?
func (h Handler) Feedback(id int64, msgId int) error {
	msg := tgbotapi.NewMessage(id, saveTxt)
	_, err := h.Bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "Send")
	}
	err = h.Sheets.SaveMsg(id, msgId)
	if err != nil {
		return errors.Wrap(err, "SaveMsg")
	}
	return nil
}

func (h Handler) Find(toId int64) error {
	fromId, msgId, err := h.Sheets.GetFirst()
	if err != nil {
		return errors.Wrap(err, "GetFirst")
	}
	forward := tgbotapi.NewForward(toId, fromId, msgId)
	_, err = h.Bot.Send(forward)
	if err != nil {
		return errors.Wrap(err, "Send")
	}
	return nil
}
