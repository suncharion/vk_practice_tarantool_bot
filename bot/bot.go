package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
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
// пароли сохраняются нешифрованными, т.к. это тестовое задание, но в целом это
// очень плохо и в продакшене так делать нельзя
func (b *Bot) Set(user int, service, pass string) error {
	tmp, err := b.db.Query(context.Background(), `insert into 
		passwords (user_id, service, password) 
		values ($1, $2, $3) 
		ON CONFLICT (user_id, service) 
		DO UPDATE SET password = excluded.password;`, user, service, pass)
	tmp.Close()
	return err
}

// получает из БД значения паролей
// пароли сохраняются нешифрованными, т.к. это тестовое задание, но в целом это
// очень плохо и в продакшене так делать нельзя
func (b *Bot) Get(user int, service string) (string, error) {
	pass := ""
	err := b.db.QueryRow(context.Background(), "select password from passwords where user_id = $1 and service = $2 order by id desc limit 1", user, service).Scan(&pass)
	if err != nil {
		return "", err
	}
	return pass, nil
}

// получает из БД список сервисов
func (b *Bot) GetList(user int) ([]string, error) {
	rows, err := b.db.Query(context.Background(), "select service from passwords where user_id = $1", user)
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()

	services := []string{}

	for rows.Next() {
		servName := ""
		err = rows.Scan(&servName)
		if err != nil {
			return []string{}, err
		}
		services = append(services, servName)
	}

	return services, nil
}

// удаляет из БД значения паролей
func (b *Bot) Del(user int, service string) error {
	rows, err := b.db.Query(context.Background(), "delete from passwords where user_id = $1 and service = $2 returning *", user, service)
	if err != nil {
		return err
	}
	defer rows.Close()
	if rows.Next() {
		return nil
	}
	return fmt.Errorf("Nothing to delete")
}

func (b *Bot) DeleteMessage(chatId, messageId int) {
	time.Sleep(time.Second * 10)
	_, _ = b.Query("deleteMessage", "POST", map[string]interface{}{
		"chat_id":    chatId,
		"message_id": messageId,
	})
}

func (b *Bot) UpdateWebhook(w http.ResponseWriter, req *http.Request) {
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Println("Error reading webhook data", err)
	}
	fmt.Println("webhook update", string(data))
	var update TelegramUpdate
	err = json.Unmarshal([]byte(data), &update)
	if err != nil {
		fmt.Println("Cannot parse JSON in update")
		return
	}

	// обрабатываем обновления, если они есть

	// разбиваем входящее сообщение на слова
	messageParts := strings.Split(update.Message.Text, " ")

	if update.Message.Text == "/start" || update.CallbackQuery.Data == "/start" { // приветствие пользователя
		// в зависимости от того, пришел пользватель с кнопки inline-клавиатуры или встроенной, ищем chatId в соответствующих местах
		var chatId int
		if update.Message.Text != "" {
			chatId = update.Message.Chat.Id
		} else {
			chatId = update.CallbackQuery.Message.Chat.Id
		}

		if chatId == 0 {
			fmt.Println("Cannot get chat id for message")
		}

		_, err := b.SendMessage(chatId, "Менеджер паролей \n Доступные команды: /set /get /del /list \n Формат ввода команд : \n /set service password \n /get service \n /del service \n /list", nil)
		if err != nil {
			fmt.Println(err)
		}
	} else if messageParts[0] == "/set" { // сохраняем пароль
		if len(messageParts) != 3 {
			_, _ = b.SendMessage(update.Message.Chat.Id, "Неверный формат команды", nil)
			return
		}

		err = b.Set(update.Message.Chat.Id, messageParts[1], messageParts[2])
		if err != nil {
			fmt.Println(err)
			b.SendMessage(update.Message.Chat.Id, "Произошла ошибка при сохранении пароля. ", nil)
			return
		}
		_, _ = b.SendMessage(update.Message.Chat.Id, "Сохранён пароль для сервиса "+messageParts[1], nil)
		go b.DeleteMessage(update.Message.Chat.Id, update.Message.MessageId)

	} else if messageParts[0] == "/get" { // получаем пароль
		if len(messageParts) != 2 {
			_, _ = b.SendMessage(update.Message.Chat.Id, "Неверный формат команды", nil)
			return
		}
		val, err := b.Get(update.Message.Chat.Id, messageParts[1])
		if err != nil {
			fmt.Println(err)
			_, _ = b.SendMessage(update.Message.Chat.Id, "Пароль для сервиса "+messageParts[1]+" не найден.", nil)
			return
		}
		answer, err := b.SendMessage(update.Message.Chat.Id, "Ваш пароль от сервиса "+messageParts[1]+" : \n"+val, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("%+v", answer)
		if answer.Ok {
			fmt.Println("deleting msg", answer.Message.Chat.Id, answer.Message.MessageId)
			go b.DeleteMessage(update.Message.Chat.Id, answer.Message.MessageId)
		}
	} else if messageParts[0] == "/del" { // удаляем пароль
		if len(messageParts) != 2 {
			_, _ = b.SendMessage(update.Message.Chat.Id, "Неверный формат команды", nil)
			return
		}
		err = b.Del(update.Message.Chat.Id, messageParts[1])
		if err != nil {
			fmt.Println(err)
			_, _ = b.SendMessage(update.Message.Chat.Id, "Имя сервиса "+messageParts[1]+" не найдено", nil)
			return
		}
		b.SendMessage(update.Message.Chat.Id, "Сервис "+messageParts[1]+" удалён ", nil)
	} else if messageParts[0] == "/list" { // список сохраненных сервисов
		servNames, err := b.GetList(update.Message.Chat.Id)
		if err != nil {
			fmt.Println(err)
			b.SendMessage(update.Message.Chat.Id, "Ошибка при получении списка сервисов", nil)
			return
		}
		text := fmt.Sprintf("Список сохраненных сервисов:\n%s", strings.Join(servNames, "\n"))
		b.SendMessage(update.Message.Chat.Id, text, nil)
	} else {
		b.SendMessage(update.Message.Chat.Id, "Введите /start для запуска бота", nil)
	}

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

// long polling версия, не используется
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
