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
	token = "TG_TOKEN"
	//google
	sheetUsers  = "SHEET_USERS"
	sheetMsg    = "SHEET_MSG"
	sheetAdmins = "SHEET_ADMINS"
	sheetBanned = "SHEET_BANNED"
	//cache
	cacheAddr = "CACHE_ADDR"
	keepTime  = "CACHE_KEEPTIME"
)

type (
	Conf struct {
		Tg     TgConfig
		Sheets SheetsConfig
		Redis  RedisConfig
	}

	TgConfig struct {
		Token string
	}

	SheetsConfig struct {
		Users  string
		Msg    string
		Admins string
		Banned string
	}

	RedisConfig struct {
		KeepTime int64
		Addr     string
	}
)

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

	keepTimeInt, err := strconv.Atoi(os.Getenv(keepTime))
	if err != nil {
		return nil, errors.Wrap(err, "keepTime")
	}

	return &Conf{
		Tg: TgConfig{
			Token: os.Getenv(token),
		},
		Sheets: SheetsConfig{
			Users:  os.Getenv(sheetUsers),
			Msg:    os.Getenv(sheetMsg),
			Admins: os.Getenv(sheetAdmins),
			Banned: os.Getenv(sheetBanned),
		},
		Redis: RedisConfig{
			KeepTime: int64(keepTimeInt),
			Addr:     os.Getenv(cacheAddr),
		},
	}, nil
}
