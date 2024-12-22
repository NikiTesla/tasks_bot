package telegram

import (
	"context"
	"errors"
	"tasks_bot/internal/domain"
	"tasks_bot/internal/errs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	startMessage = `Добро пожаловать!`
	errorReponse = "Произошла непредвиденная ошибка, уже чиним 🤕"
)

// TODO add context to handlers
func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message) {
	logger := b.logger.WithField("chatID", message.Chat.ID)

	stage, err := b.storage.GetStage(ctx, message.Chat.ID)
	if err != nil {
		if !errors.Is(err, errs.ErrNotFound) {
			logger.WithError(err).Error("failed to get stage")
			return
		}
	}

	switch stage {
	case domain.Unknown:
		b.handleUnknownStage(ctx, message)

	case domain.ContactRequest:
		b.handleContactRequest(ctx, message)

	case domain.Default:
		b.handleStart(ctx, message)

	case domain.BecomeChief, domain.BecomeExecutor, domain.BecomeObserver, domain.BecomeAdmin:
		b.handleBecomeStage(ctx, message, stage)

	case domain.AddTaskName, domain.AddTaskUser, domain.AddTaskDeadline:
		b.handleAddTaskStage(ctx, message, stage)

	case domain.MarkTaskAsClosed, domain.MarkTaskAsDone:
		b.handleMarkTaskStage(ctx, message, stage)

	case domain.DeleteTask:
		b.handleDeleteTaskStage(ctx, message)

	case domain.ChangeDeadline:
		b.handleChangeDeadlineStage(ctx, message)

	default:
		b.handleStart(ctx, message)
	}
}
