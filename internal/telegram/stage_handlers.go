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

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"

	"golang.org/x/crypto/bcrypt"
)

func (b *Bot) handleUnknownStage(ctx context.Context, message *tgbotapi.Message) {
	logger := b.logger.WithField("chatID", message.Chat.ID)

	var phone string
	if message.Contact == nil {
		contactButton := tgbotapi.KeyboardButton{
			Text:           "Поделиться номером телефона",
			RequestContact: true,
		}
		msg := tgbotapi.NewMessage(message.Chat.ID, "Пожалуйста поделитесь своим номером телефона. Это можно сделать с помощью соответствующего пункта в меню")
		msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(contactButton),
		)
		if _, err := b.bot.Send(msg); err != nil {
			logger.WithError(err).Error("failed to send button request")
			return
		}
	} else {
		phone = message.Contact.PhoneNumber
	}

	role, err := b.storage.GetRole(ctx, message.Chat.ID)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		logger.WithError(err).Error("failed to get role: %w", err)
	}
	if err := b.storage.AddChat(ctx, message.Chat.ID, message.Chat.UserName, phone, role); err != nil {
		logger.WithError(err).Error("failed to update chat info with phone number")
		return
	}

	// setting stage either to get phone number and save or default stage to continue work with bot
	if message.Contact == nil {
		if err := b.storage.SetStage(ctx, message.Chat.ID, domain.ContactRequest); err != nil && !errors.Is(err, errs.ErrNotFound) {
			logger.WithError(err).Error("failed to set contact request stage")
		}
	} else {
		if err := b.storage.SetStage(ctx, message.Chat.ID, domain.Default); err != nil && !errors.Is(err, errs.ErrNotFound) {
			logger.WithError(err).Error("failed to set contact request stage")
		}
	}
}

func (b *Bot) handleContactRequest(ctx context.Context, message *tgbotapi.Message) {
	logger := b.logger.WithField("chatID", message.Chat.ID)

	msg := tgbotapi.NewMessage(message.Chat.ID, "Спасибо!")
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	if _, err := b.bot.Send(msg); err != nil {
		logger.WithError(err).Error("failed to remove keyboard")
		return
	}
	if message.Contact != nil {
		if err := b.storage.AddChat(ctx, message.Chat.ID, message.Chat.UserName, message.Contact.PhoneNumber, domain.UnknownRole); err != nil {
			logger.WithError(err).Error("failed to update chat info with phone number")
			return
		}
	}
	if err := b.storage.SetStage(ctx, message.Chat.ID, domain.Default); err != nil {
		logger.WithError(err).Error("failed to set default stage")
		return
	}
}

