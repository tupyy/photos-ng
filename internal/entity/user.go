package entity

import "fmt"

type Role int

const (
	AdminRole Role = iota
	CreatorRole
	EditorRole
	ViewerRole
)

func (r Role) String() string {
	switch r {
	case AdminRole:
		return "admin"
	case CreatorRole:
		return "creator"
	case EditorRole:
		return "editor"
	case ViewerRole:
		return "viewer"
	default:
		return "unknown"
	}
}

func RoleFromString(s string) (Role, error) {
	switch s {
	case "admin":
		return AdminRole, nil
	case "creator":
		return CreatorRole, nil
	case "editor":
		return EditorRole, nil
	case "viewer":
		return ViewerRole, nil
	default:
		return 0, fmt.Errorf("unknown role: %s", s)
	}
}

type User struct {
	ID        string
	Username  string
	FirstName string
	LastName  string
	Role      *Role
}
