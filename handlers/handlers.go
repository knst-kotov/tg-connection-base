package handlers

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

const (
	welcome    = `Вас приветствует телеграм-бот "Мир, Прогресс и Права Человека"
оставьте сообщение и мы вам скоро ответим`

	unknownTxt = "неизвестная команда"
)

type IStorage interface {
	// admins
	LoadAdmins() (map[string]struct{}, error)
	SaveAdmin(nick string) error
	GetLast() (int64, []int, error)
	// users
	SaveContact(id int64, name, nick string) error
	GetAll() ([]int64, error)
	SaveMsg(id int64, msgId int) error
	GetStat() (map[string]int, error)
}

type ICache interface {
	SetUser(msgId int, userId int64) error
	GetUser(msgId int) (int64, error)
	SetBan(userId int64) error
	GetBan(userId int64) (bool, error)
}

type handler struct {
	cache   ICache
	storage IStorage
	bot     *tgbotapi.BotAPI
}

func New(
	cache ICache,
	sheets IStorage,
	bot *tgbotapi.BotAPI) *handler {
	return &handler{
		cache:   cache,
		storage: sheets,
		bot:     bot,
	}
}

type IHandler interface {
	Unknown(id int64) error
	//user
	Starting(id int64, name, nick string) error
	Feedback(id int64, msgId int) error
	//admin
	AddAdmin(nick string) error
	ReplyToMsg(msgId int, txt string) error
	SendAll(txt string) error
	Find(toId int64) error
	LoadAdmins() (map[string]struct{}, error)
	//	todo:?chat
	Stat(id int64) error
}

//unknown command
func (h *handler) Unknown(id int64) error {
	msg := tgbotapi.NewMessage(id, unknownTxt)
	_, err := h.bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "Send")
	}
	return nil
}

//first message, save info
func (h *handler) Starting(id int64, name, nick string) error {
	msg := tgbotapi.NewMessage(id, welcome)
	_, err := h.bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "Send")
	}
	err = h.storage.SaveContact(id, name, nick)
	if err != nil {
		return errors.Wrap(err, "SaveContact")
	}
	return nil
}

//save message id, answered when needed
func (h *handler) Feedback(id int64, msgId int) error {
	err := h.storage.SaveMsg(id, msgId)
	if err != nil {
		return errors.Wrap(err, "SaveMsg")
	}
	return nil
}

//add new admin
func (h *handler) AddAdmin(nick string) error {
	err := h.storage.SaveAdmin(nick)
	if err != nil {
		return errors.Wrap(err, "SaveAdmin")
	}
	return nil
}

//get all admins
func (h *handler) LoadAdmins() (map[string]struct{}, error) {
	return h.storage.LoadAdmins()
}

//get last user to answer
func (h *handler) Find(toId int64) error {
	fromId, msgIds, err := h.storage.GetLast()
	if err != nil {
		return errors.Wrap(err, "GetLast")
	}
	for _, id := range msgIds {
		msg := tgbotapi.ForwardConfig{
			BaseChat: tgbotapi.BaseChat{
				ChatID: toId,
			},
			FromChatID: fromId,
			MessageID:  id,
		}
		forwarded, err := h.bot.Send(msg)
		if err != nil {
			return errors.Wrap(err, "Send")
		}
		err = h.cache.SetUser(forwarded.MessageID, fromId)

		if err != nil {
			return errors.Wrap(err, "SetUser")
		}

	}
	return nil
}

//answer to message
func (h *handler) ReplyToMsg(msgId int, txt string) error {
	userId, err := h.cache.GetUser(msgId)
	if err != nil {
		return errors.Wrap(err, "GetUser")
	}
	msg := tgbotapi.NewMessage(userId, txt)
	_, err = h.bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "Send")
	}
	return nil
}

//send text to everyone
func (h *handler) SendAll(txt string) error {
	all, err := h.storage.GetAll()
	if err != nil {
		return errors.Wrap(err, "GetAll")
	}
	for _, id := range all {
		msg := tgbotapi.NewMessage(id, txt)
		_, err := h.bot.Send(msg)
		if err != nil {
			return errors.Wrap(err, "Send")
		}
	}
	return nil
}

func (h *handler) Stat(id int64) error {
	stat, err := h.storage.GetStat()
	if err != nil {
		return errors.Wrap(err, "GetStat")
	}

	stat_text := "статистика по боту\n"
	for key, value := range stat {
		stat_text += fmt.Sprintf("%v = %v\n", key, value)
	}

	msg := tgbotapi.NewMessage(id, stat_text)
	_, err = h.bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "Send")
	}
	return nil
}
