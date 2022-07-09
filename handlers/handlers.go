package handlers

import (
	"fmt"

	"github.com/CookieNyanCloud/tg-connection-base/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

const (
	welcome     = `Вас приветствует телеграм-бот "Мир, Прогресс и Права Человека"
оставьте сообщение и мы вам скоро ответим`

	adminOk     = `Новый администратор бота успешно добавлен`

	banOk      = `Пользователь забанен`

	feedback    = `Спасибо! Ваше сообщение принято. Если хотите дополнить, пишите нам ещё.`

	regionStart = `Пожалуйста, введите регион вашего проживания:`

	regionOk    = `регион успешно сохранён`

	unknownTxt  = "неизвестная команда"

	bannedTxt   = "ваш аккаунт был заблокирован"

	alreadyAnswered   = "на сообщение уже ответили"
)

type IStorage interface {
	// admins
	LoadAdmins() (map[string]database.Admin, error)
	LoadBanned() (map[string]struct{}, error)
	SaveAdmin(nick string) error
	SetBan(nick string) error
	GetLast() (int64, []int, error)
	// users
	SaveContact(id int64, name, nick string) error
	SaveRegion(id int64, region string) error
	GetAll() ([]int64, error)
	SaveMsg(id int64, msgId int) error
	GetStat() (map[string]int, error)
}

type ICache interface {
	SetUser(msgId int, userId int64) error
	GetUser(msgId int) (int64, error)
	SetBan(userId int64) error
	GetBan(userId int64) (bool, error)

	SetAnswered(msgId int, admin string) error
	GetAnswered(msgId int) (string, error)
}

type handler struct {
	cache   ICache
	storage IStorage
	bot     *tgbotapi.BotAPI

	inRegionDialog map[int64]bool

	admins      map[string]database.Admin
	bannedUsers map[string]struct{}
}

func New(cache ICache, sheets IStorage, bot *tgbotapi.BotAPI) *handler {
	admins, err := sheets.LoadAdmins()
	if err != nil {
		//log.Fatalf("loadAdmins: %v", err)
	}
	fmt.Println(admins)

	bannedUsers, _ := sheets.LoadBanned()

	return &handler{
		cache:          cache,
		storage:        sheets,
		bot:            bot,
		inRegionDialog: make(map[int64]bool),
		admins:         admins,
		bannedUsers:    bannedUsers,
	}
}

type IHandler interface {
	Unknown(id int64) error
	//user
	Starting(id int64, name, nick string) error
	Feedback(id int64, msgId int) error
	StartRegionDialog(id int64) error
	InRegionDialog(id int64) bool
	EndRegionDialog(id int64, region string) error

	IsBanned(id int64, name string) (bool, error)

	//admin
	AddAdmin(id int64, nick string) error
	SetBan(id int64, nick string) error
	ReplyToMsg(msgId int, txt string, chat_id int64, admin string) error
	SendAll(txt string) error
	Find(toId int64) error
	IsAdmin(nick string) bool
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

	msg := tgbotapi.NewMessage(id, feedback)
	_, err = h.bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "Send")
	}

	//send to all admins
	for _, admin := range h.admins {
		if admin.ChatId == 0 {
			continue
		}

		msg := tgbotapi.ForwardConfig{
			BaseChat: tgbotapi.BaseChat{
				ChatID: admin.ChatId,
			},
			FromChatID: id,
			MessageID:  msgId,
		}

		forwarded, err := h.bot.Send(msg)
		if err != nil {
			return errors.Wrap(err, "Send")
		}

		err = h.cache.SetUser(forwarded.MessageID, id)
		if err != nil {
			return errors.Wrap(err, "SetUser")
		}
	}

	return nil
}

func (h *handler) StartRegionDialog(id int64) error {
	h.inRegionDialog[id] = true
	msg := tgbotapi.NewMessage(id, regionStart)
	_, err := h.bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "Send")
	}
	
	return nil
}

func (h *handler) InRegionDialog(id int64) bool {
	_, ok := h.inRegionDialog[id]
	return ok
}

func (h *handler) IsBanned(id int64, name string) (bool, error) {
	_, banned := h.bannedUsers[name]
	if banned {
		msg := tgbotapi.NewMessage(id, bannedTxt)
		h.bot.Send(msg)
	}

	return banned, nil
}

func (h *handler) EndRegionDialog(id int64, region string) error {
	err := h.storage.SaveRegion(id, region)
	if err != nil {
		return errors.Wrap(err, "SaveRegion")
	}

	delete(h.inRegionDialog, id)

	msg := tgbotapi.NewMessage(id, regionOk)
	_, err = h.bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "Send")
	}

	return nil
}

//add new admin
func (h *handler) AddAdmin(id int64, nick string) error {
	err := h.storage.SaveAdmin(nick)
	if err != nil {
		return errors.Wrap(err, "SaveAdmin")
	}

	h.admins[nick] = database.Admin{nick, 0}

	msg := tgbotapi.NewMessage(id, adminOk)
	_, err = h.bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "Send")
	}

	return nil
}

func (h *handler) SetBan(id int64, nick string) error {
	err := h.storage.SetBan(nick)
	if err != nil {
		return errors.Wrap(err, "SetBan")
	}

	h.bannedUsers[nick] = struct{}{}

	msg := tgbotapi.NewMessage(id, banOk)
	_, err = h.bot.Send(msg)
	if err != nil {
		return errors.Wrap(err, "Send")
	}

	return nil
}

func (h *handler) IsAdmin(nick string) bool {
		_, ok := h.admins[nick]
	return ok
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
func (h *handler) ReplyToMsg(msgId int, txt string, chat_id int64, admin string) error {
	other_admin, err := h.cache.GetAnswered(msgId)
	if err != nil {
		return errors.Wrap(err, "GetAnswered")
	}

	if (other_admin != "" && other_admin != admin) {
		msg := tgbotapi.NewMessage(chat_id, alreadyAnswered)
		_, err = h.bot.Send(msg)
		if err != nil {
			return errors.Wrap(err, "Send")
		}
		return nil
	}

	userId, err := h.cache.GetUser(msgId)
	if err != nil {
		return errors.Wrap(err, "GetUser")
	}
	msg := tgbotapi.NewMessage(userId, txt)
	answer, send_err := h.bot.Send(msg)
	if send_err != nil {
		return errors.Wrap(send_err, "Send")
	}

	err = h.cache.SetAnswered(msgId, admin)

	//send answer to all admins
	for _, other_admin := range h.admins {
		if other_admin.ChatId == 0 || other_admin.Nick == admin {
			continue
		}

		msg := tgbotapi.ForwardConfig{
			BaseChat: tgbotapi.BaseChat{
				ChatID: other_admin.ChatId,
			},
			FromChatID: userId,
			MessageID:  answer.MessageID,
		}

		_, err := h.bot.Send(msg)
		if err != nil {
			return errors.Wrap(err, "Send")
		}
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
