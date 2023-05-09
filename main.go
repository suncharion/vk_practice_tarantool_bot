package main

import (
	"fmt"
	"go_bot_tg2/bot"
	"log"
	"strings"
	"time"
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
	// основной бесконечный цикл приложения
	for {
		// получаем обновления
		updates, err := tgbot.GetUpdates()
		if err != nil {
			fmt.Println(err)
		}

		// обрабатываем обновления, если они есть
		for _, update := range updates.Result {
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

				_, err := tgbot.SendMessage(chatId, "Менеджер паролей \n Доступные команды: /set /get /del \n Формат ввода команд : \n /set service password \n /get service \n /del service", nil)
				if err != nil {
					fmt.Println(err)
				}
			} else if messageParts[0] == "/set" { // сохраняем пароль
				if len(messageParts) != 3 {
					_, _ = tgbot.SendMessage(update.Message.Chat.Id, "Неверный формат команды", nil)
					continue
				}

				err = tgbot.Set(update.Message.Chat.Id, messageParts[1], messageParts[2])
				if err != nil {
					fmt.Println(err)
					tgbot.SendMessage(update.Message.Chat.Id, "Произошла ошибка при сохранении пароля. ", nil)
					continue
				}
				_, _ = tgbot.SendMessage(update.Message.Chat.Id, "Сохранён пароль для сервиса "+messageParts[1], nil)
			} else if messageParts[0] == "/get" { // получаем пароль
				if len(messageParts) != 2 {
					_, _ = tgbot.SendMessage(update.Message.Chat.Id, "Неверный формат команды", nil)
					continue
				}
				val, err := tgbot.Get(update.Message.Chat.Id, messageParts[1])
				if err != nil {
					fmt.Println(err)
					_, _ = tgbot.SendMessage(update.Message.Chat.Id, "Пароль для сервиса "+messageParts[1]+" не найден", nil)
					continue
				}
				answer, err := tgbot.SendMessage(update.Message.Chat.Id, "Ваш пароль от сервиса "+messageParts[1]+" : \n"+val, nil)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Printf("%+v", answer)
				if answer.Ok {
					fmt.Println("deleting msg", answer.Message.Chat.Id, answer.Message.MessageId)
					go tgbot.DeleteMessage(update.Message.Chat.Id, answer.Message.MessageId)
				}
			} else if messageParts[0] == "/del" { // удаляем пароль
				if len(messageParts) != 2 {
					_, _ = tgbot.SendMessage(update.Message.Chat.Id, "Неверный формат команды", nil)
					continue
				}
				err = tgbot.Del(update.Message.Chat.Id, messageParts[1])
				if err != nil {
					fmt.Println(err)
					_, _ = tgbot.SendMessage(update.Message.Chat.Id, "Имя сервиса "+messageParts[1]+" не найдено", nil)
					continue
				}
				tgbot.SendMessage(update.Message.Chat.Id, "Сервис "+messageParts[1]+" удалён ", nil)
			} else {
				tgbot.SendMessage(update.Message.Chat.Id, "Введите /start для запуска бота", nil)
			}

		}

		// опрос идет через long polling, поэтому основной цикл надо притормаживать на 1 сек, чтобы он не закидывал API запросами
		time.Sleep(time.Second * 1)
	}
}
