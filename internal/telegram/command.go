package telegram

import (
	"context"
	"fmt"
	"tasks_bot/internal/domain"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
)

const (
	startCmd          = "start"
	becomeExecutorCmd = "become_executor"
	becomeObserverCmd = "become_observer"
	becomeChiefCmd    = "become_chief"
	becomeAdminCmd    = "become_admin"
	getRoleCmd        = "get_role"
	// admin commands
	healthCmd    = "healthz"
	debugStorage = "debug"
)

func (b *Bot) handleCommand(message *tgbotapi.Message) {
	logger := b.logger.WithField("chatId", message.Chat.ID).WithField("command", message.Command())

	switch message.Command() {
	case startCmd:
		b.handleStart(message)

	case becomeExecutorCmd, becomeAdminCmd, becomeChiefCmd, becomeObserverCmd:
		b.handleBecomeCommand(message, message.Command())

	case getRoleCmd:
		b.handleGetRoleCommand(message)

	default:
		b.handleAdminCommand(logger, message)
	}
}

func (b *Bot) handleAdminCommand(logger *log.Entry, message *tgbotapi.Message) {
	role, err := b.storage.GetRole(context.Background(), message.Chat.ID)
	if err != nil {
		b.logger.WithError(err).Error("failed to check if is admin")
	}
	if err != nil || role != domain.Admin {
		if _, err := b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Нет доступа до данного функционала, обратитесь к администратору")); err != nil {
			logger.WithError(err).Error("unable to send response")
		}
		return
	}

	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "")
	switch message.Command() {
	case healthCmd:
		responseMsg.Text = "Status Ok!"

	case debugStorage:
		b.debug(message)

	default:
		responseMsg.Text = "Неизвестная команда, проверьте актуальный список доступных команд"
	}

	responseMsg.ParseMode = tgbotapi.ModeHTML
	if _, err := b.bot.Send(responseMsg); err != nil {
		logger.WithError(err).Error("unable to send response")
	}
}

func (b *Bot) handleStart(message *tgbotapi.Message) {
	_ = b.storage.SetStage(context.Background(), message.Chat.ID, domain.Default)

	msg := tgbotapi.NewMessage(message.Chat.ID, startMessage)
	if _, err := b.bot.Send(msg); err != nil {
		b.logger.WithError(err).WithField("chatID", message.Chat.ID).Error("failed to sent start message")
	}
}

func (b *Bot) debug(message *tgbotapi.Message) {
	storageDump := b.storage.DebugStorage(b.logger.Context)
	var response string
	response += fmt.Sprintf("Your chat id is %d\n", message.Chat.ID)
	response += storageDump

	if _, err := b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, response)); err != nil {
		b.logger.WithError(err).WithField("chatID", message.Chat.ID).Error("failed to sent start message")
	}
}

func (b *Bot) handleBecomeCommand(message *tgbotapi.Message, command string) {
	var stage domain.Stage
	switch command {
	case becomeAdminCmd:
		stage = domain.BecomeAdmin
	case becomeChiefCmd:
		stage = domain.BecomeChief
	case becomeExecutorCmd:
		stage = domain.BecomeExecutor
	case becomeObserverCmd:
		stage = domain.BecomeObserver
	}

	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "")
	defer func() {
		if _, err := b.bot.Send(responseMsg); err != nil {
			b.logger.WithError(err).WithField("chatID", message.Chat.ID).Error("failed to send response")
		}
	}()

	if err := b.storage.SetStage(b.logger.Context, message.Chat.ID, stage); err != nil {
		b.logger.WithError(err).Error("b.storage.SetStage: %w", err)
		responseMsg.Text = errorReponse
		return
	}
	responseMsg.Text = "Введите пароль для идентификации"
}

func (b *Bot) handleGetRoleCommand(message *tgbotapi.Message) {
	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "")
	defer func() {
		if _, err := b.bot.Send(responseMsg); err != nil {
			b.logger.WithError(err).WithField("chatID", message.Chat.ID).Error("failed to send response")
		}
	}()

	role, err := b.storage.GetRole(context.Background(), message.Chat.ID)
	if err != nil {
		responseMsg.Text = errorReponse
		return
	}
	responseMsg.Text = fmt.Sprintf("Ваша роль - %s", role)
}
