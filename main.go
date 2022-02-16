package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
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
/next - сообщения соедующего на очереди
/add (nickname) - добавть админа по нику
/all (text) - отправить всем пользователям текст
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
	redisCache := cache.NewCache(ctx, redisClient.Client, conf.Redis.KeepTime)

	//google sheets
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("sheets.json"))
	if err != nil {
		log.Fatalf("Unable to parse credantials file: %v", err)
	}
	sheetsSrv := database.NewSheetsSrv(srv, conf.Sheets.Users, conf.Sheets.Msg, conf.Sheets.Admins)

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
	handler := handlers.NewHandler(redisCache, sheetsSrv, bot)

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
					_, _ = bot.Send(msg)
				case "help":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpTxt)
					_, _ = bot.Send(msg)
				case "next":
					err := handler.Find(update.Message.Chat.ID)
					logErr("Find", err)
				case "add":
					err := handler.AddAdmin(update.Message.CommandArguments())
					admins[update.Message.CommandArguments()] = struct{}{}
					logErr("AddAdmin", err)
				case "all":
					err := handler.SendAll(update.Message.CommandArguments())
					logErr("SendAll", err)
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

		// users
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				err := handler.Starting(
					update.Message.Chat.ID,
					update.Message.From.FirstName+" "+update.Message.From.LastName,
					update.Message.Chat.UserName)
				logErr("start", err)
			default:
				err := handler.Unknown(update.Message.Chat.ID)
				logErr("Unknown", err)
			}
			continue
		}

		// message from user
		err := handler.Feedback(update.Message.Chat.ID, update.Message.MessageID)
		logErr("Feedback", err)

	}
}

func logErr(msg string, err error) {
	if err != nil {
		fmt.Printf(msg+": %v\n", err)
	}
}
