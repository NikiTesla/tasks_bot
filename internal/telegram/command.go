package telegram

import (
	"context"
	"fmt"
	"strings"
	"tasks_bot/internal/domain"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	log "github.com/sirupsen/logrus"
)

const (
	startCmd                  = "start"
	becomeExecutorCmd         = "become_executor"
	becomeObserverCmd         = "become_observer"
	becomeChiefCmd            = "become_chief"
	becomeAdminCmd            = "become_admin"
	getRoleCmd                = "get_role"
	addTaskCmd                = "add_task"
	getAllTasksCmd            = "get_all_tasks"
	getOpenTasks              = "get_open_tasks"
	getDoneTasks              = "get_done_tasks"
	getClosedTasks            = "get_closed_tasks"
	getExpiredTasksCmd        = "get_expired_tasks"
	markTaskAsDoneCommand     = "do_task"
	markTaskAsClosedCommand   = "close_task"
	changeTaskDeadlineCommand = "change_deadline"
	// admin commands
	healthCmd    = "healthz"
	debugStorage = "debug"
)

// TODO add context to handlers
func (b *Bot) handleCommand(ctx context.Context, message *tgbotapi.Message) {
	logger := b.logger.WithField("chatId", message.Chat.ID).WithField("command", message.Command())

	role, err := b.storage.GetRole(ctx, message.Chat.ID)
	if err != nil {
		logger.WithError(err).Error("failed to get role")
		return
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

func (b *Bot) processAdminCommands(ctx context.Context, message *tgbotapi.Message, logger *log.Entry) {
	switch message.Command() {
	case startCmd:
		b.handleStart(message)

	case becomeExecutorCmd, becomeAdminCmd, becomeChiefCmd, becomeObserverCmd:
		b.handleBecomeCommand(message, message.Command())

	case getRoleCmd:
		b.handleGetRoleCommand(message)

	case getAllTasksCmd:
		b.handleGetAllTasksCommand(ctx, message)

	case getExpiredTasksCmd:
		b.handleGetExpiredTasksCommand(ctx, message)

	case addTaskCmd:
		b.handleAddTask(ctx, message)

	case getOpenTasks:
		b.handleGetOpenTasksCommand(ctx, message)

	case getDoneTasks:
		b.handleGetDoneTasksCommand(ctx, message)

	case getClosedTasks:
		b.handleGetClosedTasksCommand(ctx, message)

	case markTaskAsClosedCommand:
		b.handleMarkTaskCommands(ctx, message, domain.MarkTaskAsClosed)

	case markTaskAsDoneCommand:
		b.handleMarkTaskCommands(ctx, message, domain.MarkTaskAsDone)

	case changeTaskDeadlineCommand:
		b.handleChangeTaskDeadlineCommand(ctx, message)

	case healthCmd:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Status Ok!")
		if _, err := b.bot.Send(msg); err != nil {
			logger.WithError(err).Error("unable to send response")
		}

	case debugStorage:
		msg := tgbotapi.NewMessage(message.Chat.ID, b.debug(message))
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

func (b *Bot) processExecutorCommands(ctx context.Context, message *tgbotapi.Message, logger *log.Entry) {
	switch message.Command() {
	case startCmd:
		b.handleStart(message)

	case becomeExecutorCmd, becomeAdminCmd, becomeChiefCmd, becomeObserverCmd:
		b.handleBecomeCommand(message, message.Command())

	case getRoleCmd:
		b.handleGetRoleCommand(message)

	// TODO think about adding command to get self tasks

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
		b.handleStart(message)

	case becomeExecutorCmd, becomeAdminCmd, becomeChiefCmd, becomeObserverCmd:
		b.handleBecomeCommand(message, message.Command())

	case getRoleCmd:
		b.handleGetRoleCommand(message)

	case getAllTasksCmd:
		b.handleGetAllTasksCommand(ctx, message)

	case getExpiredTasksCmd:
		b.handleGetExpiredTasksCommand(ctx, message)

	case addTaskCmd:
		b.handleAddTask(ctx, message)

	case getOpenTasks:
		b.handleGetOpenTasksCommand(ctx, message)

	case getDoneTasks:
		b.handleGetDoneTasksCommand(ctx, message)

	case getClosedTasks:
		b.handleGetClosedTasksCommand(ctx, message)

	case markTaskAsClosedCommand:
		b.handleMarkTaskCommands(ctx, message, domain.MarkTaskAsClosed)

	case markTaskAsDoneCommand:
		b.handleMarkTaskCommands(ctx, message, domain.MarkTaskAsDone)

	case changeTaskDeadlineCommand:
		b.handleChangeTaskDeadlineCommand(ctx, message)

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
		b.handleStart(message)

	case becomeExecutorCmd, becomeAdminCmd, becomeChiefCmd, becomeObserverCmd:
		b.handleBecomeCommand(message, message.Command())

	case getRoleCmd:
		b.handleGetRoleCommand(message)

	case getAllTasksCmd:
		b.handleGetAllTasksCommand(ctx, message)

	case getExpiredTasksCmd:
		b.handleGetExpiredTasksCommand(ctx, message)

	case addTaskCmd:
		b.handleAddTask(ctx, message)

	case getOpenTasks:
		b.handleGetOpenTasksCommand(ctx, message)

	case getDoneTasks:
		b.handleGetDoneTasksCommand(ctx, message)

	case markTaskAsDoneCommand:
		b.handleMarkTaskCommands(ctx, message, domain.MarkTaskAsDone)

	case changeTaskDeadlineCommand:
		b.handleChangeTaskDeadlineCommand(ctx, message)

	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неизвестная или недоступная команда, попробуйте другую")
		if _, err := b.bot.Send(msg); err != nil {
			logger.WithError(err).Error("unable to send response")
		}
	}
}

func (b *Bot) processDefaulCommands(ctx context.Context, message *tgbotapi.Message, logger *log.Entry) {
	switch message.Command() {
	case startCmd:
		b.handleStart(message)

	case becomeExecutorCmd, becomeAdminCmd, becomeChiefCmd, becomeObserverCmd:
		b.handleBecomeCommand(message, message.Command())

	case getRoleCmd:
		b.handleGetRoleCommand(message)

	default:
		msg := tgbotapi.NewMessage(message.Chat.ID, "Неизвестная или недоступная команда, попробуйте другую")
		if _, err := b.bot.Send(msg); err != nil {
			logger.WithError(err).Error("unable to send response")
		}
	}
}

func (b *Bot) handleStart(message *tgbotapi.Message) {
	_ = b.storage.SetStage(context.Background(), message.Chat.ID, domain.Default)

	msg := tgbotapi.NewMessage(message.Chat.ID, startMessage)
	if _, err := b.bot.Send(msg); err != nil {
		b.logger.WithError(err).WithField("chatID", message.Chat.ID).Error("failed to sent start message")
	}
}

func (b *Bot) debug(message *tgbotapi.Message) string {
	storageDump := b.storage.DebugStorage(b.logger.Context)
	var response string
	response += fmt.Sprintf("Your chat id is %d\n", message.Chat.ID)
	response += storageDump

	return response
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
	if err != nil {
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
	if err != nil {
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
	if err != nil {
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
	if err != nil {
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
	if err != nil {
		logger.WithError(err).Error("failed to get expired tasks")
		responseMsg.Text = errorReponse
		return
	}

	// TODO add role check
	// role, err := b.storage.GetRole(context.Background(), message.Chat.ID)
	// if err != nil {
	// 	logger.WithError(err).Error("failed to get role")
	// 	responseMsg.Text = errorReponse
	// 	return
	// }

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

func (b *Bot) handleAddTask(ctx context.Context, message *tgbotapi.Message) {
	logger := b.logger.WithField("chatID", message.Chat.ID)

	responseMsg := tgbotapi.NewMessage(message.Chat.ID, "")
	defer func() {
		if _, err := b.bot.Send(responseMsg); err != nil {
			logger.WithError(err).Error("failed to send response")
		}
	}()

	if err := b.storage.SetStage(ctx, message.Chat.ID, domain.AddTaskName); err != nil {
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

	if err := b.storage.SetStage(ctx, message.Chat.ID, stage); err != nil {
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

	if err := b.storage.SetStage(ctx, message.Chat.ID, domain.ChangeDeadline); err != nil {
		logger.WithError(err).Error("b.storage.SetStage")
		responseMsg.Text = errorReponse
		return
	}
	responseMsg.Text = "Введите номер задачи и новый дедлайн в формате \"21 21.12.2024 12:20:00\""
}
