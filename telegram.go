package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jasonlvhit/gocron"
	gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	case "download":
		downloadCmd(bot, update)
	case "history":
		historyCmd(bot, update)
	case "start":
		helpCmd(bot, update)
	case "help":
		helpCmd(bot, update)
	default:
		sendMessage(bot, update, "Sorry, I don't know that command.\nType /help to see what I know.")
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
		log.Println("An error occourced with the background task, while pulling the repo:", err.Error())
		return
	}
	cur, _ := getCurrentCommit()
	fmt.Println("New Hash", cur)

	// Get all the new commits
	newCommits, err := historySince(oldHash)
	if err != nil {
		log.Println("An error occourced with the background task, pulling worked fine but now I can't get the commits.", err.Error())
		return
	}

	// Filter commits from the user. The user commited them so why would he want to see them ?
	newFilteredCommits := make([]gitobject.Commit, 0, 0)
	for _, c := range newCommits {
		if strings.Contains(c.Author.Email, getGitUser()) {
			continue
		}

		newFilteredCommits = append(newFilteredCommits, c)
	}

	// Check if there are new commits to notify
	if len(newFilteredCommits) == 0 {
		log.Println("Backgroundjob ran, no new commits to send the user.")
		return
	}

	// Create a message and send it the user
	message := "*New commits:*🎉🎊\n"
	for _, commit := range newFilteredCommits {
		formatCommit(commit)
	}

	// Send the message
	admin, _ := strconv.ParseInt(os.Getenv("TELEGRAM_ADMIN"), 10, 64)
	msg := tgbotapi.NewMessage(admin, message)
	msg.ParseMode = "Markdown"
	_, _ = bot.Send(msg)
	log.Println("Backgroundjob ran, sent the user the updates.")
}

func lsCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	files, err := listFiles(update.Message.CommandArguments())
	if err != nil {
		sendMessage(bot, update, fmt.Sprint("An error occoured while listing the files.\n`Error: %s`", err.Error()))
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
		sendMessage(bot, update, fmt.Sprintf("An error occoured while reading a file.\n`Error: %s`", err.Error()))
		return
	}

	_, filename := filepath.Split(update.Message.CommandArguments())
	message := fmt.Sprintf("*%s*\n```%s```", filename, string(content))
	sendMessage(bot, update, message)
}

func downloadCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	path := filepath.Join(getGitDir(), update.Message.CommandArguments())

	sendFile(bot, update, path)
}

func readmeCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	content, err := readFile("README.md")
	if err != nil {
		sendMessage(bot, update, fmt.Sprintf("An error occoured while reading a file.\n`Error: %s`", err.Error()))
		return
	}

	message := fmt.Sprintf("*README.md*\n``` %s ```", content)
	sendMessage(bot, update, message)
}

func pullCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	oldHash, err := getCurrentCommit()
	if err != nil {
		sendMessage(bot, update, fmt.Sprintf("An error occoured while pulling the repository.\n`Error: %s`", err.Error()))
		return
	}

	err = pull()
	if err != nil {
		sendMessage(bot, update, fmt.Sprintf("An error occoured while pulling the repository.\n`Error: %s`", err.Error()))
		return
	}

	newCommits, err := historySince(oldHash)
	if err != nil {
		sendMessage(bot, update, fmt.Sprintf("Pulling worked fine, however I cannot get the commits new with this pull.\n`Error: %s`", err.Error()))
		return
	}

	if len(newCommits) == 0 {
		sendMessage(bot, update, "Repository is already up to date.")
		return
	}

	message := "*New commits:*\n\n"
	for _, commit := range newCommits {
		message += formatCommit(commit)
	}
	sendMessage(bot, update, message)
}

func historyCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	commits, err := history()
	if err != nil {
		sendMessage(bot, update, fmt.Sprintf("An error occoured while reading the repository.\n`Error: %s`", err.Error()))
		return
	}

	message := ""
	for _, commit := range commits {
		message += formatCommit(commit)
	}
	sendMessage(bot, update, message)
}

func helpCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	commands := `
/ls - List all files in a directory
/cat - Print a file context in a chat message
/download - Send a file
/readme - Similar to /cat README.md
/history - Print the git history
/pull - Pull the newest git changes
/help - This help
`
	sendMessage(bot, update, "*A List of what I can do:*\n"+commands)
}

func sendMessage(bot *tgbotapi.BotAPI, update *tgbotapi.Update, text string) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	_, _ = bot.Send(msg)
}

func sendFile(bot *tgbotapi.BotAPI, update *tgbotapi.Update, path string) {
	msg := tgbotapi.NewDocumentUpload(update.Message.Chat.ID, path)
	_, err := bot.Send(msg)
	if err != nil {
		log.Println("Error: ", err.Error())
		sendMessage(bot, update, fmt.Sprintf("Unable to send you the file\n`Error: %s`", err.Error()))
	}
}

func formatCommit(commit gitobject.Commit) string {
	// Get the files from the commit
	var files []string
	stats, err := commit.Stats()
	if err == nil {
		for _, stat := range stats {
			files = append(files, stat.Name)
		}
	}

	// Generate the text for the files
	fileText := ""
	if len(files) == 0 {
		fileText = "_unable to load the files_"
	} else {
		fileText = fmt.Sprintf("\\[%d] `%s`", len(files), strings.Join(files, ", "))
	}

	// Generate the message text where the first line is treated like a header and is in bold, while the rest is normal
	// text
	message := strings.SplitN(strings.TrimSpace(commit.Message), "\n", 1)
	messageText := fmt.Sprintf("*%s*", message[0])
	if len(message) > 1 {
		messageText += "\n" + message[1]
	}

	return fmt.Sprintf("%s\nHash: %s\nAuthor: %s <%s>\nDate: %s\nFiles: %s\n\n",
		messageText,
		commit.Hash,
		commit.Author.Name,
		commit.Author.Email,
		commit.Author.When.Local().Format("02.01.2006 15:04"),
		fileText,
	)
}

func isAdmin(telegramID int64) bool {
	original, _ := strconv.ParseInt(os.Getenv("TELEGRAM_ADMIN"), 10, 64)
	return telegramID == original
}
