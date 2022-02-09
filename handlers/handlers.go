package handlers

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

const (
	welcome     = "обратная связь: тест: кнопка в меню"
	feedbackTxt = "обратая связь"
	unknownTxt  = "неизвестная команда"
	saveTxt     = "обращение передано"
)

type IStorage interface {
	ClearMsgs(id int64) error
	// admins
	LoadAdmins() (map[int64]struct{}, error)
	SaveAdmin(id int64, nick string) error
	GetLast() (int64, []int, error)
	// users
	SaveContact(id int64, name, nick string) error
	GetAll() ([]int64, error)
	SaveMsg(id int64, msgId int) error
}

type ICache interface {
	SetUser(msgId int, userId int64) error
	GetUser(msgId int) (int64, error)
	SetBan(userId int64) error
	GetBan(userId int64) (bool, error)
}

type Handler struct {
	Cache   ICache
	Storage IStorage
	Bot     *tgbotapi.BotAPI
}

func NewHandler(
	cache ICache,
	sheets IStorage,
	bot *tgbotapi.BotAPI) *Handler {
	return &Handler{
		Cache:   cache,
		Storage: sheets,
		Bot:     bot,
	}
}

type IHandler interface {
	Unknown(id int64) error
	// user
	Starting(id int64, name, nick string) error
	Feedback(id int64, msgId int) error
	// admin
	AddAdmin(id int64, nick string) error
	ReplyToMsg(msgId int, txt string) error
	SendAll(txt string) error
	Find(toId int64) error
	//	todo:?chat
}

//unknown command
func (h *Handler) Unknown(id int64) error {
	msg := tgbotapi.NewMessage(id, unknownTxt)
	_, err := h.Bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "Send")
	}
	return nil
}

//first message, save info
func (h *Handler) Starting(id int64, name, nick string) error {
	msg := tgbotapi.NewMessage(id, welcome)
	_, err := h.Bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "Send")
	}
	err = h.Storage.SaveContact(id, name, nick)
	if err != nil {
		return errors.Wrap(err, "SaveContact")
	}
	return nil
}

//save message id, answered when needed
func (h *Handler) Feedback(id int64, msgId int) error {
	err := h.Storage.SaveMsg(id, msgId)
	if err != nil {
		return errors.Wrap(err, "SaveMsg")
	}
	return nil
}

//add new admin
func (h *Handler) AddAdmin(id int64, nick string) error {
	err := h.Storage.SaveAdmin(id, nick)
	if err != nil {
		return errors.Wrap(err, "SaveAdmin")
	}
	return nil
}

//get last user to answer
func (h *Handler) Find(toId int64) error {
	fromId, msgId, err := h.Storage.GetLast()
	if err != nil {
		return errors.Wrap(err, "GetLast")
	}
	for _, id := range msgId {
		forward := tgbotapi.NewForward(toId, fromId, id)
		_, err = h.Bot.Send(forward)
		if err != nil {
			return errors.Wrap(err, "Send")
		}
		err := h.Cache.SetUser(id, fromId)
		if err != nil {
			return errors.Wrap(err, "SetUser")
		}
	}
	return nil
}

//answer to message
func (h *Handler) ReplyToMsg(msgId int, txt string) error {
	userId, err := h.Cache.GetUser(msgId)
	if err != nil {
		return errors.Wrap(err, "GetUser")
	}
	msg := tgbotapi.NewMessage(userId, txt)
	_, err = h.Bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "Send")
	}
	return nil
}

//send text to everyone
func (h *Handler) SendAll(txt string) error {
	all, err := h.Storage.GetAll()
	if err != nil {
		return errors.Wrap(err, "GetAll")
	}
	for _, id := range all {
		msg := tgbotapi.NewMessage(id, txt)
		_, err := h.Bot.Send(msg)
		if err != nil {
			return errors.Wrap(err, "Send")
		}
	}
	return nil
}
