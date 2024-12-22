-- name: GetRole :one
SELECT role FROM chats WHERE chat_id = $1;

-- name: SetRole :exec
UPDATE chats SET role = $2 WHERE chat_id = $1;

-- name: AddChat :exec
INSERT INTO chats (chat_id, username, phone, role) VALUES ($1, $2, $3, $4) ON CONFLICT (chat_id)
DO UPDATE SET username = COALESCE(NULLIF(EXCLUDED.username, ''), chats.username), phone = COALESCE(NULLIF(EXCLUDED.phone, ''), chats.phone);

-- name: GetChat :one
SELECT * FROM chats WHERE username = $1 OR phone = $2;

-- name: GetObservers :many
SELECT * FROM chats WHERE role = 2;

-- name: SetStage :exec
UPDATE chats SET stage = $2 WHERE chat_id = $1;

-- name: GetStage :one
SELECT stage FROM chats WHERE chat_id = $1;

-- name: GetAllTasks :many
SELECT * FROM tasks;

-- name: GetExpiredTasks :many
SELECT * FROM tasks WHERE expired = true;

-- name: GetExpiredTasksToMark :many
SELECT * FROM tasks WHERE done = false AND expired = false AND deadline < (NOW() AT TIME ZONE 'UTC-3') FOR UPDATE;

-- name: GetOpenTasks :many
SELECT * FROM tasks WHERE closed = false;

-- name: GetDoneTasks :many
SELECT * FROM tasks WHERE done = true;

-- name: GetClosedTasks :many
SELECT * FROM tasks WHERE closed = true;

-- name: GetUserTasks :many
SELECT * FROM tasks WHERE executor_contact = $1;

-- name: AddTask :one
INSERT INTO tasks (title, executor_contact, executor_chat_id, deadline, done, closed, expired) VALUES ($1, $2, $3, $4, false, false, false) RETURNING id;

-- name: MarkTaskAsDone :execrows
UPDATE tasks SET done = true WHERE id = $1;

-- name: MarkTaskAsClosed :execrows
UPDATE tasks SET closed = true WHERE id = $1;

-- name: MarkExpiredTask :execrows
UPDATE tasks SET expired = true WHERE id = $1;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = $1;

-- name: ChangeTaskDeadline :exec
UPDATE tasks SET deadline = $2, expired = false WHERE id = $1;

-- name: GetTaskInProgress :one
SELECT * FROM tasks_in_progress WHERE chat_id = $1;

-- name: SetTaskInProgressName :exec
INSERT INTO tasks_in_progress (chat_id, title) VALUES ($1, $2) 
ON CONFLICT (chat_id) DO UPDATE SET title = EXCLUDED.title;

-- name: SetTaskInProgressUser :exec
INSERT INTO tasks_in_progress (chat_id, executor_contact, executor_chat_id) VALUES ($1, $2, $3) 
ON CONFLICT (chat_id) DO UPDATE SET executor_contact = EXCLUDED.executor_contact, executor_chat_id = EXCLUDED.executor_chat_id;

-- name: SetTaskInProgressDeadline :exec
INSERT INTO tasks_in_progress (chat_id, deadline) VALUES ($1, $2) 
ON CONFLICT (chat_id) DO UPDATE SET deadline = EXCLUDED.deadline;
