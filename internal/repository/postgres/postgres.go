package postgres

import (
	"context"
	"errors"
	"fmt"
	"tasks_bot/internal/config"
	"tasks_bot/internal/domain"
	"tasks_bot/internal/errs"
	queries "tasks_bot/internal/repository/postgres/sqlc"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Writable struct {
	db *pgxpool.Pool
}

func NewWritable(ctx context.Context, postgresCfg *config.PostgresConfig) (*Writable, error) {
	db, err := pgxpool.New(ctx, postgresCfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.Connect: %w", err)
	}

	return &Writable{
		db: db,
	}, nil
}

func (p *Writable) Close() {
	p.db.Close()
}

func (p *Writable) DebugStorage(ctx context.Context) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (p *Writable) AddChat(ctx context.Context, chatID int64, username, phone string, role domain.Role) error {
	err := queries.New(p.db).AddChat(ctx, &queries.AddChatParams{
		ChatID:   chatID,
		Username: username,
		Phone:    phone,
		Role:     pgtype.Int4{Int32: int32(role), Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("pgx.Query: %w", err)
	}
	return nil
}

func (p *Writable) GetChat(ctx context.Context, username, phone string) (*domain.Chat, error) {
	chat, err := queries.New(p.db).GetChat(ctx, &queries.GetChatParams{
		Phone:    phone,
		Username: username,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("pgx.Query: %w", err)
	}
	return &domain.Chat{
		ID:       chat.ChatID,
		Username: username,
		Stage:    domain.Stage(chat.Stage.Int32),
		Role:     domain.Role(chat.Role.Int32),
	}, nil
}

func (p *Writable) GetRole(ctx context.Context, chatID int64) (domain.Role, error) {
	role, err := queries.New(p.db).GetRole(ctx, chatID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.UnknownRole, errs.ErrNotFound
		}
		return domain.UnknownRole, fmt.Errorf("pgx.Query: %w", err)
	}
	return domain.Role(role.Int32), nil
}

func (p *Writable) SetRole(ctx context.Context, chatID int64, role domain.Role) error {
	err := queries.New(p.db).SetRole(ctx, &queries.SetRoleParams{
		ChatID: chatID,
		Role:   pgtype.Int4{Int32: int32(role), Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("pgx.Query: %w", err)
	}
	return nil
}

func (p *Writable) SetStage(ctx context.Context, chatID int64, stage domain.Stage) error {
	err := queries.New(p.db).SetStage(ctx, &queries.SetStageParams{
		ChatID: chatID,
		Stage:  pgtype.Int4{Int32: int32(stage), Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("pgx.Query: %w", err)
	}
	return nil
}

func (p *Writable) GetStage(ctx context.Context, chatID int64) (domain.Stage, error) {
	role, err := queries.New(p.db).GetStage(ctx, chatID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Unknown, errs.ErrNotFound
		}
		return domain.Unknown, fmt.Errorf("pgx.Query: %w", err)
	}
	return domain.Stage(role.Int32), nil
}

func (p *Writable) GetObservers(ctx context.Context) (map[int64]*domain.Chat, error) {
	queriesObservers, err := queries.New(p.db).GetObservers(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("pgx.Query: %w", err)
	}
	observers := make(map[int64]*domain.Chat, len(queriesObservers))
	for _, observer := range queriesObservers {
		observers[observer.ChatID] = ChatToDomain(observer)
	}
	return observers, nil
}

func (p *Writable) AddTask(ctx context.Context, task domain.Task) (int, error) {
	taskID, err := queries.New(p.db).AddTask(ctx, &queries.AddTaskParams{
		Title:           task.Title,
		ExecutorContact: task.ExecutorContact,
		ExecutorChatID:  pgtype.Int8{Int64: task.ExecutorChatID, Valid: true},
		Deadline:        pgtype.Timestamp{Time: task.Deadline, Valid: true},
	})
	if err != nil {
		return -1, fmt.Errorf("pgx.Query: %w", err)
	}
	return int(taskID + 1), nil
}

func (p *Writable) GetAllTasks(ctx context.Context) ([]domain.Task, error) {
	queriesTasks, err := queries.New(p.db).GetAllTasks(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("pgx.Query: %w", err)
	}
	tasks := make([]domain.Task, 0, len(queriesTasks))
	for _, task := range queriesTasks {
		tasks = append(tasks, TaskToDomain(task))
	}
	return tasks, nil
}

func (p *Writable) GetClosedTasks(ctx context.Context) ([]domain.Task, error) {
	queriesTasks, err := queries.New(p.db).GetClosedTasks(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("pgx.Query: %w", err)
	}
	tasks := make([]domain.Task, 0, len(queriesTasks))
	for _, task := range queriesTasks {
		tasks = append(tasks, TaskToDomain(task))
	}
	return tasks, nil
}

func (p *Writable) GetOpenTasks(ctx context.Context) ([]domain.Task, error) {
	queriesTasks, err := queries.New(p.db).GetOpenTasks(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("pgx.Query: %w", err)
	}
	tasks := make([]domain.Task, 0, len(queriesTasks))
	for _, task := range queriesTasks {
		tasks = append(tasks, TaskToDomain(task))
	}
	return tasks, nil
}

func (p *Writable) GetDoneTasks(ctx context.Context) ([]domain.Task, error) {
	queriesTasks, err := queries.New(p.db).GetDoneTasks(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("pgx.Query: %w", err)
	}
	tasks := make([]domain.Task, 0, len(queriesTasks))
	for _, task := range queriesTasks {
		tasks = append(tasks, TaskToDomain(task))
	}
	return tasks, nil
}

func (p *Writable) GetExpiredTasks(ctx context.Context) ([]domain.Task, error) {
	q := queries.New(p.db)
	queriesTasks, err := q.GetExpiredTasks(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("pgx.Query: %w", err)
	}
	tasks := make([]domain.Task, 0, len(queriesTasks))
	for _, task := range queriesTasks {
		tasks = append(tasks, TaskToDomain(task))
	}
	return tasks, nil
}

func (p *Writable) GetExpiredTasksToMark(ctx context.Context) ([]domain.Task, error) {
	q := queries.New(p.db)
	queriesTasks, err := q.GetExpiredTasksToMark(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("pgx.Query: %w", err)
	}
	tasks := make([]domain.Task, 0, len(queriesTasks))
	for _, task := range queriesTasks {
		task.Expired = true
		tasks = append(tasks, TaskToDomain(task))
		if affected, err := q.MarkExpiredTask(ctx, task.ID); err != nil || affected == 0 {
			return nil, fmt.Errorf("q.MarkExpiredTask (%d): %w", task.ID, err)
		}
	}
	return tasks, nil
}

func (p *Writable) GetUserTasks(ctx context.Context, username string) ([]domain.Task, error) {
	queriesTasks, err := queries.New(p.db).GetUserTasks(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrNotFound
		}
		return nil, fmt.Errorf("pgx.Query: %w", err)
	}
	tasks := make([]domain.Task, 0, len(queriesTasks))
	for _, task := range queriesTasks {
		tasks = append(tasks, TaskToDomain(task))
	}
	return tasks, nil
}

func (p *Writable) MarkTaskAsDone(ctx context.Context, taskID int) error {
	affectedRows, err := queries.New(p.db).MarkTaskAsDone(ctx, int64(taskID-1))
	if err != nil {
		return fmt.Errorf("pgx.Query: %w", err)
	}
	if affectedRows == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (p *Writable) MarkTaskAsClosed(ctx context.Context, taskID int) error {
	affectedRows, err := queries.New(p.db).MarkTaskAsClosed(ctx, int64(taskID-1))
	if err != nil {
		return fmt.Errorf("pgx.Query: %w", err)
	}
	if affectedRows == 0 {
		return errs.ErrNotFound
	}
	return nil
}

func (p *Writable) ChangeTaskDeadline(ctx context.Context, taskID int, newDeadline time.Time) error {
	err := queries.New(p.db).ChangeTaskDeadline(ctx, &queries.ChangeTaskDeadlineParams{
		ID:       int64(taskID - 1),
		Deadline: pgtype.Timestamp{Time: newDeadline, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("pgx.Query: %w", err)
	}
	return nil
}

func (p *Writable) GetTaskInProgress(ctx context.Context, chatID int64) (domain.Task, error) {
	queriesTaskInProgress, err := queries.New(p.db).GetTaskInProgress(ctx, chatID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Task{}, errs.ErrNotFound
		}
		return domain.Task{}, fmt.Errorf("pgx.Query: %w", err)
	}
	return domain.Task{
		Title:           queriesTaskInProgress.Title.String,
		ExecutorContact: queriesTaskInProgress.ExecutorContact.String,
		ExecutorChatID:  queriesTaskInProgress.ExecutorChatID.Int64,
		Deadline:        queriesTaskInProgress.Deadline.Time,
	}, nil
}

func (p *Writable) SetTaskInProgressName(ctx context.Context, chatID int64, name string) error {
	err := queries.New(p.db).SetTaskInProgressName(ctx, &queries.SetTaskInProgressNameParams{
		ChatID: chatID,
		Title:  pgtype.Text{String: name, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("pgx.Query: %w", err)
	}
	return nil
}

func (p *Writable) SetTaskInProgressUser(ctx context.Context, chatID int64, userContact string, userChatID int64) error {
	err := queries.New(p.db).SetTaskInProgressUser(ctx, &queries.SetTaskInProgressUserParams{
		ChatID:          chatID,
		ExecutorContact: pgtype.Text{String: userContact, Valid: true},
		ExecutorChatID:  pgtype.Int8{Int64: userChatID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("pgx.Query: %w", err)
	}
	return nil
}

func (p *Writable) SetTaskInProgressDeadline(ctx context.Context, chatID int64, deadline time.Time) error {
	err := queries.New(p.db).SetTaskInProgressDeadline(ctx, &queries.SetTaskInProgressDeadlineParams{
		ChatID:   chatID,
		Deadline: pgtype.Timestamp{},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errs.ErrNotFound
		}
		return fmt.Errorf("pgx.Query: %w", err)
	}
	return nil
}

func (p *Writable) AddMessage(ctx context.Context, message domain.Message) error {
	return nil
}

func (p *Writable) RetrieveMessages(ctx context.Context) ([]domain.Message, error) {
	return nil, nil
}

func (p *Writable) SetHandledMessage(ctx context.Context, messageID int) error {
	return nil
}
