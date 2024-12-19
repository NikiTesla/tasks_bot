package telegram

import (
	"errors"
	"tasks_bot/internal/domain"
	"tasks_bot/internal/errs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	startMessage = `Добро пожаловать!`
	errorReponse = "Произошла непредвиденная ошибка, уже чиним 🤕"
)

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	stage, err := b.storage.GetStage(b.logger.Context, message.Chat.ID)
	if err != nil {
		if !errors.Is(err, errs.ErrNotFound) {
			b.logger.WithError(err).Error("failed to get stage")
			return
		}
	}

	switch stage {
	case domain.Unknown:
		role, err := b.storage.GetRole(b.logger.Context, message.Chat.ID)
		if err != nil {
			if !errors.Is(err, errs.ErrNotFound) {
				b.logger.WithError(err).Error("b.storage.GetRole: %w", err)
			}
		}

		if err := b.storage.AddChat(b.logger.Context, message.Chat.ID, message.Chat.UserName, role); err != nil {
			b.logger.WithError(err).Error("failed to add chat")
			return
		}
		b.handleStart(message)

	case domain.Default:
		b.handleStart(message)

	case domain.BecomeChief, domain.BecomeExecutor, domain.BecomeObserver, domain.BecomeAdmin:
		b.handleBecomeStage(message, stage)

	default:
		b.handleStart(message)
	}
}
