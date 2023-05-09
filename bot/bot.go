package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const (
	apiUrl = "https://api.telegram.org/bot"
	apiKey = "6270850653:AAHuB6fAv6Vj18SjDxgz5QDg4LrR48UplwI"
)

type Bot struct {
	lastUpdate int
}

func NewBot() *Bot {
	fmt.Println("new bot")
	return &Bot{}
}

type TgKeyboardButton struct {
	Text         string `json:"text"`
	Url          string `json:"url"`
	CallbackData string `json:"callback_data"`
}

// types of data for updates

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
	Ok bool
}

type CallbackQuery struct {
	Id      string
	From    TelegramUser
	Message TelegramMessage
	Data    string
}

// map[string]interface{} !!!
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

	//		resp, err = http.Post(endpoint, "application/json", dataReader)

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

// // отправка сообщений
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

func (d *Bot) NewKeyboard(text, url, cbdata string) *TgKeyboardButton {
	return &TgKeyboardButton{
		Text:         text,
		Url:          url,
		CallbackData: cbdata,
	}
}
