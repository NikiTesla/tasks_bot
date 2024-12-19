package telegram

import (
	"context"
	"tasks_bot/internal/domain"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/crypto/bcrypt"
)

func (b *Bot) handleBecomeStage(message *tgbotapi.Message, stage domain.Stage) {
	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "Ваша роль успешно изменена")
	defer func() {
		if _, err := b.bot.Send(responseMsg); err != nil {
			b.logger.WithError(err).WithField("chatID", message.Chat.ID).Error("failed to sent response")
		}
	}()

	delMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID)
	if _, err := b.bot.Send(delMsg); err != nil {
		b.logger.WithError(err).WithField("chatID", message.Chat.ID).Error("failed to delete password message")
	}

	var role domain.Role
	switch stage {
	case domain.BecomeChief:
		if bcrypt.CompareHashAndPassword(chiefPasswordHash, []byte(message.Text)) != nil {
			responseMsg.Text = "Вы ввели неверный пароль. Попробуйте ещё"
			return
		}
		role = domain.Chief

	case domain.BecomeExecutor:
		if bcrypt.CompareHashAndPassword(executorPasswordHash, []byte(message.Text)) != nil {
			responseMsg.Text = "Вы ввели неверный пароль. Попробуйте ещё"
			return
		}
		role = domain.Executor

	case domain.BecomeObserver:
		if bcrypt.CompareHashAndPassword(observerPasswordHash, []byte(message.Text)) != nil {
			responseMsg.Text = "Вы ввели неверный пароль. Попробуйте ещё"
			return
		}
		role = domain.Observer

	case domain.BecomeAdmin:
		if bcrypt.CompareHashAndPassword(adminPasswordHash, []byte(message.Text)) != nil {
			responseMsg.Text = "Вы ввели неверный пароль. Попробуйте ещё"
			return
		}
		role = domain.Admin
	}

	if err := b.storage.SetRole(context.Background(), message.Chat.ID, role); err != nil {
		b.logger.WithError(err).Errorf("failed to set role with id: %d", message.Chat.ID)
		responseMsg.Text = errorReponse
	}
	if err := b.storage.SetStage(context.Background(), message.Chat.ID, domain.Default); err != nil {
		b.logger.WithError(err).Errorf("failed to set stage for id: %d", message.Chat.ID)
		responseMsg.Text = errorReponse
	}
}
