// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package queries

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Chat struct {
	ChatID   int64       `json:"chat_id"`
	Username string      `json:"username"`
	Role     pgtype.Int4 `json:"role"`
	Stage    pgtype.Int4 `json:"stage"`
}

type Task struct {
	ID       int64            `json:"id"`
	Title    string           `json:"title"`
	Executor string           `json:"executor"`
	Deadline pgtype.Timestamp `json:"deadline"`
	Done     bool             `json:"done"`
	Closed   bool             `json:"closed"`
	Expired  bool             `json:"expired"`
}

type TasksInProgress struct {
	ChatID   int64            `json:"chat_id"`
	Title    pgtype.Text      `json:"title"`
	Executor pgtype.Text      `json:"executor"`
	Deadline pgtype.Timestamp `json:"deadline"`
}
