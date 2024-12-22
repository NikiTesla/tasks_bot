package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"tasks_bot/internal/domain"
	"tasks_bot/internal/errs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleStart(ctx context.Context, message *tgbotapi.Message) {
	logger := b.logger.WithField("chatID", message.Chat.ID)

	msg := tgbotapi.NewMessage(message.Chat.ID, startMessage)
	msg.ParseMode = tgbotapi.ModeHTML
	if _, err := b.bot.Send(msg); err != nil {
		logger.WithError(err).Error("failed to sent start message")
		return
	}

	role, err := b.storage.GetRole(ctx, message.From.ID)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		logger.WithError(err).Error("failed to get role")
		return
	}
	if err := b.setCommands(ctx, message.Chat.ID, role); err != nil {
		logger.WithError(err).Error("failed to set commands")
	}
}

func (b *Bot) debugCommand(message *tgbotapi.Message) string {
	storageDump, err := b.storage.DebugStorage(b.logger.Context)
	if err != nil {
		b.logger.WithError(err).Error("failed to debug storage")
		return ""
	}
	var response string
	response += fmt.Sprintf("Your chat id is %d\n", message.Chat.ID)
	response += storageDump

	return response
}

func (b *Bot) handleBecomeCommand(ctx context.Context, message *tgbotapi.Message, command string) {
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

	if err := b.storage.SetStage(ctx, message.Chat.ID, stage); err != nil && !errors.Is(err, errs.ErrNotFound) {
		b.logger.WithError(err).Error("b.storage.SetStage: %w", err)
		responseMsg.Text = errorReponse
		return
	}
	responseMsg.Text = "Введите пароль для идентификации"
}

func (b *Bot) handleGetRoleCommand(ctx context.Context, message *tgbotapi.Message) {
	logger := b.logger.WithField("chatID", message.Chat.ID)

	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "")
	responseMsg.ParseMode = tgbotapi.ModeHTML
	defer func() {
		if _, err := b.bot.Send(responseMsg); err != nil {
			logger.WithError(err).Error("failed to send response")
		}
	}()

	role, err := b.storage.GetRole(context.Background(), message.Chat.ID)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		responseMsg.Text = errorReponse
		return
	}
	responseMsg.Text = fmt.Sprintf("Ваша роль - %s", role)
}

func (b *Bot) handleGetAllTasksCommand(ctx context.Context, message *tgbotapi.Message) {
	logger := b.logger.WithField("chatID", message.Chat.ID)

	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "")
	responseMsg.ParseMode = tgbotapi.ModeHTML
	defer func() {
		if _, err := b.bot.Send(responseMsg); err != nil {
			logger.WithError(err).Error("failed to send response")
		}
	}()

	tasks, err := b.storage.GetAllTasks(ctx)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		logger.WithError(err).Error("failed to get all tasks")
		responseMsg.Text = errorReponse
		return
	}

	if len(tasks) == 0 {
		responseMsg.Text = "Нет добавленных задач"
		return
	}

	builder := strings.Builder{}
	for _, task := range tasks {
		builder.WriteString(task.String())
		builder.WriteString("\n\n")
	}
	responseMsg.Text = builder.String()
}

func (b *Bot) handleGetOpenTasksCommand(ctx context.Context, message *tgbotapi.Message) {
	logger := b.logger.WithField("chatID", message.Chat.ID)

	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "")
	responseMsg.ParseMode = tgbotapi.ModeHTML
	defer func() {
		if _, err := b.bot.Send(responseMsg); err != nil {
			logger.WithError(err).Error("failed to send response")
		}
	}()

	tasks, err := b.storage.GetOpenTasks(ctx)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		logger.WithError(err).Error("failed to get open tasks")
		responseMsg.Text = errorReponse
		return
	}

	if len(tasks) == 0 {
		responseMsg.Text = "Нет открытых задач"
		return
	}

	builder := strings.Builder{}
	for _, task := range tasks {
		builder.WriteString(task.String())
		builder.WriteString("\n\n")
	}
	responseMsg.Text = builder.String()
}

func (b *Bot) handleGetClosedTasksCommand(ctx context.Context, message *tgbotapi.Message) {
	logger := b.logger.WithField("chatID", message.Chat.ID)

	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "")
	responseMsg.ParseMode = tgbotapi.ModeHTML
	defer func() {
		if _, err := b.bot.Send(responseMsg); err != nil {
			logger.WithError(err).Error("failed to send response")
		}
	}()

	tasks, err := b.storage.GetClosedTasks(ctx)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		logger.WithError(err).Error("failed to get closed tasks")
		responseMsg.Text = errorReponse
		return
	}

	if len(tasks) == 0 {
		responseMsg.Text = "Нет зыкрытых задач"
		return
	}

	builder := strings.Builder{}
	for _, task := range tasks {
		builder.WriteString(task.String())
		builder.WriteString("\n\n")
	}
	responseMsg.Text = builder.String()
}

