package main

import (
	"fmt"
	"go_bot_tg2/bot"
	"log"
	"net/http"
)

func main() {
	// инициализируем бота
	tgbot := bot.NewBot()

	// пингуем API
	_, err := tgbot.GetMe()
	if err != nil {
		fmt.Println(err)
		log.Fatal("Cannot ping Telegram API")
	}

	// запускаем https сервер для вебхука с самоподписанным сертификатом
	http.HandleFunc("/tgwebhook2", tgbot.UpdateWebhook)
	err = http.ListenAndServeTLS(":443", "YOURPUBLIC.pem", "YOURPRIVATE.key", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
