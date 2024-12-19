package repository

import (
	"context"
	"tasks_bot/internal/domain"
)

type Storage interface {
	AddChat(ctx context.Context, chatID int64, username string, role domain.Role) error
	IsAdmin(ctx context.Context, chatID int64) (bool, error)

	AddMessage(ctx context.Context, message domain.Message) error
	RetrieveMessages(ctx context.Context) ([]domain.Message, error)
	Close()
}
