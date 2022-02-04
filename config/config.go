package config

import (
	"flag"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

const (
	//tg
	token = "TOKEN"
	chat  = "CHAT"
	//cache
	redisAddr = "REDIS"
	keepTime  = "KEEP_TIME"
	//google
	sheetDB  = "SHEET_DB"
	sheetMsg = "SHEET_MSG"
)

type Conf struct {
	KeepTime  int64
	Token     string
	Chat      int64
	RedisAddr string
	SheetDB   string
	SheetMsg  string
}

func InitConf() (*Conf, error) {
	var test bool
	flag.BoolVar(&test, "test", false, "off for docker")
	flag.Parse()
	return envVar(test)
}

func envVar(test bool) (*Conf, error) {

	if test {
		err := godotenv.Load(".env")
		if err != nil {
			return nil, errors.Wrap(err, "envVar load")
		}
	}

	chatInt, err := strconv.Atoi(os.Getenv(chat))
	if err != nil {
		return nil, errors.Wrap(err, "envVar convert chat")
	}
	keepTimeInt, err := strconv.Atoi(os.Getenv(keepTime))
	if err != nil {
		return nil, errors.Wrap(err, "envVar convert time")
	}

	return &Conf{
		KeepTime:  int64(keepTimeInt),
		Token:     os.Getenv(token),
		Chat:      int64(chatInt),
		RedisAddr: os.Getenv(redisAddr),
		SheetDB:   os.Getenv(sheetDB),
		SheetMsg:  os.Getenv(sheetMsg),
	}, nil
}
