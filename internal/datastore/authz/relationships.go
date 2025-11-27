package authz

import (
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
)

// WithRelationship creates a relationship function that adds a relationship between a subject and a resource.
//
// Parameters:
//   - subject: The subject of the relationship (user, role, album, datastore)
//   - resource: The resource of the relationship (album, media, datastore, role)
//   - relationshipKind: The kind of relationship (owner, viewer, editor, parent, member, etc.)
//
// Returns:
//   - RelationshipFn: A function that can be used with WriteRelationships
//
// Example:
//
//	err := authzService.WriteRelationships(ctx,
//		WithRelationship(entity.NewUserSubject("user123"), entity.NewAlbumResource("album789"), entity.OwnerRelationship))
func WithRelationship(subject entity.Subject, resource entity.Resource, relationshipKind entity.RelationshipKind) RelationshipFn {
	return func(updates []*v1.RelationshipUpdate) []*v1.RelationshipUpdate {
		subjectRef := &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: subject.Kind.String(),
				ObjectId:   subject.ID,
			},
		}

		// Add OptionalRelation for role subjects (except for member relationships on roles)
		if subject.Kind == entity.RoleSubject {
			if resource.Kind != entity.RoleResource || relationshipKind != entity.MemberRelationship {
				subjectRef.OptionalRelation = "member"
			}
		}

		relationshipUpdate := &v1.RelationshipUpdate{
			Operation: v1.RelationshipUpdate_OPERATION_TOUCH,
			Relationship: &v1.Relationship{
				Resource: &v1.ObjectReference{
					ObjectType: resource.Kind.String(),
					ObjectId:   resource.ID,
				},
				Relation: relationshipKind.String(),
				Subject:  subjectRef,
			},
		}
		return append(updates, relationshipUpdate)
	}
}

// WithoutRelationship creates a relationship function that removes a relationship between a subject and a resource.
//
// Parameters:
//   - subject: The subject of the relationship (user, role, album, datastore)
//   - resource: The resource of the relationship (album, media, datastore, role)
//   - relationshipKind: The kind of relationship (owner, viewer, editor, parent, member, etc.)
//
// Returns:
//   - RelationshipFn: A function that can be used with WriteRelationships
//
// Example:
//
//	err := authzService.WriteRelationships(ctx,
//		WithoutRelationship(entity.NewUserSubject("user123"), entity.NewAlbumResource("album789"), entity.OwnerRelationship))
func WithoutRelationship(subject entity.Subject, resource entity.Resource, relationshipKind entity.RelationshipKind) RelationshipFn {
	return func(updates []*v1.RelationshipUpdate) []*v1.RelationshipUpdate {
		subjectRef := &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: subject.Kind.String(),
				ObjectId:   subject.ID,
			},
		}

		// Add OptionalRelation for role subjects (except for member relationships on roles)
		if subject.Kind == entity.RoleSubject {
			if resource.Kind != entity.RoleResource || relationshipKind != entity.MemberRelationship {
				subjectRef.OptionalRelation = "member"
			}
		}

		relationshipUpdate := &v1.RelationshipUpdate{
			Operation: v1.RelationshipUpdate_OPERATION_DELETE,
			Relationship: &v1.Relationship{
				Resource: &v1.ObjectReference{
					ObjectType: resource.Kind.String(),
					ObjectId:   resource.ID,
				},
				Relation: relationshipKind.String(),
				Subject:  subjectRef,
			},
		}
		return append(updates, relationshipUpdate)
	}
}
