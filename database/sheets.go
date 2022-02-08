package database

import (
	"github.com/pkg/errors"
	"google.golang.org/api/sheets/v4"
)

type SheetsSrv struct {
	srv    *sheets.Service
	db     string
	msg    string
	admins string
}

func NewSheetsSrv(
	srv *sheets.Service,
	db string,
	msg string,
	admins string) *SheetsSrv {
	return &SheetsSrv{
		srv:    srv,
		db:     db,
		msg:    msg,
		admins: admins,
	}
}

type IStorage interface {
	ClearMsgs(id int64) error
	// admins
	LoadAdmins() (map[int64]struct{}, error)
	SaveAdmin(nick string) error
	GetLast() (int64, []int, error)
	// users
	SaveContact(id int64, name, nick string) error
	GetAll() ([]int64, error)
	SaveMsg(id int64, msgId int) error
}

func (s SheetsSrv) LoadAdmins() (map[int64]struct{}, error) {
	readRange := "Sheet1!A"
	resp, err := s.srv.Spreadsheets.Values.Get(s.admins, readRange).Do()
	if err != nil {
		return nil, errors.Wrap(err, "Get")
	}
	out := make(map[int64]struct{})
	if len(resp.Values) == 0 {
		return nil, errors.New("no rows")
	}
	for _, row := range resp.Values {
		for _, cell := range row {
			c, ok := cell.(int64)
			if !ok {
				return nil, errors.New("not a number")
			}
			out[c] = struct{}{}
		}
	}
	return out, nil
}

func (s SheetsSrv) SaveAdmin(nick string) error {
	panic("implement me")
}

func (s SheetsSrv) GetLast() (int64, []int, error) {
	panic("implement me")
}

func (s SheetsSrv) SaveContact(id int64, name, nick string) error {
	panic("implement me")
}

func (s SheetsSrv) GetAll() ([]int64, error) {
	panic("implement me")
}

func (s SheetsSrv) SaveMsg(id int64, msgId int) error {
	panic("implement me")
}

func (s SheetsSrv) ClearMsgs(id int64) error {
	panic("implement me")
}
