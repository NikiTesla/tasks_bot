package repository

import (
	"context"
	"tasks_bot/internal/domain"
	"time"
)

type Storage interface {
	DebugStorage(ctx context.Context) string

	// chats
	AddChat(ctx context.Context, chatID int64, username string, role domain.Role) error

	// role
	GetRole(ctx context.Context, chatID int64) (domain.Role, error)
	SetRole(ctx context.Context, chatID int64, role domain.Role) error

	// stage
	SetStage(ctx context.Context, chatID int64, stage domain.Stage) error
	GetStage(ctx context.Context, chatID int64) (domain.Stage, error)

	GetObservers(ctx context.Context) (map[int64]*domain.Chat, error)

	// tasks
	AddTask(ctx context.Context, task domain.Task) error
	GetAllTasks(ctx context.Context) ([]domain.Task, error)
	GetClosedTasks(ctx context.Context) ([]domain.Task, error)
	GetOpenTasks(ctx context.Context) ([]domain.Task, error)
	GetDoneTasks(ctx context.Context) ([]domain.Task, error)
	GetExpiredTasks(ctx context.Context) ([]domain.Task, error)
	MarkTaskAsDone(ctx context.Context, taskID int) error
	MarkTaskAsClosed(ctx context.Context, taskID int) error
	ChangeTaskDeadline(ctx context.Context, taskID int, newDeadline time.Time) error

	GetTaskInProgress(ctx context.Context, chatID int64) (domain.Task, error)
	SetTaskInProgressName(ctx context.Context, chatID int64, name string) error
	SetTaskInProgressUser(ctx context.Context, chatID int64, user string) error
	SetTaskInProgressDeadline(ctx context.Context, chatID int64, deadline time.Time) error

	// messages
	AddMessage(ctx context.Context, message domain.Message) error
	RetrieveMessages(ctx context.Context) ([]domain.Message, error)
	SetHandledMessage(ctx context.Context, messageID int) error

	Close()
}
