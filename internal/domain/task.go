package domain

import (
	"fmt"
	"time"
)

const (
	DeadlineLayout = "02.01.2006 15:04:05"

	UnknownTask TaskStatus = iota
	ClosedTask
	DoneTask
	ExpiredTask
	OpenTask
)

type TaskStatus int

type Task struct {
	ID       int
	Title    string
	Executor string
	Deadline time.Time
	Done     bool
	Expired  bool
	Closed   bool
}

func (t Task) String() string {
	if t.Done || t.Closed {
		return fmt.Sprintf("<b>Задача №%d</b>\n<b>Название:</b> %s\n<b>Дедлайн %s</b>\n<b>Статус:</b> %s\n<b>Исполнитель:</b> @%s",
			t.ID,
			t.Title,
			t.Deadline.Format(DeadlineLayout),
			t.GetStatus(),
			t.Executor,
		)
	}

	if time.Now().After(t.Deadline) {
		return fmt.Sprintf("<b>Задача №%d</b>\n<b>Название:</b> %s\n<b>Дедлайн:</b> %s\n<b>Статус:</b> %s\n<b>Исполнитель:</b> @%s",
			t.ID,
			t.Title,
			t.Deadline.Format(DeadlineLayout),
			ExpiredTask,
			t.Executor,
		)
	}
	return fmt.Sprintf("<b>Задача №%d</b>\n<b>Название:</b> %s\n<b>Дедлайн %s</b>\n<b>Статус:</b> %s\n<b>Исполнитель:</b> @%s",
		t.ID,
		t.Title,
		t.Deadline.Format(DeadlineLayout),
		t.GetStatus(),
		t.Executor,
	)
}

func (t Task) GetStatus() TaskStatus {
	switch {
	case t.Closed:
		return ClosedTask
	case t.Done:
		return DoneTask
	case t.Expired:
		return ExpiredTask
	default:
		return OpenTask
	}
}

func (ts TaskStatus) String() string {
	switch ts {
	case ClosedTask:
		return "закрыта"
	case DoneTask:
		return "выполнена"
	case ExpiredTask:
		return "просрочена"
	case OpenTask:
		return "открыта"
	case UnknownTask:
		return "неизвестен"
	default:
		return "unknown"
	}
}
