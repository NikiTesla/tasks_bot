package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	startMessage = `Добро пожаловать!`
	errorReponse = "Произошла непредвиденная ошибка, уже чиним 🤕"
)

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	b.handleStart(message)
}
