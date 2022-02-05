package service

import (
	"google.golang.org/api/sheets/v4"
)

type SheetsSrv struct {
	srv *sheets.Service
}

func NewSheetsSrv(srv *sheets.Service) *SheetsSrv {
	return &SheetsSrv{srv: srv}
}

type ISheets interface {
	SaveContact(id int64, name, nick string) error
	GetAll() error
	SaveMsg(id int64, msgId int) error
	GetFirst() (int64, int, error)
}

func (s SheetsSrv) SaveContact(id int64, name, nick string) error {
	panic("implement me")
}

func (s SheetsSrv) GetAll() error {
	panic("implement me")
}

func (s SheetsSrv) SaveMsg(id int64, msgId int) error {
	panic("implement me")
}

func (s SheetsSrv) GetFirst() (int64, int, error) {
	panic("implement me")
}
