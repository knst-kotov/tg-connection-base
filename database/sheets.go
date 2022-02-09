package database

import (
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
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
	SaveAdmin(id int64, nick string) error
	GetLast() (int64, []int, error)
	// users
	SaveContact(id int64, name, nick string) error
	GetAll() ([]int64, error)
	SaveMsg(id int64, msgId int) error
}

func (s SheetsSrv) LoadAdmins() (map[int64]struct{}, error) {
	out := make(map[int64]struct{})
	rsp, err := s.srv.Spreadsheets.Values.Get(s.admins, "Sheet1!A1:A1").Do()
	if err != nil {
		return nil, errors.Wrap(err, "Get")
	}
	for _, row := range rsp.Values {
		id, ok := row[0].(int64)
		if !ok {
			return nil, errors.Wrap(err, "not an int64")
		}
		out[id] = struct{}{}
	}
	return out, nil
}

func (s SheetsSrv) SaveAdmin(id int64, nick string) error {
	//todo:range for append
	inValue := make([]interface{}, 1)
	inValue[0] = id
	inValue[1] = nick
	outValue := make([][]interface{}, 1)
	outValue[0] = inValue
	valRen := sheets.ValueRange{
		MajorDimension:  "ROWS",
		Range:           "Sheet1!A1:A1",
		Values:          outValue,
		ServerResponse:  googleapi.ServerResponse{},
		ForceSendFields: nil,
		NullFields:      nil,
	}
	_, err := s.srv.Spreadsheets.Values.
		Append(s.admins, "Sheet1!A1:A1", &valRen).
		ValueInputOption("RAW").
		Do()
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve files")
	}
	return nil
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

func (s SheetsSrv) searchRows(srv *sheets.Service, sheetId, search string) (*sheets.ValueRange, []int, error) {
	readRange := "Sheet1!A:B"
	resp, err := srv.Spreadsheets.Values.Get(sheetId, readRange).Do()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Unable to retrieve data from sheet")
	}
	var y []int
	if len(resp.Values) == 0 {
		return nil, nil, errors.New("no rows")
	}
	for i, row := range resp.Values {
		for _, cell := range row {
			c1, ok := cell.(string)
			if !ok {
				return nil, nil, errors.Wrap(err, "cast")
			}
			if strings.Contains(c1, search) {
				y = append(y, i+1)
			}
		}
	}
	return resp, y, nil
}
