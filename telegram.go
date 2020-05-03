package main

import (
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/jasonlvhit/gocron"
	gitobject "gopkg.in/src-d/go-git.v4/plumbing/object"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func handleMessage(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	// Call the right function to handle the command
	switch update.Message.Command() {
	case "ls":
		lsCmd(bot, update)
	case "pull":
		pullCmd(bot, update)
	case "readme":
		readmeCmd(bot, update)
	case "exercise":
		exerciseCmd(bot, update)
	case "subscribe":
		subscribeCmd(bot, update)
	case "unsubscribe":
		unsubscribeCmd(bot, update)
	case "cat":
		catCmd(bot, update)
	case "download":
		downloadCmd(bot, update)
	case "history":
		historyCmd(bot, update)
	case "statistic":
		statisticCmd(bot, update)
	case "start":
		helpCmd(bot, update)
	case "help":
		helpCmd(bot, update)
	default:
		// Ignore non-command messages in group chats
		if update.Message.Chat.Type != "private" && !update.Message.IsCommand() {
			return
		}

		// Send a message to show that the bot is confused
		sendMessage(bot, update, "Sorry, I don't know that command.\nType /help to see what I know.")
	}
}

func handleCallBackQuery(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	log.Printf("[%s] %s", update.CallbackQuery.From.UserName, update.CallbackQuery.Data)

	data := strings.SplitN(update.CallbackQuery.Data, " ", 2)
	switch data[0] {
	case "download":
		// Check if the file exists
		file := data[1]
		err := checkPath(file)
		if err != nil {
			log.Printf("Unable to find file: %s", file)
			return
		}

		// Set the action
		_, _ = bot.Send(tgbotapi.NewChatAction(update.CallbackQuery.Message.Chat.ID, tgbotapi.ChatUploadDocument))

		// Upload a file
		msg := tgbotapi.NewDocumentUpload(update.CallbackQuery.Message.Chat.ID, path.Join(getGitDir(), file))
		msg.Caption = hideSecrets(file)
		_, _ = bot.Send(msg)
	}
}

// This function must not be called as a goroutine because it should block and will return once the automatic background
// jobs are correctly set up
func startBackgroundManager(bot *tgbotapi.BotAPI) {
	// First call the background task to ensure it ran once
	backgroundJob(bot)

	// Setup the automatic call of backgroundJob
	go func() {
		gocron.Every(30).Minutes().Do(backgroundJob, bot)
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
		if strings.Contains(c.Author.Email, getGitUser()) && getGitUser() != "" {
			continue
		}

		newFilteredCommits = append(newFilteredCommits, c)
	}

	// Check if there are new commits to notify
	if len(newFilteredCommits) == 0 {
		log.Println("Backgroundjob ran, no new commits to send the users.")
		return
	}

	// Create a message and send it the user
	message := "*New commits:*🎉🎊\n"
	for _, commit := range newFilteredCommits {
		message += formatCommit(commit)
	}

	// Send the messages to the subscribed users
	subscribed := getUsers()
	for _, subscription := range subscribed {
		msg := tgbotapi.NewMessage(subscription, hideSecrets(message))
		msg.ParseMode = "Markdown"
		_, _ = bot.Send(msg)
	}

	log.Println("Backgroundjob ran, sent the users the updates.")
}

func lsCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	// Only admin is allowed to list files
	if !isAdmin(update.Message.From.ID) {
		sendMessageAdminNeeded(bot, update)
		return
	}

	files, err := listFiles(update.Message.CommandArguments())
	if err != nil {
		sendMessage(bot, update, fmt.Sprintf("An error occoured while listing the files.\n`Error: %s`", err.Error()))
		return
	}

	message := ""
	for _, file := range files {
		message += file + "\n"
	}
	sendMessage(bot, update, message)
}

func catCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	// Only admin is allowed to read files
	if !isAdmin(update.Message.From.ID) {
		sendMessageAdminNeeded(bot, update)
		return
	}

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
	// Only admin is allowed to read files
	if !isAdmin(update.Message.From.ID) {
		sendMessageAdminNeeded(bot, update)
		return
	}

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

func exerciseCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	// If there is an argument we try to parse it as a number
	arguments := update.Message.CommandArguments()
	if arguments != "" {
		number, err := strconv.Atoi(arguments)
		if err != nil {
			sendMessage(bot, update, "The argument musst be a number but was: "+arguments)
			return
		}

		file := fmt.Sprintf("angabe/Aufgabenblatt%d.pdf", number)
		_, err = readFile(file)
		if err != nil {
			sendMessage(bot, update, fmt.Sprintf("There is no exercise %d", number))
			return
		}

		sendFile(bot, update, path.Join(getGitDir(), file))
		return
	}

	// Get all files of angabe
	allFiles, err := listFilesRaw("angabe")
	if err != nil {
		sendMessage(bot, update, fmt.Sprintf("An error occoured while reading the exercise directory.\n`Error: %s`", err.Error()))
		return
	}

	// Save the PDFs to a new list
	files := make([]string, 0, len(allFiles)/2)
	for _, file := range allFiles {
		if strings.HasSuffix(file, ".pdf") {
			files = append(files, file)
		}
	}

	// Build the inline keyboard
	rows := make([][]tgbotapi.InlineKeyboardButton, 0)
	for _, file := range files {
		callback := fmt.Sprintf("download angabe/%s", file)
		row := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(file, callback)}
		rows = append(rows, row)
	}
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	// Show the user all possible exercises
	message := fmt.Sprintf("There are %d exercises:", len(files))
	message = hideSecrets(message)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, message)
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true
	msg.ReplyMarkup = keyboard
	_, _ = bot.Send(msg)

}

func subscribeCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if isUser(update.Message.Chat.ID) {
		sendMessage(bot, update, "This channel is already subscribed")
		return
	}

	err := addUser(update.Message.Chat.ID)
	if err != nil {
		sendMessage(bot, update, fmt.Sprintf("An error occoured while reading adding the subscription.\n`Error: %s`", err.Error()))
		return
	}
	message := "This channel is now subscribed"
	sendMessage(bot, update, message)
}

func unsubscribeCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	if !isUser(update.Message.Chat.ID) {
		sendMessage(bot, update, "This channel was not subscribed")
		return
	}

	err := removeUser(update.Message.Chat.ID)
	if err != nil {
		sendMessage(bot, update, fmt.Sprintf("An error occoured while reading deleting the subscription.\n`Error: %s`", err.Error()))
		return
	}
	message := "This channel is no longer subscribed"
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
	// Get all commits from the repository
	commits, err := history()
	if err != nil {
		sendMessage(bot, update, fmt.Sprintf("An error occoured while reading the repository.\n`Error: %s`", err.Error()))
		return
	}

	// For non-admin users we filter the commits so that they can only see the ones by faculty members
	if !isAdmin(update.Message.From.ID) {
		filtered := make([]gitobject.Commit, 0, len(commits))
		for _, c := range commits {
			// Don't append the commits where the matriculation number appears in
			if strings.Contains(c.Author.Email, getGitUser()) && getGitUser() != "" {
				continue
			}

			filtered = append(filtered, c)
		}
		commits = filtered
	}

	// Split the arguments
	rawarg := strings.TrimSpace(update.Message.CommandArguments())
	args := strings.SplitN(rawarg, " ", 2)

	// By default we return 5 commits, unless the user specifies otherwise
	number := 5
	if n, err := strconv.ParseInt(args[len(args)-1], 10, 0); err == nil {
		number = int(n)
	}

	// By default we only give the user the tail of the commits, unless he specifies the head
	if (args[0] == "head") && len(commits) > number {
		commits = commits[:number]
	} else if len(commits) > number {
		commits = commits[len(commits)-number:]
	}

	// Create the message from the selected commits
	message := ""
	for _, commit := range commits {
		message += formatCommit(commit)
	}
	sendMessage(bot, update, message)
}

func statisticCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	users := len(getUsers())
	message := fmt.Sprintf("Subscribed channels: %d", users)
	sendMessage(bot, update, message)
}

func helpCmd(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	commands := `
/ls - List all files in a directory
/cat - Print a file context in a chat message
/download - Send a file
/readme - Similar to /cat README.md
/exercise - Display the exercise PDFs
/subscribe - Send updates when new exercises get added
/unsubscribe - Unsubscribe from the updates
/history - Send the git history
/statistic - Send some information about the bot
/pull - Pull the newest git changes
/help - This help
`

	if !isAdmin(update.Message.From.ID) {
		commands += "\nUnfortunately, you are not the admin of this bot, so many commands might not work. " +
			"However, you can download the bot at the link below and run it on your own server, " +
			"so that you are the admin of your instance."
	}

	about := `
I was developed by my creator [flofriday](https://github.com/flofriday), and my source is publicly available on [GitHub](https://github.com/flofriday/EP2-Bot) and [GitLab](https://gitlab.com/flofriday/EP2-Bot).
`

	sendMessage(bot, update, fmt.Sprintf("*A List of things I can do:*%s\n%s", commands, about))
}

func sendMessageAdminNeeded(bot *tgbotapi.BotAPI, update *tgbotapi.Update) {
	message := "Sorry, but for security reasons, only the admin is allowed to perform this action.\n\n" +
		"However, there are good news 😄, you can download my code and deploy me on your own server, " +
		"so that you can be the admin:\nhttps://github.com/flofriday/EP2-Bot"
	sendMessage(bot, update, message)
}

func sendMessage(bot *tgbotapi.BotAPI, update *tgbotapi.Update, text string) {
	text = hideSecrets(text)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "Markdown"
	msg.DisableWebPagePreview = true
	_, _ = bot.Send(msg)
}

func sendFile(bot *tgbotapi.BotAPI, update *tgbotapi.Update, path string) {
	err := checkPath(path)
	if err != nil {
		sendMessage(bot, update, fmt.Sprintf("Unable to send you the file\n`Error: %s`", err.Error()))
		return
	}

	// Tell the client that we are uploading a file
	sendAction(bot, update, tgbotapi.ChatUploadDocument)

	// Upload a file
	msg := tgbotapi.NewDocumentUpload(update.Message.Chat.ID, path)
	msg.Caption = hideSecrets(msg.Caption)
	_, err = bot.Send(msg)
	if err != nil {
		log.Println("Error: ", err.Error())
		sendMessage(bot, update, fmt.Sprintf("Unable to send you the file\n`Error: %s`", err.Error()))
	}
}

func sendAction(bot *tgbotapi.BotAPI, update *tgbotapi.Update, action string) {
	_, _ = bot.Send(tgbotapi.NewChatAction(update.Message.Chat.ID, action))
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
	message := strings.SplitN(strings.TrimSpace(commit.Message), "\n", 2)
	messageText := fmt.Sprintf("*%s*", message[0])
	if len(message) > 1 {
		messageText += "\n" + strings.TrimSpace(message[1])
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

func hideSecrets(text string) string {
	text = strings.ReplaceAll(text, os.Getenv("GIT_URL"), "$GIT_URL")

	// Only replace the username if there is a username in the url.
	userName := getGitUser()
	if userName != "" {
		text = strings.ReplaceAll(text, getGitUser(), "$USER")
	}
	return text
}

func isAdmin(telegramID int) bool {
	original, _ := strconv.ParseInt(os.Getenv("TELEGRAM_ADMIN"), 10, 32)
	return telegramID == int(original)
}
