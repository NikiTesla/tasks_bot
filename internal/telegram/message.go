package telegram

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"tasks_bot/internal/domain"
	"tasks_bot/internal/errs"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/crypto/bcrypt"
)

const (
	startMessage = `Добро пожаловать!`
	errorReponse = "Произошла непредвиденная ошибка, уже чиним 🤕"
)

// TODO add context to handlers
func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message) {
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

	case domain.AddTaskName, domain.AddTaskUser, domain.AddTaskDeadline:
		b.handleAddTaskStage(ctx, message, stage)

	case domain.MarkTaskAsClosed, domain.MarkTaskAsDone:
		b.handleMarkTask(ctx, message, stage)

	case domain.ChangeDeadline:
		b.handleChangeDeadlineStage(ctx, message)

	default:
		b.handleStart(message)
	}
}

func (b *Bot) handleBecomeStage(message *tgbotapi.Message, stage domain.Stage) {
	logger := b.logger.WithField("chatID", message.Chat.ID)
	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "Ваша роль успешно изменена")
	defer func() {
		if _, err := b.bot.Send(responseMsg); err != nil {
			logger.WithError(err).Error("failed to sent response")
		}
	}()

	delMsg := tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID)
	if _, err := b.bot.Send(delMsg); err != nil {
		logger.WithError(err).Error("failed to delete password message")
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
		logger.WithError(err).Error("failed to set role with")
		responseMsg.Text = errorReponse
	}
	if err := b.storage.SetStage(context.Background(), message.Chat.ID, domain.Default); err != nil {
		logger.WithError(err).Error("failed to set stage")
		responseMsg.Text = errorReponse
	}
}

func (b *Bot) handleAddTaskStage(ctx context.Context, message *tgbotapi.Message, stage domain.Stage) {
	logger := b.logger.WithField("chatID", message.Chat.ID)

	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "Задча успешно создана")
	defer func() {
		if _, err := b.bot.Send(responseMsg); err != nil {
			logger.WithError(err).Error("failed to send response")
		}
	}()

	taskInProgress, err := b.storage.GetTaskInProgress(ctx, message.Chat.ID)
	if err != nil {
		logger.WithError(err).Error("failed to get task in progress for the chat")
		responseMsg.Text = errorReponse
		return
	}

	nextStage := domain.Default
	switch stage {
	case domain.AddTaskName:
		nextStage = domain.AddTaskUser
		if err = b.storage.SetTaskInProgressName(ctx, message.Chat.ID, message.Text); err != nil {
			logger.WithError(err).Error("failed to set task in progress name for the chat")
			responseMsg.Text = errorReponse
			return
		}
		responseMsg.Text = "Введите ник исполнителя в формате @username"

	case domain.AddTaskUser:
		nextStage = domain.AddTaskDeadline
		username := strings.Trim(message.Text, "@")
		if err = b.storage.SetTaskInProgressUser(ctx, message.Chat.ID, username); err != nil {
			logger.WithError(err).Error("failed to set task in progress user for the chat")
			responseMsg.Text = errorReponse
			return
		}
		responseMsg.Text = "Введите дедлайн задачи в формате 21.12.2024 12:20:00"

	case domain.AddTaskDeadline:
		nextStage = domain.Default
		timestamp, err := time.ParseInLocation(domain.DeadlineLayout, message.Text, time.Local)
		if err != nil {
			responseMsg.Text = "Некорректный формат даты-времени, проверьте, что вы вводите дату и время в формате, похожем на 21.12.2024 12:20:00"
			return
		}
		taskInProgress.Deadline = timestamp

		if err = b.storage.AddTask(ctx, taskInProgress); err != nil {
			logger.WithError(err).Error("failed to add task")
			responseMsg.Text = errorReponse
			return
		}
		responseMsg.ParseMode = tgbotapi.ModeHTML
		responseMsg.Text = fmt.Sprintf("Вы успешно добавили задачу: \n\n%s", taskInProgress)

	}

	if err := b.storage.SetStage(ctx, message.Chat.ID, nextStage); err != nil {
		logger.WithError(err).Error("failed to set next stage")
		responseMsg.Text = errorReponse
		return
	}

	if err := b.NotifyObservers(ctx, taskInProgress, message.Chat.ID); err != nil {
		logger.WithError(err).Error("failed to notify observers")
		responseMsg.Text = errorReponse
		return
	}
}

