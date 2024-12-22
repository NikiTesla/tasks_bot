-- Schema for chats table
CREATE TABLE chats (
    chat_id BIGINT PRIMARY KEY,
    username TEXT UNIQUE,
    phone TEXT UNIQUE,
    role INT DEFAULT 0,
    stage INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Schema for tasks table
CREATE TABLE tasks (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    executor_contact TEXT NOT NULL,
    executor_chat_id BIGINT,
    deadline TIMESTAMP NOT NULL,
    done BOOLEAN NOT NULL DEFAULT FALSE,
    closed BOOLEAN NOT NULL DEFAULT FALSE,
    expired BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Schema for tasks_in_progress table
CREATE TABLE tasks_in_progress (
    chat_id BIGINT PRIMARY KEY,
    title TEXT,
    executor_contact TEXT,
    executor_chat_id BIGINT,
    deadline TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
