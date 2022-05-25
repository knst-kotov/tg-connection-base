package database

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
	banned string
}

func NewSheetsSrv(
	srv *sheets.Service,
	db string,
	msg string,
	admins string,
	banned string) *sheetsSrv {
	return &sheetsSrv{
		srv:    srv,
		db:     db,
		msg:    msg,
		admins: admins,
		banned: banned,
	}
}

func LoadColumnAsString(s sheetsSrv, table string) (map[string]struct{}, error) {
	out := make(map[string]struct{})
	rsp, err := s.srv.Spreadsheets.Values.Get(table, "Sheet1!A:A").Do()
	if err != nil {
		return nil, errors.Wrap(err, "Get")
	}
	for _, row := range rsp.Values {
		nick, ok := row[0].(string)
		if !ok {
			return nil, errors.Wrap(err, "not a string")
		}
		out[nick] = struct{}{}
	}
	return out, nil
}

func SaveValue(s sheetsSrv, table string, value string) error {
	inValue := make([]interface{}, 1)
	inValue[0] = value
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
		Append(table, "Sheet1!A:B", &valRen).
		ValueInputOption("RAW").
		Do()
	if err != nil {
		return errors.Wrap(err, "Unable to append value")
	}
	return nil
}

func (s sheetsSrv) LoadAdmins() (map[string]struct{}, error) {
	return LoadColumnAsString(s, s.admins)
}

func (s sheetsSrv) LoadBanned() (map[string]struct{}, error) {
	return LoadColumnAsString(s, s.banned)
}

func (s sheetsSrv) SaveAdmin(nick string) error {
	return SaveValue(s, s.admins, nick)
}

func (s sheetsSrv) SetBan(nick string) error {
	return SaveValue(s, s.banned, nick)
}

func (s sheetsSrv) GetLast() (int64, []int, error) {
	err := s.sortSheet()
	if err != nil {
		return 0, nil, errors.Wrap(err, "sortSheet")
	}
	rsp, err := s.srv.Spreadsheets.Values.
		Get(s.msg, "Sheet1!A1:B1").
		Do()
	if err != nil {
		return 0, nil, errors.Wrap(err, "Get")
	}
	if len(rsp.Values) == 0 {
		return 0, nil, errNoRows
	}
	id, err := strconv.Atoi(rsp.Values[0][0].(string))
	if err != nil {
		return 0, nil, errors.Wrap(err, "Atoi")
	}
	idsStr := rsp.Values[0][1].(string)
	msgIdsStrs := strings.Split(idsStr, ",")
	msgIds := make([]int, len(msgIdsStrs))
	for i, id := range msgIdsStrs {
		idint, err := strconv.Atoi(id)
		if err != nil {
			return 0, nil, errors.Wrap(err, "Atoi")
		}
		msgIds[len(msgIds)-i-1] = idint
	}
	err = s.clearRow()
	if err != nil {
		return 0, nil, errors.Wrap(err, "clearRow")
	}
	err = s.sortSheet()
	if err != nil {
		return 0, nil, errors.Wrap(err, "sortSheet")
	}
	return int64(id), msgIds, nil
}

func (s sheetsSrv) SaveContact(id int64, name, nick string) error {
	_, ints, err := s.searchRows(s.db, strconv.Itoa(int(id)), "Sheet1!A:A")
	if err != nil && err != errNoRows {
		return errors.Wrap(err, "searchRows")
	}
	if len(ints) > 0 {
		return errors.New("duplicate")
	}
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
	_, err = s.srv.Spreadsheets.Values.
		Append(s.db, "Sheet1!A:C", &valRen).
		ValueInputOption("RAW").
		Do()
	if err != nil {
		return errors.Wrap(err, "Unable to retrieve files")
	}
	return nil
}

func (s sheetsSrv) SaveRegion(id int64, region string) error {
	_, ints, err := s.searchRows(s.db, strconv.Itoa(int(id)), "Sheet1!A:A")
	if err != nil && err != errNoRows {
		return errors.Wrap(err, "searchRows")
	}
	if len(ints) != 1 {
		return errors.New("contact not found")
	}

	inValue := make([]interface{}, 1)
	inValue[0] = region
	outValue := make([][]interface{}, 1)
	outValue[0] = inValue
	r := fmt.Sprintf("Sheet1!D%d:D%d", ints[0], ints[0])
	valRen := sheets.ValueRange{
		MajorDimension:  "ROWS",
		Range:           "",
		Values:          outValue,
		ServerResponse:  googleapi.ServerResponse{},
		ForceSendFields: nil,
		NullFields:      nil,
	}
	_, err = s.srv.Spreadsheets.Values.
		Update(s.db, r, &valRen).
		ValueInputOption("RAW").
		Do()
	if err != nil {
		return errors.Wrap(err, "unable to insert values")
	}

	return nil
}

