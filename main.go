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
	dbPassword = "***REMOVED***"
	dbName     = "nubipbot"
)

const token = "1675953280:AAEarqSDBOY6M_O1j7AHC_5uua7oHw28Ze4"

var db *sql.DB

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

func serveUser(text string, chatID int64) string {
	parts := strings.Split(text, "\n")
	if len(parts) < 2 {
		return "Неверный формат"
	}

	username := parts[0]
	password := parts[1]

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
