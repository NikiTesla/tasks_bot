package repository

import (
	"context"
	"tasks_bot/internal/domain"
)

type Storage interface {
	DebugStorage(ctx context.Context) string

	AddChat(ctx context.Context, chatID int64, username string, role domain.Role) error
	GetRole(ctx context.Context, chatID int64) (domain.Role, error)
	SetRole(ctx context.Context, chatID int64, role domain.Role) error
	SetStage(ctx context.Context, chatID int64, stage domain.Stage) error
	GetStage(ctx context.Context, chatID int64) (domain.Stage, error)

	AddMessage(ctx context.Context, message domain.Message) error
	RetrieveMessages(ctx context.Context) ([]domain.Message, error)
	SetHandledMessage(ctx context.Context, messageID int) error
	Close()
}
