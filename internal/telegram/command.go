package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
)

const (
	startCommand  = "start"
	healthCommand = "healthz"
)

func (b *Bot) handleCommand(message *tgbotapi.Message) {
	logger := b.logger.WithField("chatId", message.Chat.ID).WithField("command", message.Command())

	switch message.Command() {
	case startCommand:
		b.handleStart(message)

	default:
		b.handleAdminCommand(logger, message)
	}
}

func (b *Bot) handleAdminCommand(logger *log.Entry, message *tgbotapi.Message) {
	isAdmin, err := b.storage.IsAdmin(context.Background(), message.Chat.ID)
	if err != nil {
		b.logger.WithError(err).Error("failed to check if is admin")
	}
	if err != nil || !isAdmin {
		if _, err := b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Нет доступа до данного функционала, обратитесь к администратору")); err != nil {
			logger.WithError(err).Error("unable to send response")
		}
		return
	}

	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "")
	switch message.Command() {
	case healthCommand:
		responseMsg.Text = "Status Ok!"

	default:
		responseMsg.Text = "Неизвестная команда, проверьте актуальный список доступных команд"
	}

	responseMsg.ParseMode = tgbotapi.ModeHTML
	if _, err := b.bot.Send(responseMsg); err != nil {
		logger.WithError(err).Error("unable to send response")
	}
}

func (b *Bot) handleStart(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, startMessage)
	if _, err := b.bot.Send(msg); err != nil {
		b.logger.WithError(err).WithField("chatID", message.Chat.ID).Error("failed to sent start message")
	}
}
