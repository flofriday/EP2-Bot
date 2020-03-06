package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jasonlvhit/gocron"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func handleMessage(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	// Only allow the owner access
	if isAdmin(update.Message.Chat.ID) == false {
		sendMessage(bot, update, "Sorry, only my master is allowed to access me.")
		return
	}

	// call the right function to handle the command
	switch update.Message.Command() {
	case "ls":
		lsCmd(bot, update)
	case "pull":
		pullCmd(bot, update)
	case "readme":
		readmeCmd(bot, update)
	case "cat":
		catCmd(bot, update)
	case "history":
		historyCmd(bot, update)
	default:
		sendMessage(bot, update, "Sorry, I don't know that command.")
	}

}

// This function must not be called as a goroutine because it should block and will return once the automatic background
// jobs are correctly set up
func startBackgroundManager(bot *tgbotapi.BotAPI) {
	// First call the background task to ensure it ran once
	backgroundJob(bot)

	// Setup the automatic call of backgroundJob
	go func() {
		gocron.Every(10).Minutes().Do(backgroundJob, bot)
		<-gocron.Start()
	}()
}

func backgroundJob(bot *tgbotapi.BotAPI) {
	// Get the current Hash
	oldHash, err := getCurrentCommit()
	if err != nil {
		log.Println("An error occourced with the background task:", err.Error())
		return
	}
	fmt.Println("Old Hash", oldHash)

	// Pull the repo
	err = pull()
	if err != nil {
		log.Println("An error occourced with the background task, when pulling the repo:", err.Error())
		return
	}
	cur, _ := getCurrentCommit()
	fmt.Println("new Hash", cur)

	// Get all the new commits
	newCommits, err := historySince(oldHash)
	if err != nil {
		log.Println("An error occourced with the background task, pulling worked fine but now I can't get the commits.", err.Error())
		return
	}

	// Send the message about all the new commits
	if len(newCommits) == 0 {
		log.Println("Backgroundjob ran, no new commits to send the user.")
		return
	}

	message := "*New commits:*🎉🎊\n"
	for _, commit := range newCommits {
		message += fmt.Sprintf("*commit %s*\nAuthor: %s<%s>\nDate: %s\n``` %s ```\n\n",
			commit.Hash, commit.Author.Name, commit.Author.Email, commit.Author.When.Local().Format("02.01.2006 15:04"), commit.Message)
	}

	// Send the message
	log.Println("Backgroundjob ran, sent the user the updates.")
	admin, _ := strconv.ParseInt(os.Getenv("TELEGRAM_ADMIN"), 10, 64)
	msg := tgbotapi.NewMessage(admin, message)
	msg.ParseMode = "Markdown"
	_, _ = bot.Send(msg)
}

func lsCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	files, err := listFiles(update.Message.CommandArguments())
	if err != nil {
		sendMessage(bot, update, fmt.Sprint("An error occoured when listing the files.\n`Error: %s`", err.Error()))
		return
	}

	message := ""
	for _, file := range files {
		message += file + "\n"
	}
	sendMessage(bot, update, message)
}

func catCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	content, err := readFile(update.Message.CommandArguments())
	if err != nil {
		sendMessage(bot, update, fmt.Sprintf("An error occoured when reading a file.\n`Error: %s`", err.Error()))
		return
	}

	_, filename := filepath.Split(update.Message.CommandArguments())
	message := fmt.Sprintf("*%s*\n```%s```", filename, content)
	sendMessage(bot, update, message)
}

func readmeCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	content, err := readFile("README.md")
	if err != nil {
		sendMessage(bot, update, fmt.Sprintf("An error occoured when reading a file.\n`Error: %s`", err.Error()))
		return
	}

	message := fmt.Sprintf("*README.md*\n``` %s ```", content)
	sendMessage(bot, update, message)
}

func pullCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	oldHash, err := getCurrentCommit()
	if err != nil {
		sendMessage(bot, update, fmt.Sprintf("An error occoured when pulling the repository.\n`Error: %s`", err.Error()))
		return
	}

	err = pull()
	if err != nil {
		sendMessage(bot, update, fmt.Sprintf("An error occoured when pulling the repository.\n`Error: %s`", err.Error()))
		return
	}

	newCommits, err := historySince(oldHash)
	if err != nil {
		sendMessage(bot, update, fmt.Sprintf("Pulling worked fine, however I cannot get the commits new with this pull.\n`Error: %s`", err.Error()))
		return
	}

	if len(newCommits) == 0 {
		sendMessage(bot, update, "No new commits on origin. Repository is already up to date.")
		return
	}

	message := "*New commits:*\n\n"
	for _, commit := range newCommits {
		message += fmt.Sprintf("*commit %s*\nAuthor: %s<%s>\nDate: %s\n``` %s ```\n\n",
			commit.Hash, commit.Author.Name, commit.Author.Email, commit.Author.When.Local().Format("02.01.2006 15:04"), commit.Message)
	}
	sendMessage(bot, update, message)
}

func historyCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	commits, err := history()
	if err != nil {
		sendMessage(bot, update, fmt.Sprintf("An error occoured when reading the repository.\n`Error: %s`", err.Error()))
		return
	}

	message := ""
	for _, commit := range commits {
		message += fmt.Sprintf("*commit %s*\nAuthor: %s<%s>\nDate: %s\n``` %s ```\n\n",
			commit.Hash, commit.Author.Name, commit.Author.Email, commit.Author.When.Local().Format("02.01.2006 15:04"), commit.Message)
	}
	sendMessage(bot, update, message)
}

func sendMessage(bot *tgbotapi.BotAPI, update *tgbotapi.Update, text string) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	_, _ = bot.Send(msg)
}

func isAdmin(telegramID int64) bool {
	original, _ := strconv.ParseInt(os.Getenv("TELEGRAM_ADMIN"), 10, 64)
	return telegramID == original
}
