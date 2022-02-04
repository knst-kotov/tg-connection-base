package service

import (
	"time"

	"google.golang.org/api/sheets/v4"
)

type Contact struct {
	Nic  string
	Name string
}

type Appeal struct {
	Nic       string
	Name      string
	Time      time.Time
	Locations string
	Text      string
}

type SheetsSrv struct {
	srv *sheets.Service
}

func NewSheetsSrv(srv *sheets.Service) *SheetsSrv {
	return &SheetsSrv{srv: srv}
}

type ISheets interface {
	SaveContact() error
	GetAll() error
	SaveMsg() error
}