func (s sheetsSrv) GetAll() ([]int64, error) {
	out := make([]int64, 0)
	rsp, err := s.srv.Spreadsheets.Values.
		Get(s.db, "Sheet1!A:A").
		Do()
	if err != nil {
		return nil, errors.Wrap(err, "Get")
	}
	if len(rsp.Values) == 0 {
		return nil, errNoRows
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
	fmt.Println("start")
	//check if exists
	valueRange, ints, err := s.searchRows(s.msg, strconv.FormatInt(id, 10), "Sheet1!A:C", )
	//not a single line error
	if err != nil {
		fmt.Println("err 1")
		return errors.Wrap(err, "searchRows")
	}
	fmt.Println(1)
	//	insert instead of update
	if len(ints) == 0 {
		fmt.Println("first")
		inValue := make([]interface{}, 3)
		inValue[0] = id
		inValue[1] = msgId
		inValue[2] = time.Now().Unix()
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
			Append(s.msg, "Sheet1!A:C", &valRen).
			ValueInputOption("RAW").
			Do()
		if err != nil {
			return errors.Wrap(err, "Append")
		}
		fmt.Println("end")
		return nil
	}
	fmt.Println(2)
	if len(ints) != 1 {
		return errors.New("not a single row")
	}
	inValue := make([]interface{}, 3)
	inValue[0] = id
	inValue[1] = strconv.Itoa(msgId) + "," + valueRange.Values[ints[0]-1][1].(string)
	inValue[2] = time.Now().Unix()
	outValue := make([][]interface{}, 1)
	outValue[0] = inValue
	r := fmt.Sprintf("Sheet1!A%d:C%d", ints[0], ints[0])
	valRen := sheets.ValueRange{
		MajorDimension:  "ROWS",
		Range:           r,
		Values:          outValue,
		ServerResponse:  googleapi.ServerResponse{},
		ForceSendFields: nil,
		NullFields:      nil,
	}
	fmt.Println(r)
	_, err = s.srv.Spreadsheets.Values.
		Update(s.msg, r, &valRen).
		ValueInputOption("RAW").
		Do()
	if err != nil {
		return errors.Wrap(err, "Update")
	}
	fmt.Println("end")
	return nil
}

func (s sheetsSrv) searchRows(sheetId, searchInput, searchRange string) (*sheets.ValueRange, []int, error) {
	resp, err := s.srv.Spreadsheets.Values.
		Get(sheetId, searchRange).
		Do()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Unable to retrieve data from sheet")
	}
	y := make([]int, 0)
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

func (s *sheetsSrv) sortSheet() error {
	sort := &sheets.SortRangeRequest{
		Range: &sheets.GridRange{
			EndColumnIndex:   3,
			EndRowIndex:      0,
			SheetId:          0,
			StartColumnIndex: 0,
			StartRowIndex:    0,
			ForceSendFields:  nil,
			NullFields:       nil,
		},
		SortSpecs: []*sheets.SortSpec{
			{
				BackgroundColor:           nil,
				BackgroundColorStyle:      nil,
				DataSourceColumnReference: nil,
				DimensionIndex:            2,
				ForegroundColor:           nil,
				ForegroundColorStyle:      nil,
				SortOrder:                 "ASCENDING",
				ForceSendFields:           nil,
				NullFields:                nil,
			},
		},
		ForceSendFields: nil,
		NullFields:      nil,
	}
	req := []*sheets.Request{{
		SortRange: sort,
	}}
	reqSort := &sheets.BatchUpdateSpreadsheetRequest{
		IncludeSpreadsheetInResponse: false,
		Requests:                     req,
		ResponseIncludeGridData:      false,
		ResponseRanges:               nil,
		ForceSendFields:              nil,
		NullFields:                   nil,
	}
	_, err := s.srv.Spreadsheets.BatchUpdate(s.msg, reqSort).Do()
	return err
}

func (s *sheetsSrv) clearRow() error {
	inValue := make([]interface{}, 3)
	inValue[0] = ""
	inValue[1] = ""
	inValue[2] = ""
	outValue := make([][]interface{}, 1)
	outValue[0] = inValue
	valRen := sheets.ValueRange{
		MajorDimension:  "ROWS",
		Range:           "Sheet1!A1:C1",
		Values:          outValue,
		ServerResponse:  googleapi.ServerResponse{},
		ForceSendFields: nil,
		NullFields:      nil,
	}
	_, err := s.srv.Spreadsheets.Values.
		Update(s.msg, "Sheet1!A1:C1", &valRen).
		ValueInputOption("RAW").
		Do()
	fmt.Println("OUT")
	return err
}

func (s sheetsSrv) GetStat() (map[string]int, error) {
	out := make(map[string]int)

	rsp, err := s.srv.Spreadsheets.Values.Get(s.db, "Sheet1!A:A").Do()
	if err != nil {
		return nil, errors.Wrap(err, "Get")
	}
	out["contacts"] = len(rsp.Values)

	rsp, err = s.srv.Spreadsheets.Values.Get(s.msg, "Sheet1!A:A").Do()
	if err != nil {
		return nil, errors.Wrap(err, "Get")
	}
	out["messages"] = len(rsp.Values)

	return out, nil
}
