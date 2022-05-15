package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/CookieNyanCloud/tg-connection-base/cache"
	"github.com/CookieNyanCloud/tg-connection-base/config"
	"github.com/CookieNyanCloud/tg-connection-base/database"
	"github.com/CookieNyanCloud/tg-connection-base/handlers"
	"github.com/CookieNyanCloud/tg-connection-base/pkg"
	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var helpTxt = `
/help - помощь
/next - сообщения следующего на очереди
/add (nickname) - добавить админа по нику
/all (text) - отправить всем пользователям текст
/stat - статистика по боту
`

func main() {

	var ctx = context.Background()

	//env vars
	conf, err := config.InitConf()
	if err != nil {
		log.Fatalf("conf: %v", err)
	}

	//cache
	redisClient, err := pkg.NewRedisClient(conf.Redis.Addr, ctx)
	if err != nil {
		log.Fatalf("redis client: %v", err)
	}
	redisCache := cache.New(ctx, redisClient.Client, conf.Redis.KeepTime)

	//google sheets
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("sheets.json"))
	if err != nil {
		log.Fatalf("Unable to parse credantials file: %v", err)
	}
	sheetsSrv := database.NewSheetsSrv(srv,
		conf.Sheets.Users, conf.Sheets.Msg, conf.Sheets.Admins, conf.Sheets.Banned)

	//graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	go func(ctx context.Context, cache *redis.Client) {
		<-quit
		fmt.Println("shutdown")
		const timeout = 5 * time.Second
		ctx, shutdown := context.WithTimeout(context.Background(), timeout)
		defer shutdown()
		if err := cache.Close(); err != nil {
			log.Fatalf("closing cache: %v", err)
		}
		os.Exit(1)
	}(ctx, redisClient.Client)

	//tg
	bot, updates, err := pkg.StartBot(conf.Tg.Token)
	if err != nil {
		log.Fatalf("tg: %v", err)
	}
	handler := handlers.New(redisCache, sheetsSrv, bot)

	//load admins from google sheet
	admins, err := handler.LoadAdmins()
	if err != nil {
		log.Fatalf("loadAdmins: %v", err)
	}

	for update := range updates {

		if update.Message == nil {
			continue
		}

		// admins
		if _, ok := admins[update.Message.Chat.UserName]; ok {
			//commands
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "АДМИН\n"+helpTxt)
					_, err = bot.Send(msg)
					logErr("Send", err)
				case "help":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpTxt)
					_, err = bot.Send(msg)
					logErr("Send", err)
				case "next":
					err := handler.Find(update.Message.Chat.ID)
					logErr("Find", err)
				case "add":
					nick := strings.Trim(update.Message.CommandArguments(), "@")
					err := handler.AddAdmin(nick)
					admins[nick] = struct{}{}
					logErr("AddAdmin", err)
				case "all":
					err := handler.SendAll(update.Message.CommandArguments())
					logErr("SendAll", err)
				case "stat":
					err := handler.Stat(update.Message.Chat.ID)
					logErr("Stat", err)
				default:
					err := handler.Unknown(update.Message.Chat.ID)
					logErr("Unknown", err)
				}
			}

			// answer to user
			if update.Message.ReplyToMessage != nil {
				err := handler.ReplyToMsg(update.Message.ReplyToMessage.MessageID, update.Message.Text)
				logErr("ReplyToMsg", err)
				continue
			}
			continue
		}

		banned,err := handler.IsBanned(update.Message.Chat.ID, update.Message.Chat.UserName)
		if err != nil {
			logErr("IsBanned", err)
		} else if banned {
			fmt.Printf("user %v is banned\n", update.Message.Chat.UserName)
			continue
		}

		// users
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				err := handler.Starting(
					update.Message.Chat.ID,
					update.Message.From.FirstName + " " + update.Message.From.LastName,
					update.Message.Chat.UserName)
				logErr("start", err)
			case "region":
				err := handler.StartRegionDialog(update.Message.Chat.ID)
				logErr("StartRegionDialog", err)
			default:
				err := handler.Unknown(update.Message.Chat.ID)
				logErr("Unknown", err)
			}
			continue
		}

		// message from user
		if handler.InRegionDialog(update.Message.Chat.ID) {
			region := update.Message.Text
			err := handler.EndRegionDialog(update.Message.Chat.ID, region)
			logErr("EndRegionDialog", err)
		} else {
			err := handler.Feedback(update.Message.Chat.ID, update.Message.MessageID)
			logErr("Feedback", err)
		}

	}
}

func logErr(msg string, err error) {
	if err != nil {
		fmt.Printf(msg+": %v\n", err)
	}
}
