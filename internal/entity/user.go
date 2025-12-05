package entity

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

type User struct {
	ID        string
	Username  string
	FirstName string
	LastName  string
	Role      *Role
}
