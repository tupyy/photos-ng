package entity

const (
	LocalDatastore  = "local"
	AdminRoleName   = "admin"
	ViewerRoleName  = "viewer"
	EditorRoleName  = "editor"
	CreatorRoleName = "creator"
)

var (
	UserPermissions      = []Permission{ViewPermission, EditPermission, DeletePermission}
	DatastorePermissions = []Permission{CreatePermission, ViewPermission, EditPermission, DeletePermission}
	AllPermissions       = []Permission{ViewPermission, EditPermission, EditPermission, CreatePermission}
)

type Permission int

func (p Permission) String() string {
	switch p {
	case ViewPermission:
		return "view"
	case EditPermission:
		return "edit"
	case DeletePermission:
		return "delete"
	case CreatePermission:
		return "create"
	case SyncPermission:
		return "sync"
	default:
		return "unknown"
	}
}

const (
	ViewPermission Permission = iota
	EditPermission
	CreatePermission
	DeletePermission
	SyncPermission
)

type ResourceKind int

func (rk ResourceKind) String() string {
	switch rk {
	case AlbumResource:
		return "album"
	case MediaResource:
		return "media"
	case DatastoreResource:
		return "datastore"
	case RoleResource:
		return "role"
	default:
		return "unknown"
	}
}

const (
	AlbumResource ResourceKind = iota
	MediaResource
	DatastoreResource
	RoleResource
)

type SubjectKind int

func (sk SubjectKind) String() string {
	switch sk {
	case RoleSubject:
		return "role"
	case UserSubject:
		return "user"
	case AlbumSubject:
		return "album"
	case DatastoreSubject:
		return "datastore"
	default:
		return "unknown"
	}
}

const (
	RoleSubject SubjectKind = iota
	UserSubject
	AlbumSubject
	DatastoreSubject
)

type Resource struct {
	ID   string
	Kind ResourceKind
}

func NewRoleResource(id string) Resource {
	return Resource{ID: id, Kind: RoleResource}
}

func NewAlbumResource(id string) Resource {
	return Resource{ID: id, Kind: AlbumResource}
}

func NewMediaResource(id string) Resource {
	return Resource{ID: id, Kind: MediaResource}
}

func NewDatastoreResource(id string) Resource {
	return Resource{ID: id, Kind: DatastoreResource}
}

type Subject struct {
	ID   string
	Kind SubjectKind
}

func NewUserSubject(userID string) Subject {
	return Subject{ID: userID, Kind: UserSubject}
}

func NewAlbumSubject(id string) Subject {
	return Subject{ID: id, Kind: AlbumSubject}
}

func NewRoleSubject(id string) Subject {
	return Subject{ID: id, Kind: RoleSubject}
}

func NewDatastoreSubject(id string) Subject {
	return Subject{ID: id, Kind: DatastoreSubject}
}

type Relationship struct {
	Subject  Subject
	Resource Resource
	Kind     RelationshipKind
}

func NewRelationship(subject Subject, resource Resource, kind RelationshipKind) Relationship {
	return Relationship{
		Subject:  subject,
		Resource: resource,
		Kind:     kind,
	}
}

type RelationshipKind int

func (rk RelationshipKind) String() string {
	switch rk {
	case AdminRelationship:
		return "admin"
	case CreatorRelationship:
		return "creator"
	case ViewerRelationship:
		return "viewer"
	case EditorRelationship:
		return "editor"
	case ParentRelationship:
		return "parent"
	case OwnerRelationship:
		return "owner"
	case DatastoreRelationship:
		return "datastore"
	case MemberRelationship:
		return "member"
	default:
		return "unknown"
	}
}

const (
	AdminRelationship RelationshipKind = iota
	CreatorRelationship
	ViewerRelationship
	EditorRelationship
	ParentRelationship
	OwnerRelationship
	DatastoreRelationship
	MemberRelationship
)
