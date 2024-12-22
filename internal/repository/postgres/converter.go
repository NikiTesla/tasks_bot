package postgres

import (
	"tasks_bot/internal/domain"
	queries "tasks_bot/internal/repository/postgres/sqlc"
)

func TaskToDomain(task *queries.Task) domain.Task {
	return domain.Task{
		ID:       int(task.ID) + 1,
		Title:    task.Title,
		Executor: task.Executor,
		Deadline: task.Deadline.Time,
		Done:     task.Done,
		Expired:  task.Expired,
		Closed:   task.Closed,
	}
}

func ChatToDomain(chat *queries.Chat) *domain.Chat {
	return &domain.Chat{
		ID:       chat.ChatID,
		Username: chat.Username,
		Stage:    domain.Stage(chat.Stage.Int32),
		Role:     domain.Role(chat.Role.Int32),
	}
}
