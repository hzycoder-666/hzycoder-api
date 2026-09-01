package model

import "time"

type Role string

type User struct {
	ID        int64     `db:"id"`
	Username  string    `db:"username"`
	Password  string    `db:"password"`
	Nickname  *string   `db:"nickname"`
	Role      Role      `db:"role"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

const (
	RoleGuest  Role = "guest"
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
)

func IsAdmin(r Role) bool {
	return r == RoleAdmin
}

func IsValid(r Role) bool {
	switch r {
	case RoleGuest, RoleMember, RoleAdmin:
		return true
	default:
		return false
	}
}
