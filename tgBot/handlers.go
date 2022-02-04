package tgBot

import (
	"context"

	"github.com/CookieNyanCloud/tg-connection-base/repo"
	"github.com/CookieNyanCloud/tg-connection-base/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	Starting() error
	ReplyToMsg(chatId int, txt string) error
	SendMsg(chatId int64, msgId int) error
	SendAll(txt string) error
	SaveMsg() error
	//CheckBanUser(chatId int64) bool
	//BanUser(msgId int)
}

func (h Handler) Starting() error {
	panic("implement me")
}
func (h Handler) ReplyToMsg(chatId int, txt string) error {
	panic("implement me")
}

func (h Handler) SendMsg(chatId int64, msgId int) error {
	panic("implement me")
}

func (h Handler) SendAll(txt string) error {
	panic("implement me")
}

func (h Handler) SaveMsg() error {
	panic("implement me")
}
