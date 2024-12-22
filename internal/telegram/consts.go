package telegram

import (
	"tasks_bot/internal/domain"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	getSelfTasksCmd           = "get_self_tasks"
	getExpiredTasksCmd        = "get_expired_tasks"
	markTaskAsDoneCommand     = "do_task"
	markTaskAsClosedCommand   = "close_task"
	changeTaskDeadlineCommand = "change_deadline"
	// admin commands
	healthCmd    = "healthz"
	debugStorage = "debug"
)

var role2commands = map[domain.Role][]tgbotapi.BotCommand{
	domain.UnknownRole: {
		{Command: startCmd, Description: "Начать"},
		{Command: getRoleCmd, Description: "Получить свою роль"},
		{Command: becomeExecutorCmd, Description: "Стать исполнителем"},
		{Command: becomeChiefCmd, Description: "Стать шефом"},
		{Command: becomeObserverCmd, Description: "Стать наблюдателем"},
		{Command: becomeAdminCmd, Description: "Стать администратором"},
	},
	domain.Executor: {
		{Command: startCmd, Description: "Начать"},
		{Command: getRoleCmd, Description: "Получить свою роль"},
		{Command: getSelfTasksCmd, Description: "Получить свои задачи"},
		{Command: becomeExecutorCmd, Description: "Стать исполнителем"},
		{Command: becomeChiefCmd, Description: "Стать шефом"},
		{Command: becomeObserverCmd, Description: "Стать наблюдателем"},
		{Command: becomeAdminCmd, Description: "Стать администратором"},
	},
	domain.Chief: {
		{Command: startCmd, Description: "Начать"},
		{Command: getRoleCmd, Description: "Получить свою роль"},
		{Command: addTaskCmd, Description: "Добавить задачу"},
		{Command: getAllTasksCmd, Description: "Получить все задачи"},
		{Command: getExpiredTasksCmd, Description: "Получить просроченные задачи"},
		{Command: getOpenTasks, Description: "Получить открытые задачи"},
		{Command: getDoneTasks, Description: "Получить выполненные задачи"},
		{Command: markTaskAsDoneCommand, Description: "Отметить задачу выполненной"},
		{Command: changeTaskDeadlineCommand, Description: "Изменить дедлайн задачи"},
		{Command: becomeExecutorCmd, Description: "Стать исполнителем"},
		{Command: becomeChiefCmd, Description: "Стать шефом"},
		{Command: becomeObserverCmd, Description: "Стать наблюдателем"},
		{Command: becomeAdminCmd, Description: "Стать администратором"},
	},
	domain.Observer: {
		{Command: startCmd, Description: "Начать"},
		{Command: getRoleCmd, Description: "Получить свою роль"},
		{Command: addTaskCmd, Description: "Добавить задачу"},
		{Command: getAllTasksCmd, Description: "Получить все задачи"},
		{Command: getExpiredTasksCmd, Description: "Получить просроченные задачи"},
		{Command: getOpenTasks, Description: "Получить открытые задачи"},
		{Command: getDoneTasks, Description: "Получить выполненные задачи"},
		{Command: getClosedTasks, Description: "Получить закрытые задачи"},
		{Command: markTaskAsClosedCommand, Description: "Закрыть задачу"},
		{Command: markTaskAsDoneCommand, Description: "Отметить задачу выполненной"},
		{Command: changeTaskDeadlineCommand, Description: "Изменить дедлайн задачи"},
		{Command: becomeExecutorCmd, Description: "Стать исполнителем"},
		{Command: becomeChiefCmd, Description: "Стать шефом"},
		{Command: becomeObserverCmd, Description: "Стать наблюдателем"},
		{Command: becomeAdminCmd, Description: "Стать администратором"},
	},
	domain.Admin: {
		{Command: startCmd, Description: "Начать"},
		{Command: getRoleCmd, Description: "Получить свою роль"},
		{Command: addTaskCmd, Description: "Добавить задачу"},
		{Command: getAllTasksCmd, Description: "Получить все задачи"},
		{Command: getExpiredTasksCmd, Description: "Получить просроченные задачи"},
		{Command: getOpenTasks, Description: "Получить открытые задачи"},
		{Command: getDoneTasks, Description: "Получить выполненные задачи"},
		{Command: getClosedTasks, Description: "Получить закрытые задачи"},
		{Command: markTaskAsClosedCommand, Description: "Закрыть задачу"},
		{Command: markTaskAsDoneCommand, Description: "Отметить задачу выполненной"},
		{Command: changeTaskDeadlineCommand, Description: "Изменить дедлайн задачи"},
		{Command: becomeExecutorCmd, Description: "Стать исполнителем"},
		{Command: becomeChiefCmd, Description: "Стать шефом"},
		{Command: becomeObserverCmd, Description: "Стать наблюдателем"},
		{Command: becomeAdminCmd, Description: "Стать администратором"},
		{Command: healthCmd, Description: "Проверить состояние"},
		{Command: debugStorage, Description: "Отладка хранилища"},
	},
}
