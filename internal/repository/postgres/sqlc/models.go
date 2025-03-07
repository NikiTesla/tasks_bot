// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package queries

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Chat struct {
	ChatID    int64            `json:"chat_id"`
	Username  pgtype.Text      `json:"username"`
	Phone     pgtype.Text      `json:"phone"`
	Role      pgtype.Int4      `json:"role"`
	Stage     pgtype.Int4      `json:"stage"`
	CreatedAt pgtype.Timestamp `json:"created_at"`
}

type Task struct {
	ID              int64            `json:"id"`
	Title           string           `json:"title"`
	ExecutorContact string           `json:"executor_contact"`
	ExecutorChatID  pgtype.Int8      `json:"executor_chat_id"`
	Deadline        pgtype.Timestamp `json:"deadline"`
	Done            bool             `json:"done"`
	Closed          bool             `json:"closed"`
	Expired         bool             `json:"expired"`
	CreatedAt       pgtype.Timestamp `json:"created_at"`
}

type TasksInProgress struct {
	ChatID          int64            `json:"chat_id"`
	Title           pgtype.Text      `json:"title"`
	ExecutorContact pgtype.Text      `json:"executor_contact"`
	ExecutorChatID  pgtype.Int8      `json:"executor_chat_id"`
	Deadline        pgtype.Timestamp `json:"deadline"`
	CreatedAt       pgtype.Timestamp `json:"created_at"`
}
