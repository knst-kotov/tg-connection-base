package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	cache2 "github.com/CookieNyanCloud/tg-connection-base/cache"
	"github.com/CookieNyanCloud/tg-connection-base/config"
	"github.com/CookieNyanCloud/tg-connection-base/database"
	"github.com/CookieNyanCloud/tg-connection-base/handlers"
	"github.com/CookieNyanCloud/tg-connection-base/pkg"
	"github.com/go-redis/redis/v8"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

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
	cache := cache2.NewCache(ctx, redisClient.Client, conf.Redis.KeepTime)

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
	handler := handlers.NewHandler(cache, sheetsSrv, bot)

	//load admins from google sheet
	admins, err := handler.Storage.LoadAdmins()
	if err != nil {
		log.Fatalf("loadAdmins: %v", err)
	}

	for update := range updates {

		if update.Message == nil {
			continue
		}
		// admins
		if _, ok := admins[update.Message.Chat.ID]; ok {
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				// get next users messages
				case "last":
					err := handler.Find(update.Message.Chat.ID)
					logErr("Find", err)
				// add admin
				case "add":
					err := handler.AddAdmin(update.Message.CommandArguments())
					logErr("AddAdmin", err)

				case "all":
					err := handler.SendAll(update.Message.CommandArguments())
					logErr("SendAll", err)

				// no such command
				default:
					err := handler.Unknown(update.Message.Chat.ID)
					logErr("Unknown", err)
				}

				continue
			}

			// answer to user
			if update.Message.ReplyToMessage != nil {
				err := handler.ReplyToMsg(update.Message.ReplyToMessage.MessageID, update.Message.Text)
				logErr("ReplyToMsg", err)
			}
		}

		// users
		if update.Message.IsCommand() {
			switch update.Message.Command() {

			// save user inter
			case "start":
				err := handler.Starting(
					update.Message.Chat.ID,
					update.Message.From.FirstName+" "+update.Message.From.LastName,
					update.Message.Chat.UserName)
				logErr("start", err)

			// no such command
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
