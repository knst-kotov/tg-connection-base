package service

import (
	"fmt"
	"log"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
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
	SaveAdmin(id int64, nick string) error
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

func (s SheetsSrv) SaveAdmin(id int64, nick string) error {
	return nil
}

func searchRows(srv *sheets.Service, id, name string) (*sheets.ValueRange, []int, error) {
	readRange := "Sheet1!A:B"
	resp, err := srv.Spreadsheets.Values.Get(id, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}
	var y []int
	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
		return &sheets.ValueRange{
			MajorDimension:  "",
			Range:           "",
			Values:          nil,
			ServerResponse:  googleapi.ServerResponse{},
			ForceSendFields: nil,
			NullFields:      nil,
		}, []int{}, errors.New("no rows")
	} else {
		for i, row := range resp.Values {
			for _, cell := range row {
				c1 := strings.ToLower(fmt.Sprintf("%v", cell))
				c2 := strings.ToLower(fmt.Sprintf("%v", cell))
				if strings.Contains(c1, strings.ToLower(name)) || strings.Contains(c2, strings.ToLower(name)) {
					y = append(y, i+1)
				}
			}
		}
	}
	return resp, y, nil
}
