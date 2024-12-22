package domain

type Chat struct {
	ID       int64
	Username string
	Phone    string
	Stage    Stage
	Role     Role
}
