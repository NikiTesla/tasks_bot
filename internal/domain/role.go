package domain

type Role int

const (
	UnknownRole Role = iota
	Executor
	Observer
	Chief
	Admin
)

func (r Role) String() string {
	switch r {
	case Executor:
		return "исполнитель"
	case Observer:
		return "наблюдатель"
	case Chief:
		return "шеф"
	case Admin:
		return "админ"
	case UnknownRole:
		return "неизвестно"
	default:
		return "неизвестно"
	}
}
