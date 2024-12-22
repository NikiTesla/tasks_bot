package domain

type Stage int

const (
	Unknown Stage = iota
	Default
	BecomeExecutor
	BecomeObserver
	BecomeChief
	BecomeAdmin
	AddTaskName
	AddTaskUser
	AddTaskDeadline
	MarkTaskAsDone
	MarkTaskAsClosed
	ChangeDeadline
	ContactRequest
	DeleteTask
)
