package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/UQuark0/nubipbot/nubip"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/lib/pq"
)

const (
	dbHost     = "37.57.205.178"
	dbPort     = 5432
	dbUser     = "developer"
	dbPassword = ""
	dbName     = "nubipbot"
)

const token = "1675953280:AAEarqSDBOY6M_O1j7AHC_5uua7oHw28Ze4"

var db *sql.DB

var bot *tgbotapi.BotAPI

type userData struct {
	id       int
	username string
	password string
	chatID   int64
	isAdmin  bool
}

func getUserByUsername(username string) (*userData, string) {
	row := db.QueryRow(`SELECT * FROM "user" WHERE nubip_username = $1`, username)
	var user userData
	err := row.Scan(&user.id, &user.username, &user.password, &user.chatID, &user.isAdmin)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, "Пользователь не найден"
		}
		return nil, err.Error()
	}

	return &user, ""
}

func isAdmin(chatID int64) bool {
	row := db.QueryRow(`SELECT * FROM "admin" WHERE chat_id = $1`, chatID)
	err := row.Scan()
	return err == nil
}

func registerUser(username string, password string, chatID int64) (*userData, string) {
	_, err := nubip.NewNubipAPI(username, password)
	if err != nil {
		return nil, err.Error()
	}

	_, err = db.Exec(`INSERT INTO "user" (nubip_username, nubip_password, chat_id) VALUES ($1, $2, $3)`, username, password, chatID)
	if err != nil {
		return nil, err.Error()
	}

	user, _ := getUserByUsername(username)
	return user, "Пользователь добавлен"
}

func deleteUser(id int) string {
	_, err := db.Exec(`DELETE FROM "user" WHERE id = $1`, id)
	if err == nil {
		return "Пользователь удалён"
	}
	return err.Error()
}

func raidContest(contestID string) error {
	rows, err := db.Query(`SELECT * FROM "user"`)
	if err != nil {
		return err
	}

	for rows.Next() {
		var user userData
		err = rows.Scan(&user.id, &user.username, &user.password, &user.chatID, &user.isAdmin)
		if err != nil {
			continue
		}
		n, err := nubip.NewNubipAPI(user.username, user.password)
		if err != nil {
			continue
		}
		err = n.LoginContest(contestID)
		if err != nil {
			continue
		}
		err = n.SendHelloWorld()
		if err != nil {
			continue
		}
		msg := tgbotapi.NewMessage(user.chatID, "Дед получил от тебя Hello World")
		bot.Send(msg)
	}

	return nil
}

func serveUser(text string, chatID int64) string {
	parts := strings.Split(text, "\n")
	if len(parts) < 1 {
		return "Неверный формат"
	}

	cmd := parts[0]
	if cmd == "user" {
		if len(parts) < 3 {
			return "Неверный формат"
		}

		username := parts[1]
		password := parts[2]

		user, _ := getUserByUsername(username)

		if user == nil {
			_, result := registerUser(username, password, chatID)
			return result
		}

		if password == user.password {
			result := deleteUser(user.id)
			return result
		}

		return "Неверный пароль"
	}

	if cmd == "contest" {
		if len(parts) < 2 {
			return "Неверный формат"
		}

		if !isAdmin(chatID) {
			return "Недостаточно прав"
		}

		contest := parts[1]
		err := raidContest(contest)
		if err != nil {
			return err.Error()
		}
		return "Набег завершён"
	}

	return "Неверный формат"
}

func main() {
	var err error

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates, err := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message == nil || update.Message.Text == "" {
			continue
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, serveUser(update.Message.Text, update.Message.Chat.ID))
		bot.Send(msg)
	}
}
