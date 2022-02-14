package database

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/sheets/v4"
)

var errNoRows = errors.New("no rows")

type sheetsSrv struct {
	srv    *sheets.Service
	db     string
	msg    string
	admins string
}

func NewSheetsSrv(
	srv *sheets.Service,
	db string,
	msg string,
	admins string) *sheetsSrv {
	return &sheetsSrv{
		srv:    srv,
		db:     db,
		msg:    msg,
		admins: admins,
	}
}

func (s sheetsSrv) LoadAdmins() (map[int64]struct{}, error) {
	//todo:check range
	out := make(map[int64]struct{})
	rsp, err := s.srv.Spreadsheets.Values.Get(s.admins, "Sheet1!A:A").Do()
	if err != nil {
		return nil, errors.Wrap(err, "Get")
	}
	for _, row := range rsp.Values {
		id, err := strconv.Atoi(row[0].(string))
		if err != nil {
			return nil, errors.Wrap(err, "not an int64")
		}
		out[int64(id)] = struct{}{}
	}
	return out, nil
}

func (s sheetsSrv) SaveAdmin(nick string) error {
	//todo:check range for append
	inValue := make([]interface{}, 1)
	inValue[0] = nick
	outValue := make([][]interface{}, 1)
	outValue[0] = inValue
	valRen := sheets.ValueRange{
		MajorDimension: "ROWS",
		//todo:is required?
		Range:           "",
		Values:          outValue,
		ServerResponse:  googleapi.ServerResponse{},
		ForceSendFields: nil,
		NullFields:      nil,
	}
	_, err := s.srv.Spreadsheets.Values.
		Append(s.admins, "Sheet1!A:B", &valRen).
		ValueInputOption("RAW").
		Do()
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve files")
	}
	return nil
}

func (s sheetsSrv) GetLast() (int64, []int, error) {
	msgIds := make([]int, 0)
	rsp, err := s.srv.Spreadsheets.Values.Get(s.db, "Sheet1!A1:C1").Do()
	if err != nil {
		return 0, nil, errors.Wrap(err, "Get")
	}
	id, err := strconv.Atoi(rsp.Values[0][0].(string))
	if err != nil {
		return 0, nil, errors.Wrap(err, "Atoi")
	}
	idsStr := rsp.Values[0][0].(string)
	msgIdsStr := strings.Split(idsStr, ",")
	for _, id := range msgIdsStr {
		idint, err := strconv.Atoi(id)
		if err != nil {
			return 0, nil, errors.Wrap(err, "Atoi")
		}
		msgIds = append(msgIds, idint)
	}
	opt := sheets.ClearValuesRequest{}
	s.srv.Spreadsheets.Values.Clear(s.msg, "Sheet1!A1:C1", &opt)
	return int64(id), msgIds, nil
}

func (s sheetsSrv) SaveContact(id int64, name, nick string) error {
	//todo:duplicates
	inValue := make([]interface{}, 3)
	inValue[0] = id
	inValue[1] = name
	inValue[2] = nick
	outValue := make([][]interface{}, 1)
	outValue[0] = inValue
	valRen := sheets.ValueRange{
		MajorDimension:  "ROWS",
		Range:           "",
		Values:          outValue,
		ServerResponse:  googleapi.ServerResponse{},
		ForceSendFields: nil,
		NullFields:      nil,
	}
	_, err := s.srv.Spreadsheets.Values.
		Append(s.db, "Sheet1!A:C", &valRen).
		ValueInputOption("RAW").
		Do()
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve files")
	}
	return nil
}

func (s sheetsSrv) GetAll() ([]int64, error) {
	out := make([]int64, 0)
	rsp, err := s.srv.Spreadsheets.Values.Get(s.db, "Sheet1!A:A").Do()
	if err != nil {
		return nil, errors.Wrap(err, "Get")
	}
	for _, row := range rsp.Values {
		id, err := strconv.Atoi(row[0].(string))
		if err != nil {
			return nil, errors.Wrap(err, "Atoi")
		}
		out = append(out, int64(id))
	}
	return out, nil
}

func (s sheetsSrv) SaveMsg(id int64, msgId int) error {
	fmt.Println(id, msgId)
	//check if exists
	valueRange, ints, err := s.searchRows(s.msg, string(id), "Sheet1!A:B", )
	if err != nil {
		fmt.Println(1)
		//not a single line error
		if err != errNoRows {
			fmt.Println(2)
			return errors.Wrap(err, "searchRows")
		}
		fmt.Println(3)
		//	insert instead of update
		inValue := make([]interface{}, 2)
		inValue[0] = id
		inValue[1] = strconv.Itoa(msgId)
		outValue := make([][]interface{}, 1)
		outValue[0] = inValue
		valRen := sheets.ValueRange{
			MajorDimension:  "ROWS",
			Range:           "",
			Values:          outValue,
			ServerResponse:  googleapi.ServerResponse{},
			ForceSendFields: nil,
			NullFields:      nil,
		}
		_, err = s.srv.Spreadsheets.Values.
			Append(s.msg, "Sheet1!A:B", &valRen).
			ValueInputOption("RAW").
			Do()
		if err != nil {
			return errors.Wrap(err, "Unable to retrieve files")
		}
		return nil
	}
	fmt.Println(4)
	if len(ints) > 1 {
		return errors.New("not a single row")
	}
	inValue := make([]interface{}, 2)
	inValue[0] = id
	inValue[1] = strconv.Itoa(msgId) + "," + valueRange.Values[0][1].(string)
	outValue := make([][]interface{}, 1)
	outValue[0] = inValue
	valRen := sheets.ValueRange{
		MajorDimension:  "ROWS",
		Range:           valueRange.Range,
		Values:          outValue,
		ServerResponse:  googleapi.ServerResponse{},
		ForceSendFields: nil,
		NullFields:      nil,
	}
	_, err = s.srv.Spreadsheets.Values.
		Update(s.msg, valueRange.Range, &valRen).
		ValueInputOption("RAW").
		Do()
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve files")
	}
	return nil
}

func (s sheetsSrv) searchRows(sheetId, searchInput, searchRange string) (*sheets.ValueRange, []int, error) {
	resp, err := s.srv.Spreadsheets.Values.Get(sheetId, searchRange).Do()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Unable to retrieve data from sheet")
	}
	y := make([]int, 0)
	if len(resp.Values) == 0 {
		return nil, nil, errNoRows
	}
	for i, row := range resp.Values {
		c1, ok := row[0].(string)
		if !ok {
			return nil, nil, errors.Wrap(err, "cast")
		}
		if strings.Contains(c1, searchInput) {
			y = append(y, i+1)
		}
	}
	return resp, y, nil
}
