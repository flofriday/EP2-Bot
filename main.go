package main

import (
	"log"
	"os"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	// Check if all the right environment variables are set.
	checkEnvironment()

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
		// Handle the current update in a new go routine
		if update.Message != nil {
			go handleMessage(bot, &update)
		}

		// Handle the current update in a new go routine
		if update.CallbackQuery != nil {
			go handleCallBackQuery(bot, &update)
		}

	}
}

// This function checks if the bot got started with all necessary environment variables set.
// If not it will print an error message into the terminal and terminate the program.
func checkEnvironment() {
	isOk := true
	if os.Getenv("TELEGRAM_TOKEN") == "" {
		isOk = false
		log.Println("The TELEGRAM_TOKEN environment variable is not set.")
	}
	if os.Getenv("TELEGRAM_ADMIN") == "" {
		isOk = false
		log.Println("The TELEGRAM_ADMIN environment variable is not set.")
	}
	if os.Getenv("GIT_URL") == "" {
		isOk = false
		log.Println("The GIT_URL environment variable is not set.")
	}

	if isOk == false {
		log.Println("You can find more information about how to configure the bot at:")
		log.Println("https://github.com/flofriday/EP2-Bot")
		os.Exit(1)
	}
}
