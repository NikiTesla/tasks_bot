package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"tasks_bot/internal/domain"
	"tasks_bot/internal/errs"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteStorage struct {
	db *sql.DB
}

func NewSQLiteStorage(ctx context.Context, dbFile string) (*SQLiteStorage, error) {
	db, err := connectDB(ctx, dbFile)
	if err != nil {
		return nil, fmt.Errorf("connecting db: %w", err)
	}
	return &SQLiteStorage{
		db: db,
	}, nil
}

func connectDB(ctx context.Context, dbFile string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}
	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db.Ping: %w", err)
	}
	_, err = db.ExecContext(ctx, `-- Schema for chats table
CREATE TABLE IF NOT EXISTS chats (
	chat_id INTEGER PRIMARY KEY,
	username TEXT,
	phone TEXT,
	role INTEGER DEFAULT 0,
	stage INTEGER DEFAULT 0,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Schema for tasks table
CREATE TABLE IF NOT EXISTS tasks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	executor_contact TEXT NOT NULL,
	executor_chat_id INTEGER,
	deadline TIMESTAMP NOT NULL,
	done BOOLEAN NOT NULL DEFAULT FALSE,
	closed BOOLEAN NOT NULL DEFAULT FALSE,
	expired BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Schema for tasks_in_progress table
CREATE TABLE IF NOT EXISTS tasks_in_progress (
	chat_id INTEGER PRIMARY KEY,
	title TEXT,
	executor_contact TEXT,
	executor_chat_id INTEGER,
	deadline TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`)
	if err != nil {
		return nil, fmt.Errorf("failed to exec migration")
	}
	return db, nil
}

func (s *SQLiteStorage) Close() {
	s.db.Close()
}

func (s *SQLiteStorage) DebugStorage(ctx context.Context) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (s *SQLiteStorage) AddChat(ctx context.Context, chatID int64, username, phone string, role domain.Role) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO chats (chat_id, username, phone, role) VALUES (?, ?, ?, ?)
		ON CONFLICT(chat_id) DO UPDATE SET 
			username = COALESCE(NULLIF(EXCLUDED.username, ''), chats.username), 
			phone = COALESCE(NULLIF(EXCLUDED.phone, ''), chats.phone)`,
		chatID, username, phone, int(role))
	if err != nil {
		return fmt.Errorf("sqlite.Exec: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) GetChat(ctx context.Context, username, phone string) (*domain.Chat, error) {
	row := s.db.QueryRowContext(ctx, `SELECT chat_id, username, phone, role, stage FROM chats WHERE username = ? OR phone = ?`, username, phone)
	var chat domain.Chat
	var role, stage int
	if err := row.Scan(&chat.ID, &chat.Username, &chat.Phone, &role, &stage); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("sqlite.QueryRow: %w", err)
	}
	chat.Role = domain.Role(role)
	chat.Stage = domain.Stage(stage)
	return &chat, nil
}

func (s *SQLiteStorage) GetRole(ctx context.Context, chatID int64) (domain.Role, error) {
	row := s.db.QueryRowContext(ctx, `SELECT role FROM chats WHERE chat_id = ?`, chatID)
	var role int
	if err := row.Scan(&role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.UnknownRole, errs.ErrNotFound
		}
		return domain.UnknownRole, fmt.Errorf("sqlite.QueryRow: %w", err)
	}
	return domain.Role(role), nil
}

func (s *SQLiteStorage) SetRole(ctx context.Context, chatID int64, role domain.Role) error {
	_, err := s.db.ExecContext(ctx, `UPDATE chats SET role = ? WHERE chat_id = ?`, int(role), chatID)
	if err != nil {
		return fmt.Errorf("sqlite.Exec: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) SetStage(ctx context.Context, chatID int64, stage domain.Stage) error {
	_, err := s.db.ExecContext(ctx, `UPDATE chats SET stage = ? WHERE chat_id = ?`, int(stage), chatID)
	if err != nil {
		return fmt.Errorf("sqlite.Exec: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) GetStage(ctx context.Context, chatID int64) (domain.Stage, error) {
	row := s.db.QueryRowContext(ctx, `SELECT stage FROM chats WHERE chat_id = ?`, chatID)
	var stage int
	if err := row.Scan(&stage); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Unknown, errs.ErrNotFound
		}
		return domain.Unknown, fmt.Errorf("sqlite.QueryRow: %w", err)
	}
	return domain.Stage(stage), nil
}

func (s *SQLiteStorage) GetObservers(ctx context.Context) (map[int64]*domain.Chat, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT chat_id, username, phone, role, stage FROM chats WHERE role = 2`)
	if err != nil {
		return nil, fmt.Errorf("sqlite.Query: %w", err)
	}
	defer rows.Close()

	observers := make(map[int64]*domain.Chat)
	for rows.Next() {
		var chat domain.Chat
		var role, stage int
		if err := rows.Scan(&chat.ID, &chat.Username, &chat.Phone, &role, &stage); err != nil {
			return nil, fmt.Errorf("sqlite.Scan: %w", err)
		}
		chat.Role = domain.Role(role)
		chat.Stage = domain.Stage(stage)
		observers[chat.ID] = &chat
	}
	return observers, nil
}

