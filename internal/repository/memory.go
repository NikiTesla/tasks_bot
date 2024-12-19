package repository

import (
	"context"
	"sync"
	"sync/atomic"
	"tasks_bot/internal/domain"
	"tasks_bot/internal/errs"
)

const (
	queueSize = 10
)

type MemoryStorage struct {
	mu *sync.RWMutex

	chats        map[int64]*domain.Chat
	messageQueue []domain.Message

	closed atomic.Bool
}

func NewMemoryStorage(_ context.Context) (*MemoryStorage, error) {
	return &MemoryStorage{
		mu:           &sync.RWMutex{},
		chats:        make(map[int64]*domain.Chat),
		messageQueue: make([]domain.Message, 0, 100),
		closed:       atomic.Bool{},
	}, nil
}

func (ms *MemoryStorage) AddChat(ctx context.Context, chatID int64, username string, role domain.Role) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.chats[chatID] = &domain.Chat{
		Username: username,
		Role:     role,
	}

	return nil
}

func (ms *MemoryStorage) IsAdmin(ctx context.Context, chatID int64) (bool, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	chat, ok := ms.chats[chatID]
	if !ok {
		return false, errs.ErrNotFound
	}

	return chat.Role == domain.Admin, nil
}

func (ms *MemoryStorage) AddMessage(ctx context.Context, message domain.Message) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.messageQueue = append(ms.messageQueue, message)

	return nil
}

func (ms *MemoryStorage) RetrieveMessages(ctx context.Context) ([]domain.Message, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	messages := make([]domain.Message, len(ms.messageQueue))
	copy(messages, ms.messageQueue)
	ms.messageQueue = ms.messageQueue[:0]

	return messages, nil
}

func (ms *MemoryStorage) Close() {
}
