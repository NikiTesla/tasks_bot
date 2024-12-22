package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"tasks_bot/internal/domain"

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

	query := `CREATE TABLE IF NOT EXISTS chats (id INTEGER PRIMARY KEY,
		username TEXT,
		step INT DEFAULT -1,
		correct_answers INT DEFAULT 0,
        is_admin BOOLEAN DEFAULT 0)`
	if _, err := db.ExecContext(ctx, query); err != nil {
		return nil, fmt.Errorf("db.ExecContext: failed to create table 'chats': %w", err)
	}

	return db, nil
}

func (s *SQLiteStorage) AddChat(ctx context.Context, chatID int64, username, phone string, role domain.Role) error {
	query := `INSERT INTO chats (id, username, step, is_admin) VALUES (?, ?, -1, ?) 
		ON CONFLICT(id) DO UPDATE SET is_admin=excluded.is_admin, username=excluded.username, step=-1, correct_answers=0`
	if _, err := s.db.ExecContext(ctx, query, chatID, username, role); err != nil {
		return fmt.Errorf("db.Exec: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) IsAdmin(ctx context.Context, chatID int64) (bool, error) {
	query := `SELECT is_admin FROM chats WHERE id = ?`
	isAdmin := false
	if err := s.db.QueryRowContext(ctx, query, chatID).Scan(&isAdmin); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("db.QueryRow: %w", err)
	}
	return isAdmin, nil
}
