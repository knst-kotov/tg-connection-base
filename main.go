package main

import (
	"context"
	"fmt"
	"log"

	"github.com/CookieNyanCloud/tg-connection-base/config"
	"github.com/CookieNyanCloud/tg-connection-base/service"
	"github.com/CookieNyanCloud/tg-connection-base/tgBot"
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
	//redisClient, err := repo.NewRedisClient(conf.RedisAddr, ctx)
	//if err != nil {
	//	log.Fatalf("redis client: %v", err)
	//}
	//cache := repo.NewCache(redisClient.Client, conf.KeepTime)

	//google sheets
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile("drive.json"))
	if err != nil {
		log.Fatalf("Unable to parse credantials file: %v", err)
	}
	sheets := service.NewSheetsSrv(srv)

	//graceful shutdown
	//quit := make(chan os.Signal, 1)
	//signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	//go func(ctx context.Context, db *redis.Client) {
	//	<-quit
	//	fmt.Println("shutdown")
	//	const timeout = 5 * time.Second
	//	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	//	defer shutdown()
	//	if err := db.Close(); err != nil {
	//		log.Fatalf("closing db: %v", err)
	//	}
	//	os.Exit(1)
	//}(ctx, redisClient.Client)

	//tg
	bot, updates, err := tgBot.StartBot(conf.Token)
	if err != nil {
		log.Fatalf("tg: %v", err)
	}
	//handler := tgBot.NewHandler(ctx, conf.Chat, cache, sheets, bot)
	handler := tgBot.NewHandler(ctx, conf.Chat, sheets, bot)

	for update := range updates {

		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				err := handler.Starting()
				logErr("start: %v", err)
			default:
				err := handler.Unknown()
				logErr("start: %v", err)
			}
			continue
		}

		if update.Message.Text == "обратная связь" {
			err := handler.Feedback()
			logErr("feedback: %v", err)
			continue
		}
	}
}

func logErr(msg string, err error) {
	if err != nil {
		fmt.Printf(msg, err)
	}
}
