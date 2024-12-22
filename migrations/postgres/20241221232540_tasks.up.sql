-- Schema for chats table
CREATE TABLE chats (
    chat_id BIGINT PRIMARY KEY,
    username TEXT NOT NULL,
    role INT DEFAULT 0,
    stage INT DEFAULT 0
);

-- Schema for tasks table
CREATE TABLE tasks (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    executor TEXT NOT NULL,
    deadline TIMESTAMP NOT NULL,
    done BOOLEAN NOT NULL DEFAULT FALSE,
    closed BOOLEAN NOT NULL DEFAULT FALSE,
    expired BOOLEAN NOT NULL DEFAULT FALSE
);

-- Schema for tasks_in_progress table
CREATE TABLE tasks_in_progress (
    chat_id BIGINT PRIMARY KEY,
    title TEXT,
    executor TEXT,
    deadline TIMESTAMP
);