func (s *SQLiteStorage) AddTask(ctx context.Context, task domain.Task) (int, error) {
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO tasks (title, executor_contact, executor_chat_id, deadline, done, closed, expired) 
		VALUES (?, ?, ?, ?, false, false, false)`, task.Title, task.ExecutorContact, task.ExecutorChatID, task.Deadline)
	if err != nil {
		return -1, fmt.Errorf("sqlite.Exec: %w", err)
	}
	taskID, err := result.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("sqlite.LastInsertId: %w", err)
	}
	return int(taskID), nil
}

func (s *SQLiteStorage) GetAllTasks(ctx context.Context) ([]domain.Task, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, executor_contact, executor_chat_id, deadline, done, closed, expired FROM tasks`)
	if err != nil {
		return nil, fmt.Errorf("sqlite.Query: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var task domain.Task
		if err := rows.Scan(&task.ID, &task.Title, &task.ExecutorContact, &task.ExecutorChatID, &task.Deadline, &task.Done, &task.Closed, &task.Expired); err != nil {
			return nil, fmt.Errorf("sqlite.Scan: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (s *SQLiteStorage) GetClosedTasks(ctx context.Context) ([]domain.Task, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, executor_contact, executor_chat_id, deadline, done, closed, expired FROM tasks WHERE closed = true`)
	if err != nil {
		return nil, fmt.Errorf("sqlite.Query: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var task domain.Task
		if err := rows.Scan(&task.ID, &task.Title, &task.ExecutorContact, &task.ExecutorChatID, &task.Deadline, &task.Done, &task.Closed, &task.Expired); err != nil {
			return nil, fmt.Errorf("sqlite.Scan: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (s *SQLiteStorage) GetOpenTasks(ctx context.Context) ([]domain.Task, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, executor_contact, executor_chat_id, deadline, done, closed, expired FROM tasks WHERE closed = false`)
	if err != nil {
		return nil, fmt.Errorf("sqlite.Query: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var task domain.Task
		if err := rows.Scan(&task.ID, &task.Title, &task.ExecutorContact, &task.ExecutorChatID, &task.Deadline, &task.Done, &task.Closed, &task.Expired); err != nil {
			return nil, fmt.Errorf("sqlite.Scan: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (s *SQLiteStorage) GetDoneTasks(ctx context.Context) ([]domain.Task, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, executor_contact, executor_chat_id, deadline, done, closed, expired FROM tasks WHERE done = true`)
	if err != nil {
		return nil, fmt.Errorf("sqlite.Query: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var task domain.Task
		if err := rows.Scan(&task.ID, &task.Title, &task.ExecutorContact, &task.ExecutorChatID, &task.Deadline, &task.Done, &task.Closed, &task.Expired); err != nil {
			return nil, fmt.Errorf("sqlite.Scan: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (s *SQLiteStorage) GetExpiredTasks(ctx context.Context) ([]domain.Task, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, executor_contact, executor_chat_id, deadline, done, closed, expired FROM tasks WHERE expired = true`)
	if err != nil {
		return nil, fmt.Errorf("sqlite.Query: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var task domain.Task
		if err := rows.Scan(&task.ID, &task.Title, &task.ExecutorContact, &task.ExecutorChatID, &task.Deadline, &task.Done, &task.Closed, &task.Expired); err != nil {
			return nil, fmt.Errorf("sqlite.Scan: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (s *SQLiteStorage) GetExpiredTasksToMark(ctx context.Context) ([]domain.Task, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, executor_contact, executor_chat_id, deadline, done, closed, expired FROM tasks WHERE done = false AND expired = false AND deadline < ?`, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("sqlite.Query: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var task domain.Task
		if err := rows.Scan(&task.ID, &task.Title, &task.ExecutorContact, &task.ExecutorChatID, &task.Deadline, &task.Done, &task.Closed, &task.Expired); err != nil {
			return nil, fmt.Errorf("sqlite.Scan: %w", err)
		}
		task.Expired = true
		tasks = append(tasks, task)
		if _, err := s.db.ExecContext(ctx, `UPDATE tasks SET expired = true WHERE id = ?`, task.ID); err != nil {
			return nil, fmt.Errorf("sqlite.Exec: %w", err)
		}
	}
	return tasks, nil
}

func (s *SQLiteStorage) GetUserTasks(ctx context.Context, username, phone string) ([]domain.Task, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, title, executor_contact, executor_chat_id, deadline, done, closed, expired FROM tasks WHERE executor_contact = ? OR executor_contact = ?`, username, phone)
	if err != nil {
		return nil, fmt.Errorf("sqlite.Query: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		var task domain.Task
		if err := rows.Scan(&task.ID, &task.Title, &task.ExecutorContact, &task.ExecutorChatID, &task.Deadline, &task.Done, &task.Closed, &task.Expired); err != nil {
			return nil, fmt.Errorf("sqlite.Scan: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (s *SQLiteStorage) MarkTaskAsDone(ctx context.Context, taskID int) error {
	result, err := s.db.ExecContext(ctx, `UPDATE tasks SET done = true WHERE id = ?`, taskID)
	if err != nil {
		return fmt.Errorf("sqlite.Exec: %w", err)
	}
	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("sqlite.RowsAffected: %w", err)
	}
	if affectedRows == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (s *SQLiteStorage) MarkTaskAsClosed(ctx context.Context, taskID int) error {
	result, err := s.db.ExecContext(ctx, `UPDATE tasks SET closed = true WHERE id = ?`, taskID)
	if err != nil {
		return fmt.Errorf("sqlite.Exec: %w", err)
	}
	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("sqlite.RowsAffected: %w", err)
	}
	if affectedRows == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (s *SQLiteStorage) DeleteTask(ctx context.Context, taskID int) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, taskID)
	if err != nil {
		return fmt.Errorf("sqlite.Exec: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) ChangeTaskDeadline(ctx context.Context, taskID int, newDeadline time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE tasks SET deadline = ?, expired = false WHERE id = ?`, newDeadline, taskID)
	if err != nil {
		return fmt.Errorf("sqlite.Exec: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) GetTaskInProgress(ctx context.Context, chatID int64) (domain.Task, error) {
	row := s.db.QueryRowContext(ctx, `SELECT title, executor_contact, executor_chat_id, deadline FROM tasks_in_progress WHERE chat_id = ?`, chatID)
	var task domain.Task
	var deadline sql.NullTime
	if err := row.Scan(&task.Title, &task.ExecutorContact, &task.ExecutorChatID, &deadline); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Task{}, errs.ErrNotFound
		}
		return domain.Task{}, fmt.Errorf("sqlite.QueryRow: %w", err)
	}
	if deadline.Valid {
		task.Deadline = deadline.Time
	}
	return task, nil
}

func (s *SQLiteStorage) SetTaskInProgressName(ctx context.Context, chatID int64, name string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO tasks_in_progress (chat_id, title) VALUES (?, ?)
		ON CONFLICT(chat_id) DO UPDATE SET title = EXCLUDED.title`, chatID, name)
	if err != nil {
		return fmt.Errorf("sqlite.Exec: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) SetTaskInProgressUser(ctx context.Context, chatID int64, userContact string, userChatID int64) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO tasks_in_progress (chat_id, executor_contact, executor_chat_id) VALUES (?, ?, ?)
		ON CONFLICT(chat_id) DO UPDATE SET executor_contact = EXCLUDED.executor_contact, executor_chat_id = EXCLUDED.executor_chat_id`, chatID, userContact, userChatID)
	if err != nil {
		return fmt.Errorf("sqlite.Exec: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) SetTaskInProgressDeadline(ctx context.Context, chatID int64, deadline time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO tasks_in_progress (chat_id, deadline) VALUES (?, ?)
		ON CONFLICT(chat_id) DO UPDATE SET deadline = EXCLUDED.deadline`, chatID, deadline)
	if err != nil {
		return fmt.Errorf("sqlite.Exec: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) AddMessage(ctx context.Context, message domain.Message) error {
	return nil
}

func (s *SQLiteStorage) RetrieveMessages(ctx context.Context) ([]domain.Message, error) {
	return nil, nil
}

func (s *SQLiteStorage) SetHandledMessage(ctx context.Context, messageID int) error {
	return nil
}
