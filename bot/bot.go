package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	apiUrl = "https://api.telegram.org/bot"
	apiKey = "6270850653:AAHuB6fAv6Vj18SjDxgz5QDg4LrR48UplwI"
)

type Bot struct {
	lastUpdate int
	db         *pgx.Conn
}

func NewBot() *Bot {
	fmt.Println("new bot")
	bot := &Bot{}

	conn, err := pgx.Connect(context.Background(), "postgres://postgres@90.156.210.232:5432/passwords?password=example")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	bot.db = conn
	fmt.Println("Database connection established")
	return bot
}

type TgKeyboardButton struct {
	Text         string `json:"text"`
	Url          string `json:"url"`
	CallbackData string `json:"callback_data"`
}

type GetUpdatesAnswer struct {
	Ok     bool
	Result []TelegramUpdate
}

type TelegramUpdate struct {
	UpdateId      int             `json:"update_id"`
	Message       TelegramMessage `json:"message"`
	CallbackQuery CallbackQuery   `json:"callback_query"`
}

type TelegramMessage struct {
	MessageId int          `json:"message_id"`
	Text      string       `json:"text"`
	Chat      TelegramChat `json:"chat"`
	From      TelegramUser `json:"from"`
}

type TelegramUser struct {
	Id int `json:"id"`
}

type TelegramChat struct {
	Id int `json:"id"`
}

type MessageAnswer struct {
	Ok      bool
	Message TelegramMessage `json:"result"`
}

type CallbackQuery struct {
	Id      string
	From    TelegramUser
	Message TelegramMessage
	Data    string
}

// отправляет в БД значения паролей
func (b *Bot) Set(user int, service, pass string) error {
	tmp, err := b.db.Query(context.Background(), "insert into passwords (user_id, service, password) values ($1, $2, $3)", user, service, pass)
	tmp.Close()
	return err
}

// получает из БД значения паролей
func (b *Bot) Get(user int, service string) (string, error) {
	pass := ""
	err := b.db.QueryRow(context.Background(), "select password from passwords where user_id = $1 and service = $2 order by id desc limit 1", user, service).Scan(&pass)
	if err != nil {
		return "", err
	}
	return pass, nil
}

// удаляет из БД значения паролей
func (b *Bot) Del(user int, service string) error {
	rows, err := b.db.Query(context.Background(), "delete from passwords where user_id = $1 and service = $2", user, service)
	if err != nil {
		return err
	}
	rows.Close()
	return nil
}

func (b *Bot) DeleteMessage(chatId, messageId int) {
	time.Sleep(time.Second * 10)
	_, _ = b.Query("deleteMessage", "POST", map[string]interface{}{
		"chat_id":    chatId,
		"message_id": messageId,
	})
}

// совершает запросы методов апи телеграм
func (b *Bot) Query(method string, methodtype string, data map[string]interface{}) (string, error) {
	var resultRaw *http.Response
	var err error

	dataJSON, _ := json.Marshal(data)
	dataReader := bytes.NewBuffer(dataJSON)
	if methodtype == "GET" {
		resultRaw, err = http.Get(apiUrl + apiKey + "/" + method)
	} else {
		resultRaw, err = http.Post(apiUrl+apiKey+"/"+method, "application/json", dataReader)
	}

	if err != nil {
		return "", err
	}
	result, _ := io.ReadAll(resultRaw.Body)
	return string(result), nil
}
func (a *Bot) GetMe() (string, error) {
	result, _ := a.Query("getMe", "GET", map[string]interface{}{})
	return result, nil
}

func (a *Bot) GetUpdates() (*GetUpdatesAnswer, error) {
	result, _ := a.Query("getUpdates?offset="+strconv.Itoa(a.lastUpdate), "GET", map[string]interface{}{})
	var parsed GetUpdatesAnswer
	fmt.Printf("%+v\n\n", result)
	err := json.Unmarshal([]byte(result), &parsed)
	if err != nil {
		return nil, err
	}
	if len(parsed.Result) > 0 {
		a.lastUpdate = parsed.Result[len(parsed.Result)-1].UpdateId + 1
	}
	return &parsed, nil
}

// отправка сообщений
func (c *Bot) SendMessage(chat_id int, text string, keyboard []*TgKeyboardButton) (*MessageAnswer, error) {
	data := map[string]interface{}{
		"chat_id": chat_id,
		"text":    text,
	}
	if keyboard != nil {
		data["reply_markup"] = map[string]interface{}{
			"inline_keyboard": [][]*TgKeyboardButton{
				keyboard,
			},
		}
	}

	result, err := c.Query("sendMessage", "POST", data)
	if err != nil {
		return nil, err
	}
	fmt.Println(result)
	var parsed MessageAnswer
	err = json.Unmarshal([]byte(result), &parsed)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// инициализируем объект клавиатуры
func (d *Bot) NewKeyboard(text, url, cbdata string) *TgKeyboardButton {
	return &TgKeyboardButton{
		Text:         text,
		Url:          url,
		CallbackData: cbdata,
	}
}
