package telegram

import (
	"context"
	"errors"
	"fmt"
	"tasks_bot/internal/domain"
	"tasks_bot/internal/errs"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

const (
	enterTaskNumberText = "Введите номер задачи"
)

func (b *Bot) handleCommand(ctx context.Context, message *tgbotapi.Message) {
	logger := b.logger.WithField("chatId", message.Chat.ID).WithField("command", message.Command())
	role, err := b.storage.GetRole(ctx, message.Chat.ID)
	if err != nil {
		if !errors.Is(err, errs.ErrNotFound) {
			logger.WithError(err).Error("failed to get role")
			return
		}
	}

	switch role {
	case domain.Admin:
		b.processAdminCommands(ctx, message, logger)
	case domain.Chief:
		b.processChiefCommands(ctx, message, logger)
	case domain.Executor:
		b.processExecutorCommands(ctx, message, logger)
	case domain.Observer:
		b.processObserverCommands(ctx, message, logger)
	default:
		b.processDefaulCommands(ctx, message, logger)
	}
}

func (b *Bot) setCommands(_ context.Context, chatID int64, role domain.Role) error {
	commands, ok := role2commands[role]
	if !ok {
		return fmt.Errorf("no commands found for role %s", role)
	}

	commandCfg := tgbotapi.NewSetMyCommandsWithScopeAndLanguage(
		tgbotapi.BotCommandScope{
			Type:   "chat",
			ChatID: chatID,
		},
		"",
		commands...,
	)

	if _, err := b.bot.Request(commandCfg); err != nil {
		return fmt.Errorf("b.bot.Request: %w", err)
	}
	return nil
}

func (b *Bot) processDefaulCommands(ctx context.Context, message *tgbotapi.Message, logger *log.Entry) {
	switch message.Command() {
	case startCmd:
		b.handleStart(ctx, message)

	case becomeExecutorCmd, becomeAdminCmd, becomeChiefCmd, becomeObserverCmd:
		b.handleBecomeCommand(ctx, message, message.Command())

	case getRoleCmd:
		b.handleGetRoleCommand(ctx, message)

	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неизвестная или недоступная команда, попробуйте другую")
		if _, err := b.bot.Send(msg); err != nil {
			logger.WithError(err).Error("unable to send response")
		}
	}
}

func (b *Bot) processExecutorCommands(ctx context.Context, message *tgbotapi.Message, logger *log.Entry) {
	switch message.Command() {
	case startCmd:
		b.handleStart(ctx, message)
	case becomeAdminCmd, becomeChiefCmd, becomeObserverCmd:
		b.handleBecomeCommand(ctx, message, message.Command())
	case getRoleCmd:
		b.handleGetRoleCommand(ctx, message)
	case markTaskAsDoneCommand:
		b.setNextStageWithMessage(ctx, message, domain.MarkTaskAsDone, enterTaskNumberText)
	case getSelfTasksCmd:
		b.handleGetSelfTasksCommand(ctx, message)
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неизвестная или недоступная команда, попробуйте другую")
		if _, err := b.bot.Send(msg); err != nil {
			logger.WithError(err).Error("unable to send response")
		}
	}
}

func (b *Bot) processChiefCommands(ctx context.Context, message *tgbotapi.Message, logger *log.Entry) {
	switch message.Command() {
	case startCmd:
		b.handleStart(ctx, message)
	case becomeExecutorCmd, becomeAdminCmd, becomeObserverCmd:
		b.handleBecomeCommand(ctx, message, message.Command())
	case getRoleCmd:
		b.handleGetRoleCommand(ctx, message)
	case getAllTasksCmd:
		b.handleGetAllTasksCommand(ctx, message)
	case getExpiredTasksCmd:
		b.handleGetExpiredTasksCommand(ctx, message)
	case addTaskCmd:
		b.setNextStageWithMessage(ctx, message, domain.AddTaskName, "Введите название задачи")
	case getOpenTasks:
		b.handleGetOpenTasksCommand(ctx, message)
	case getDoneTasks:
		b.handleGetDoneTasksCommand(ctx, message)
	case markTaskAsDoneCommand:
		b.setNextStageWithMessage(ctx, message, domain.MarkTaskAsDone, enterTaskNumberText)
	case changeTaskDeadlineCommand:
		b.setNextStageWithMessage(ctx, message, domain.ChangeDeadline, "Введите номер задачи и новый дедлайн в формате \"21 21.12.2024 12:20:00\"")

	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неизвестная или недоступная команда, попробуйте другую")
		if _, err := b.bot.Send(msg); err != nil {
			logger.WithError(err).Error("unable to send response")
		}
	}
}

func (b *Bot) processObserverCommands(ctx context.Context, message *tgbotapi.Message, logger *log.Entry) {
	switch message.Command() {
	case startCmd:
		b.handleStart(ctx, message)
	case becomeExecutorCmd, becomeAdminCmd, becomeChiefCmd:
		b.handleBecomeCommand(ctx, message, message.Command())
	case getRoleCmd:
		b.handleGetRoleCommand(ctx, message)
	case getAllTasksCmd:
		b.handleGetAllTasksCommand(ctx, message)
	case getExpiredTasksCmd:
		b.handleGetExpiredTasksCommand(ctx, message)
	case addTaskCmd:
		b.setNextStageWithMessage(ctx, message, domain.AddTaskName, "Введите название задачи")
	case getOpenTasks:
		b.handleGetOpenTasksCommand(ctx, message)
	case getDoneTasks:
		b.handleGetDoneTasksCommand(ctx, message)
	case getClosedTasks:
		b.handleGetClosedTasksCommand(ctx, message)
	case markTaskAsClosedCommand:
		b.setNextStageWithMessage(ctx, message, domain.MarkTaskAsClosed, enterTaskNumberText)
	case markTaskAsDoneCommand:
		b.setNextStageWithMessage(ctx, message, domain.MarkTaskAsDone, enterTaskNumberText)
	case deleteTaskCommand:
		b.setNextStageWithMessage(ctx, message, domain.DeleteTask, enterTaskNumberText)
	case changeTaskDeadlineCommand:
		b.setNextStageWithMessage(ctx, message, domain.ChangeDeadline, "Введите номер задачи и новый дедлайн в формате \"21 21.12.2024 12:20:00\"")
	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неизвестная или недоступная команда, попробуйте другую")
		if _, err := b.bot.Send(msg); err != nil {
			logger.WithError(err).Error("unable to send response")
		}
	}
}

func (b *Bot) processAdminCommands(ctx context.Context, message *tgbotapi.Message, logger *log.Entry) {
	switch message.Command() {
	case startCmd:
		b.handleStart(ctx, message)
	case becomeExecutorCmd, becomeChiefCmd, becomeObserverCmd:
		b.handleBecomeCommand(ctx, message, message.Command())
	case getRoleCmd:
		b.handleGetRoleCommand(ctx, message)
	case getAllTasksCmd:
		b.handleGetAllTasksCommand(ctx, message)
	case getExpiredTasksCmd:
		b.handleGetExpiredTasksCommand(ctx, message)
	case addTaskCmd:
		b.setNextStageWithMessage(ctx, message, domain.AddTaskName, "Введите название задачи")
	case getOpenTasks:
		b.handleGetOpenTasksCommand(ctx, message)
	case getDoneTasks:
		b.handleGetDoneTasksCommand(ctx, message)
	case getClosedTasks:
		b.handleGetClosedTasksCommand(ctx, message)
	case markTaskAsClosedCommand:
		b.setNextStageWithMessage(ctx, message, domain.MarkTaskAsClosed, "Введите номер задачи")
	case markTaskAsDoneCommand:
		b.setNextStageWithMessage(ctx, message, domain.MarkTaskAsDone, "Введите номер задачи")
	case deleteTaskCommand:
		b.setNextStageWithMessage(ctx, message, domain.DeleteTask, "Введите номер задачи")
	case changeTaskDeadlineCommand:
		b.setNextStageWithMessage(ctx, message, domain.ChangeDeadline, "Введите номер задачи и новый дедлайн в формате \"21 21.12.2024 12:20:00\"")

	case healthCmd:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Status Ok!")
		if _, err := b.bot.Send(msg); err != nil {
			logger.WithError(err).Error("unable to send response")
		}

	case debugStorage:
		msg := tgbotapi.NewMessage(message.Chat.ID, b.debugCommand(message))
		if _, err := b.bot.Send(msg); err != nil {
			logger.WithError(err).Error("unable to send response")
		}

	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неизвестная команда, попробуйте другую")
		if _, err := b.bot.Send(msg); err != nil {
			logger.WithError(err).Error("unable to send response")
		}
	}
}
