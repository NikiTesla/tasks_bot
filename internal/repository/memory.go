package repository

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"tasks_bot/internal/domain"
	"tasks_bot/internal/errs"

	log "github.com/sirupsen/logrus"
)

const (
	queueSize = 1000
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
		messageQueue: make([]domain.Message, 0, queueSize),
		closed:       atomic.Bool{},
	}, nil
}

func (ms *MemoryStorage) GetRole(ctx context.Context, chatID int64) (domain.Role, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	chat, ok := ms.chats[chatID]
	if !ok {
		return domain.UnknownRole, errs.ErrNotFound
	}

	return chat.Role, nil
}

func (ms *MemoryStorage) SetRole(ctx context.Context, chatID int64, role domain.Role) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	chat, ok := ms.chats[chatID]
	if !ok {
		return errs.ErrNotFound
	}
	chat.Role = role

	return nil
}

func (ms *MemoryStorage) AddChat(ctx context.Context, chatID int64, username string, role domain.Role) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.chats[chatID] = &domain.Chat{
		Username: username,
		Stage:    domain.Default,
		Role:     role,
	}

	return nil
}

func (ms *MemoryStorage) SetStage(ctx context.Context, chatID int64, stage domain.Stage) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	chat, ok := ms.chats[chatID]
	if !ok {
		return errs.ErrNotFound
	}
	chat.Stage = stage

	return nil
}
func (ms *MemoryStorage) GetStage(ctx context.Context, chatID int64) (domain.Stage, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	chat, ok := ms.chats[chatID]
	if !ok {
		return domain.Unknown, errs.ErrNotFound
	}

	return chat.Stage, nil
}

func (ms *MemoryStorage) AddMessage(ctx context.Context, message domain.Message) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	message.ID = len(ms.messageQueue)
	ms.messageQueue = append(ms.messageQueue, message)

	return nil
}

func (ms *MemoryStorage) RetrieveMessages(ctx context.Context) ([]domain.Message, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	messages := make([]domain.Message, 0, len(ms.messageQueue))
	for _, message := range ms.messageQueue {
		if message.IsHandled {
			continue
		}
		messages = append(messages, message)
	}

	return messages, nil
}

func (ms *MemoryStorage) SetHandledMessage(ctx context.Context, messageID int) error {
	ms.messageQueue[messageID].IsHandled = true
	return nil
}

func (ms *MemoryStorage) DebugStorage(ctx context.Context) string {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	builder := strings.Builder{}
	log.Debug("debuggin storage")
	for id, chat := range ms.chats {
		builder.WriteString(fmt.Sprintf("chat with id: %d, chat: %+v\n", id, chat))
	}

	for _, message := range ms.messageQueue {
		builder.WriteString(fmt.Sprintf("message from queue: %+v\n", message))
	}

	return builder.String()
}

func (ms *MemoryStorage) Close() {
}
