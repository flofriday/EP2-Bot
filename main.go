package main

import (
	"log"
	"os"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	// Clone the repo if it does not exist
	err := cloneIfNotExist()
	if err != nil {
		log.Fatal("Unable to download the repository: ", err.Error())
		return
	}

	// Setup the telegram repo
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Setup the background task (git crawling)
	startBackgroundManager(bot)

	// Handle the updates
	updates, err := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		go handleMessage(bot, &update)
	}
}
