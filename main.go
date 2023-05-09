package main

import (
	"fmt"
	"go_bot_tg2/bot"
	"strings"
	"time"
)

var passwords map[string]string = map[string]string{}

func main() {
	tgbot := bot.NewBot()
	_, err := tgbot.GetMe()
	if err != nil {
		fmt.Println(err)
	}
	// основной бесконечный цикл приложения
	for {
		updates, err := tgbot.GetUpdates()
		if err != nil {
			fmt.Println(err)
		}
		// /start
		for _, update := range updates.Result {
			fmt.Printf("update: %+v\n", update)
			fmt.Println("cbdata", update.CallbackQuery.Data)

			/*if update.CallbackQuery.Data == "pos4" {

				var keyboard []*bot.TgKeyboardButton
				keyboard = append(keyboard, bot.NewBot().NewKeyboard("Ваша реклама", "https://vk.com", ""))
				keyboard = append(keyboard, bot.NewBot().NewKeyboard("Ваша реклама", "https://vk.com", ""))
				keyboard = append(keyboard, bot.NewBot().NewKeyboard("назад", "", "/start"))
				tgbot.SendMessage(update.CallbackQuery.Message.Chat.Id, "Ваша реклама", keyboard)

			} */

			/*
			 /set github password
			*/
			messageParts := strings.Split(update.Message.Text, " ")
			fmt.Println("message parts", messageParts)

			if update.Message.Text == "/start" || update.CallbackQuery.Data == "/start" {
				var chatId int
				if update.Message.Text != "" {
					chatId = update.Message.Chat.Id
				} else {
					chatId = update.CallbackQuery.Message.Chat.Id
				}

				_, err := tgbot.SendMessage(chatId, "Менеджер паролей \n Доступные команды: /set /get /del \n Формат ввода команд : \n /set service password \n /get service \n /del service", nil)
				if err != nil {
					fmt.Println(err)
				}
			} else if messageParts[0] == "/set" {
				// /set github qweqweqwe
				if len(messageParts) != 3 {
					_, _ = tgbot.SendMessage(update.Message.Chat.Id, "Неверный формат команды", nil)
					continue
				}

				_, _ = tgbot.SendMessage(update.Message.Chat.Id, "Сохраняем пароль для сервиса "+messageParts[1], nil)
				passwords[messageParts[1]] = messageParts[2]
			} else if messageParts[0] == "/get" {
				if len(messageParts) != 2 {
					_, _ = tgbot.SendMessage(update.Message.Chat.Id, "Неверный формат команды", nil)
					continue
				}
				val, ok := passwords[messageParts[1]]
				if !ok {
					_, _ = tgbot.SendMessage(update.Message.Chat.Id, "Пароль для сервиса "+messageParts[1]+" не найден", nil)
					continue
				}
				tgbot.SendMessage(update.Message.Chat.Id, "Ваш пароль от сервиса, "+messageParts[1]+" : \n"+val, nil)
			} else if messageParts[0] == "/del" {
				if len(messageParts) != 2 {
					_, _ = tgbot.SendMessage(update.Message.Chat.Id, "Неверный формат команды", nil)
					continue
				}
				_, ok := passwords[messageParts[1]]
				if !ok {
					_, _ = tgbot.SendMessage(update.Message.Chat.Id, "Имя сервиса "+messageParts[1]+" не найдено", nil)
					continue
				}
				delete(passwords, messageParts[1])
				tgbot.SendMessage(update.Message.Chat.Id, "Сервис "+messageParts[1]+" удалён ", nil)

			} else {
				tgbot.SendMessage(update.Message.Chat.Id, "Введите /start для запуска бота", nil)
			}
			// callback_query
			// /get github

		}

		time.Sleep(time.Second * 1)
	}
}
