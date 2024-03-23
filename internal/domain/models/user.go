package models

type User struct {
	Id         int64
	PassHash   []byte
	Role       UserRole
	RoleString string
	Email      string
	Deleted    bool
}
