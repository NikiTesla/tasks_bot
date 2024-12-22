package domain

type Chat struct {
	ID       int64
	Username string
	Stage    Stage
	Role     Role
}
