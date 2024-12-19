package domain

type Role int

const (
	Executor Role = iota
	Observer
	Chief
	Admin
)

func (r Role) String() string {
	switch r {
	case Executor:
		return "executor"
	case Observer:
		return "observer"
	case Chief:
		return "chief"
	case Admin:
		return "admin"
	default:
		return ""
	}
}
