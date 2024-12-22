package repository

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"tasks_bot/internal/domain"
	"tasks_bot/internal/errs"
	"time"
)

const (
	queueSize = 1000
)

type MemoryStorage struct {
	mu *sync.RWMutex

	chats           map[int64]*domain.Chat
	tasks           []domain.Task
	tasksInProgress map[int64]domain.Task
	messageQueue    []domain.Message

	closed atomic.Bool
}

func NewMemoryStorage(_ context.Context) (*MemoryStorage, error) {
	return &MemoryStorage{
		mu:              &sync.RWMutex{},
		chats:           make(map[int64]*domain.Chat),
		tasks:           make([]domain.Task, 0, queueSize),
		tasksInProgress: make(map[int64]domain.Task, queueSize),
		messageQueue:    make([]domain.Message, 0, queueSize),
		closed:          atomic.Bool{},
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

func (ms *MemoryStorage) GetObservers(ctx context.Context) (map[int64]*domain.Chat, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	observers := make(map[int64]*domain.Chat, 1)
	for chatID, chat := range ms.chats {
		if chat.Role == domain.Observer {
			observers[chatID] = chat
		}
	}

	return observers, nil
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
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.messageQueue[messageID].IsHandled = true
	return nil
}

func (ms *MemoryStorage) GetAllTasks(ctx context.Context) ([]domain.Task, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	tasks := make([]domain.Task, len(ms.tasks))
	copy(tasks, ms.tasks)

	return tasks, nil
}

func (ms *MemoryStorage) GetExpiredTasks(ctx context.Context) ([]domain.Task, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	tasks := make([]domain.Task, 0)
	for _, task := range ms.tasks {
		if !task.Done && task.Expired {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (ms *MemoryStorage) GetExpiredTasksToMark(ctx context.Context) ([]domain.Task, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	tasks := make([]domain.Task, 0)
	for i, task := range ms.tasks {
		if task.Done || task.Expired || time.Now().Before(task.Deadline) {
			continue
		}
		ms.tasks[i].Expired = true
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (ms *MemoryStorage) GetOpenTasks(ctx context.Context) ([]domain.Task, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	tasks := make([]domain.Task, 0, len(ms.tasks))
	for _, task := range ms.tasks {
		if task.Closed {
			continue
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (ms *MemoryStorage) GetDoneTasks(ctx context.Context) ([]domain.Task, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	tasks := make([]domain.Task, 0, len(ms.tasks))
	for _, task := range ms.tasks {
		if task.Done {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (ms *MemoryStorage) GetClosedTasks(ctx context.Context) ([]domain.Task, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	tasks := make([]domain.Task, 0, len(ms.tasks))
	for _, task := range ms.tasks {
		if task.Closed {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (ms *MemoryStorage) GetUserTasks(ctx context.Context, username string) ([]domain.Task, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	tasks := make([]domain.Task, 0, len(ms.tasks))
	for _, task := range ms.tasks {
		if task.Executor == username {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

func (ms *MemoryStorage) AddTask(ctx context.Context, task domain.Task) (int, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.tasks = append(ms.tasks, task)

	return len(ms.tasks) + 1, nil
}

func (ms *MemoryStorage) MarkTaskAsDone(ctx context.Context, taskID int) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for i, task := range ms.tasks {
		if task.ID == taskID {
			ms.tasks[i].Done = true
			return nil
		}
	}
	return errs.ErrNotFound
}

func (ms *MemoryStorage) MarkTaskAsClosed(ctx context.Context, taskID int) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for i, task := range ms.tasks {
		if task.ID == taskID {
			ms.tasks[i].Closed = true
			return nil
		}
	}
	return errs.ErrNotFound
}

func (ms *MemoryStorage) ChangeTaskDeadline(ctx context.Context, taskID int, newDeadline time.Time) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for i, task := range ms.tasks {
		if task.ID == taskID {
			ms.tasks[i].Deadline = newDeadline
			ms.tasks[i].Expired = false
			return nil
		}
	}
	return errs.ErrNotFound
}

func (ms *MemoryStorage) GetTaskInProgress(ctx context.Context, chatID int64) (domain.Task, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return ms.tasksInProgress[chatID], nil
}

func (ms *MemoryStorage) SetTaskInProgressName(ctx context.Context, chatID int64, name string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	task := ms.tasksInProgress[chatID]
	task.ID = len(ms.tasks) + 1
	task.Title = name
	ms.tasksInProgress[chatID] = task

	return nil
}

func (ms *MemoryStorage) SetTaskInProgressUser(ctx context.Context, chatID int64, user string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	task := ms.tasksInProgress[chatID]
	task.Executor = user
	task.ID = len(ms.tasks) + 1
	ms.tasksInProgress[chatID] = task

	return nil
}

func (ms *MemoryStorage) SetTaskInProgressDeadline(ctx context.Context, chatID int64, deadline time.Time) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	task := ms.tasksInProgress[chatID]
	task.Deadline = deadline
	task.ID = len(ms.tasks) + 1
	ms.tasksInProgress[chatID] = task

	return nil
}

func (ms *MemoryStorage) DebugStorage(ctx context.Context) (string, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	builder := strings.Builder{}
	for id, chat := range ms.chats {
		builder.WriteString(fmt.Sprintf("chat with id: %d, chat: %+v\n", id, chat))
	}

	builder.WriteString("\n")
	for _, message := range ms.messageQueue {
		builder.WriteString(fmt.Sprintf("message from queue: %+v\n", message))
	}

	builder.WriteString("\n")
	for _, task := range ms.tasks {
		builder.WriteString(fmt.Sprintf("task: %s\n", task))
	}

	builder.WriteString("\n")
	for chatID, task := range ms.tasksInProgress {
		builder.WriteString(fmt.Sprintf("task in progress for chat %d: %s\n", chatID, task))
	}

	return builder.String(), nil
}

func (ms *MemoryStorage) Close() {
}