func (b *Bot) handleBecomeStage(ctx context.Context, message *tgbotapi.Message, stage domain.Stage) {
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

	if err := b.storage.SetRole(ctx, message.Chat.ID, role); err != nil && !errors.Is(err, errs.ErrNotFound) {
		logger.WithError(err).Error("failed to set role with")
		responseMsg.Text = errorReponse
	}
	if err := b.storage.SetStage(ctx, message.Chat.ID, domain.Default); err != nil && !errors.Is(err, errs.ErrNotFound) {
		logger.WithError(err).Error("failed to set stage")
		responseMsg.Text = errorReponse
	}
	if err := b.setCommands(ctx, message.Chat.ID, role); err != nil {
		logger.WithError(err).Error("failed to set commands")
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

	nextStage := domain.Default
	switch stage {
	case domain.AddTaskName:
		nextStage = domain.AddTaskUser
		if err := b.storage.SetTaskInProgressName(ctx, message.Chat.ID, message.Text); err != nil {
			logger.WithError(err).Error("failed to set task in progress name for the chat")
			responseMsg.Text = errorReponse
			return
		}
		responseMsg.Text = "Введите ник исполнителя в формате @username или телефон в формате 79xxxxxxxxx"

	case domain.AddTaskUser:
		nextStage = domain.AddTaskDeadline
		userContact := strings.Trim(message.Text, "@")

		executorChat, err := b.storage.GetChat(ctx, userContact, userContact)
		if err != nil && !errors.Is(err, errs.ErrNotFound) {
			responseMsg.Text = errorReponse
			logger.WithError(err).Error("failed to get chat")
			return
		}
		var chatID int64
		if executorChat != nil {
			chatID = executorChat.ID
		}
		if err = b.storage.SetTaskInProgressUser(ctx, message.Chat.ID, userContact, chatID); err != nil {
			logger.WithError(err).Error("failed to set task in progress user for the chat")
			responseMsg.Text = errorReponse
			return
		}
		responseMsg.Text = "Введите дедлайн задачи в формате 21.12.2024 12:20:00"

	case domain.AddTaskDeadline:
		nextStage = domain.Default
		taskInProgress, err := b.storage.GetTaskInProgress(ctx, message.Chat.ID)
		if err != nil && !errors.Is(err, errs.ErrNotFound) {
			logger.WithError(err).Error("failed to get task in progress for the chat")
			responseMsg.Text = errorReponse
			return
		}
		if b.handleAddTaskDeadline(ctx, logger, message, &taskInProgress, &responseMsg) {
			return
		}
	}

	if err := b.storage.SetStage(ctx, message.Chat.ID, nextStage); err != nil && !errors.Is(err, errs.ErrNotFound) {
		logger.WithError(err).Error("failed to set next stage")
	}
}

// handleAddTaskDeadline parses deadline timestamp, creates task, notify observers and executor.
// returns bool if external call should return after
func (b *Bot) handleAddTaskDeadline(
	ctx context.Context,
	logger *log.Entry,
	message *tgbotapi.Message,
	taskInProgress *domain.Task,
	responseMsg *tgbotapi.MessageConfig,
) bool {
	timestamp, err := time.ParseInLocation(domain.DeadlineLayout, message.Text, time.Local)
	if err != nil {
		responseMsg.Text = "Некорректный формат даты-времени, проверьте, что вы вводите дату и время в формате, похожем на 21.12.2024 12:20:00"
		return true
	}
	if timestamp.Before(time.Now()) {
		responseMsg.Text = "Некорректное время дедлайна. Убедитесь, что вы ввели время момента в будущем в качестве дедлайна"
		return true
	}
	taskInProgress.Deadline = timestamp

	taskID, err := b.storage.AddTask(ctx, *taskInProgress)
	if err != nil {
		logger.WithError(err).Error("failed to add task")
		responseMsg.Text = errorReponse
		return true
	}
	taskInProgress.ID = taskID

	responseMsg.ParseMode = tgbotapi.ModeHTML
	responseMsg.Text = fmt.Sprintf("Вы успешно добавили задачу: \n\n%s", taskInProgress)

	if err := b.NotifyObservers(ctx, *taskInProgress, message.Chat.ID); err != nil {
		logger.WithError(err).Error("failed to notify observers")
	}
	// if executor's chat id set - send a notification about created task
	if taskInProgress.ExecutorChatID != 0 {
		if err := b.NotifyExecutor(ctx, *taskInProgress, message.Chat.ID); err != nil {
			logger.WithError(err).Error("failed to notify executor")
		}
	}
	return false
}

func (b *Bot) handleMarkTaskStage(ctx context.Context, message *tgbotapi.Message, stage domain.Stage) {
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

	if err := b.storage.SetStage(ctx, message.Chat.ID, domain.Default); err != nil && !errors.Is(err, errs.ErrNotFound) {
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

	if err := b.storage.SetStage(ctx, message.Chat.ID, domain.Default); err != nil && !errors.Is(err, errs.ErrNotFound) {
		logger.WithError(err).Error("failed to set next stage")
		responseMsg.Text = errorReponse
		return
	}

	responseMsg.ParseMode = tgbotapi.ModeHTML
	responseMsg.Text = fmt.Sprintf("Дедлайн задачи №%d успешно изменен на %s", taskID, deadline)
}

func (b *Bot) handleDeleteTaskStage(ctx context.Context, message *tgbotapi.Message) {
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

	if err := b.storage.DeleteTask(ctx, taskID); err != nil {
		logger.WithError(err).Error("b.storage.MarkTaskAsClosed")
		responseMsg.Text = errorReponse
		return
	}
	responseMsg.Text = "Задача успешно удалена"

	if err := b.storage.SetStage(ctx, message.Chat.ID, domain.Default); err != nil && !errors.Is(err, errs.ErrNotFound) {
		logger.WithError(err).Error("failed to set next stage")
		responseMsg.Text = errorReponse
		return
	}
}