func (b *Bot) handleGetDoneTasksCommand(ctx context.Context, message *tgbotapi.Message) {
	logger := b.logger.WithField("chatID", message.Chat.ID)

	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "")
	responseMsg.ParseMode = tgbotapi.ModeHTML
	defer func() {
		if _, err := b.bot.Send(responseMsg); err != nil {
			logger.WithError(err).Error("failed to send response")
		}
	}()

	tasks, err := b.storage.GetClosedTasks(ctx)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		logger.WithError(err).Error("failed to get closed tasks")
		responseMsg.Text = errorReponse
		return
	}

	if len(tasks) == 0 {
		responseMsg.Text = "Нет выполненных задач"
		return
	}

	builder := strings.Builder{}
	for _, task := range tasks {
		builder.WriteString(task.String())
		builder.WriteString("\n\n")
	}
	responseMsg.Text = builder.String()
}

func (b *Bot) handleGetExpiredTasksCommand(ctx context.Context, message *tgbotapi.Message) {
	logger := b.logger.WithField("chatID", message.Chat.ID)

	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "")
	responseMsg.ParseMode = tgbotapi.ModeHTML
	defer func() {
		if _, err := b.bot.Send(responseMsg); err != nil {
			logger.WithError(err).Error("failed to send response")
		}
	}()

	tasks, err := b.storage.GetExpiredTasks(ctx)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		logger.WithError(err).Error("failed to get expired tasks")
		responseMsg.Text = errorReponse
		return
	}

	if len(tasks) == 0 {
		responseMsg.Text = "Нет просроченных задач"
		return
	}

	builder := strings.Builder{}
	for _, task := range tasks {
		builder.WriteString(task.String())
		builder.WriteString("\n\n")
	}
	responseMsg.Text = builder.String()
}

func (b *Bot) handleGetSelfTasksCommand(ctx context.Context, message *tgbotapi.Message) {
	logger := b.logger.WithField("chatID", message.Chat.ID)

	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "")
	responseMsg.ParseMode = tgbotapi.ModeHTML
	defer func() {
		if _, err := b.bot.Send(responseMsg); err != nil {
			logger.WithError(err).Error("failed to send response")
		}
	}()

	tasks, err := b.storage.GetUserTasks(ctx, message.Chat.UserName)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		logger.WithError(err).Error("failed to get user's tasks")
		responseMsg.Text = errorReponse
		return
	}

	if len(tasks) == 0 {
		responseMsg.Text = "Нет зыкрытых задач"
		return
	}

	builder := strings.Builder{}
	for _, task := range tasks {
		builder.WriteString(task.String())
		builder.WriteString("\n\n")
	}
	responseMsg.Text = builder.String()
}

func (b *Bot) handleAddTaskCommand(ctx context.Context, message *tgbotapi.Message) {
	logger := b.logger.WithField("chatID", message.Chat.ID)

	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "")
	defer func() {
		if _, err := b.bot.Send(responseMsg); err != nil {
			logger.WithError(err).Error("failed to send response")
		}
	}()

	if err := b.storage.SetStage(ctx, message.Chat.ID, domain.AddTaskName); err != nil && !errors.Is(err, errs.ErrNotFound) {
		logger.WithError(err).Error("b.storage.SetStage")
		responseMsg.Text = errorReponse
		return
	}
	responseMsg.Text = `Введите название задачи`
}

func (b *Bot) handleMarkTaskCommands(ctx context.Context, message *tgbotapi.Message, stage domain.Stage) {
	logger := b.logger.WithField("chatID", message.Chat.ID)

	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "")
	defer func() {
		if _, err := b.bot.Send(responseMsg); err != nil {
			logger.WithError(err).Error("failed to send response")
		}
	}()

	if err := b.storage.SetStage(ctx, message.Chat.ID, stage); err != nil && !errors.Is(err, errs.ErrNotFound) {
		logger.WithError(err).Error("b.storage.SetStage")
		responseMsg.Text = errorReponse
		return
	}
	responseMsg.Text = `Введите номер задачи`
}

func (b *Bot) handleChangeTaskDeadlineCommand(ctx context.Context, message *tgbotapi.Message) {
	logger := b.logger.WithField("chatID", message.Chat.ID)

	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "")
	defer func() {
		if _, err := b.bot.Send(responseMsg); err != nil {
			logger.WithError(err).Error("failed to send response")
		}
	}()

	if err := b.storage.SetStage(ctx, message.Chat.ID, domain.ChangeDeadline); err != nil && !errors.Is(err, errs.ErrNotFound) {
		logger.WithError(err).Error("b.storage.SetStage")
		responseMsg.Text = errorReponse
		return
	}
	responseMsg.Text = "Введите номер задачи и новый дедлайн в формате \"21 21.12.2024 12:20:00\""
}
