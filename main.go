package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	event_consumer "github.com/KokorishviliK/telegram-bot-GEO-citizenship-tests/internal/consumer/event-consumer"
	"github.com/KokorishviliK/telegram-bot-GEO-citizenship-tests/internal/events/telegram"
	"github.com/KokorishviliK/telegram-bot-GEO-citizenship-tests/internal/storage/files"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	storagePath = "internal/storage/files"
	batchSize   = 100
)

func main() {

	bot, err := tgbotapi.NewBotAPI(mustToken())
	if err != nil {
		log.Panic(err)
	}

	eventsProcessor := telegram.New(bot, files.New(storagePath))

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)
	if err := consumer.Start(); err != nil {
		log.Fatal("service is stopped", err)
	}

}

func Handler(rw http.ResponseWriter, r *http.Request) {

	bot, err := tgbotapi.NewBotAPI(mustToken())
	if err != nil {
		log.Panic(err)
	}

	eventsProcessor := telegram.New(bot, files.New(storagePath))

	err = eventsProcessor.HandleUpdate(r)
	if err != nil {
		log.Fatal(err)
	}

}

func mustToken() string {

	var token string

	tokenFlag := flag.String("tg-bot-token", "", "token for telegram bot")
	flag.Parse()

	if *tokenFlag == "" {
		token = os.Getenv("TELEGRAM_APITOKEN")
	} else {
		token = *tokenFlag
	}

	if token == "" {
		log.Fatal("token not specified")
	}

	return token
}