func (b *Bot) handleMarkTask(ctx context.Context, message *tgbotapi.Message, stage domain.Stage) {
	logger := b.logger.WithField("chatID", message.Chat.ID)

	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "")
	defer func() {
		if _, err := b.bot.Send(responseMsg); err != nil {
			logger.WithError(err).Error("failed to send response")
		}
	}()

	taskID, err := strconv.Atoi(message.Text)
	if err != nil {
		responseMsg.Text = "Некорректный номер задачи, должно быть число"
		return
	}

	var newTaskStatus domain.TaskStatus
	switch stage {
	case domain.MarkTaskAsClosed:
		if err := b.storage.MarkTaskAsClosed(ctx, taskID); err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				responseMsg.Text = fmt.Sprintf("Задача с номером %d не найдена", taskID)
				return
			}
			logger.WithError(err).Error("b.storage.MarkTaskAsClosed")
			responseMsg.Text = errorReponse
			return
		}
		newTaskStatus = domain.ClosedTask

	case domain.MarkTaskAsDone:
		if err := b.storage.MarkTaskAsDone(ctx, taskID); err != nil {
			if errors.Is(err, errs.ErrNotFound) {
				responseMsg.Text = fmt.Sprintf("Задача с номером %d не найдена", taskID)
				return
			}
			logger.WithError(err).Error("b.storage.MarkTaskAsDone")
			responseMsg.Text = errorReponse
			return
		}
		newTaskStatus = domain.DoneTask
	}

	if err := b.storage.SetStage(ctx, message.Chat.ID, domain.Default); err != nil {
		logger.WithError(err).Error("failed to set next stage")
		responseMsg.Text = errorReponse
		return
	}
	responseMsg.Text = fmt.Sprintf("Статус задачи успешно изменен на \"%s\"", newTaskStatus)
}

func (b *Bot) handleChangeDeadlineStage(ctx context.Context, message *tgbotapi.Message) {
	logger := b.logger.WithField("chatID", message.Chat.ID)

	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "Задча успешно создана")
	defer func() {
		if _, err := b.bot.Send(responseMsg); err != nil {
			logger.WithError(err).Error("failed to send response")
		}
	}()

	taskIDRaw, deadlineRaw, ok := strings.Cut(message.Text, " ")
	if !ok {
		responseMsg.Text = "Некорректный формат, убедитесь что формат аналогичен \"21 21.12.2024 12:20:00\""
		return
	}
	taskID, err := strconv.Atoi(taskIDRaw)
	if err != nil {
		responseMsg.Text = "Неверный номер задачи"
		return
	}
	deadline, err := time.ParseInLocation(domain.DeadlineLayout, deadlineRaw, time.Local)
	if err != nil {
		responseMsg.Text = "Некорректный формат даты-времени, проверьте, что вы вводите дату и время в формате, похожем на 21.12.2024 12:20:00"
		return
	}

	if err := b.storage.ChangeTaskDeadline(ctx, taskID, deadline); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			responseMsg.Text = fmt.Sprintf("Задача с номером %d не найдена", taskID)
			return
		}
		logger.WithError(err).Error("failed to change task deadline")
		responseMsg.Text = errorReponse
		return
	}

	if err := b.storage.SetStage(ctx, message.Chat.ID, domain.Default); err != nil {
		logger.WithError(err).Error("failed to set next stage")
		responseMsg.Text = errorReponse
		return
	}

	responseMsg.ParseMode = tgbotapi.ModeHTML
	responseMsg.Text = fmt.Sprintf("Дедлайн задачи №%d успешно изменен на %s", taskID, deadline)
}
