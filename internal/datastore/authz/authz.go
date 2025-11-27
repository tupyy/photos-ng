package authz

import (
	"context"
	"io"

	"github.com/authzed/authzed-go/v1"

	v1pb "github.com/authzed/authzed-go/proto/authzed/api/v1"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/logger"
)

type RelationshipFn func(updates []*v1pb.RelationshipUpdate) []*v1pb.RelationshipUpdate

//go:generate go run github.com/ecordell/optgen -output zz_generated.options.go . ListOptions
type ListOptions struct {
	Limit  uint32  `debugmap:"visible"`
	Cursor *string `debugmap:"visible"`
}

type Datastore struct {
	client *authzed.Client
}

func NewAuthzDatastore(client *authzed.Client) *Datastore {
	return &Datastore{
		client: client,
	}
}

func (d *Datastore) ListResources(ctx context.Context, token, userID string, permission entity.Permission, resource entity.Resource) ([]string, error) {
	logger := logger.New("authz_store").
		WithContext(ctx).
		Operation("list_resources").
		Build()

	// Lookup resources for which the user has the specified permission
	req := &v1pb.LookupResourcesRequest{
		ResourceObjectType: resource.Kind.String(),
		Permission:         permission.String(),
		Subject: &v1pb.SubjectReference{
			Object: &v1pb.ObjectReference{
				ObjectType: "user",
				ObjectId:   userID,
			},
		},
	}

	logger.Step("list_resources").
		WithString("user_id", userID).
		WithString("permission", permission.String()).
		WithString("resource_kind", resource.Kind.String()).
		Log()

	// Use token for at least as fresh consistency
	req.Consistency = &v1pb.Consistency{
		Requirement: &v1pb.Consistency_AtLeastAsFresh{
			AtLeastAsFresh: &v1pb.ZedToken{Token: token},
		},
	}

	resp, err := d.client.LookupResources(ctx, req)
	if err != nil {
		return nil, err
	}

	ids := []string{}
	for {
		resource, err := resp.Recv()
		if err != nil {
			// Check if we've reached the end of the stream
			if err == io.EOF {
				break
			}
			return nil, err
		}

		ids = append(ids, resource.ResourceObjectId)
	}

	logger.Debug("list_resources").WithInt("count", len(ids))

	return ids, nil
}

func (d *Datastore) WriteRelationships(ctx context.Context, relationships ...RelationshipFn) (string, error) {
	logger := logger.New("authz_store").
		WithContext(ctx).
		Operation("write_relationships").
		Build()

	logger.Step("execute_relationship_writes").Log()
	relationshipsUpdate := []*v1pb.RelationshipUpdate{}

	if len(relationships) == 0 {
		return "", nil
	}

	for _, fn := range relationships {
		relationshipsUpdate = fn(relationshipsUpdate)
	}

	resp, err := d.client.WriteRelationships(ctx, &v1pb.WriteRelationshipsRequest{
		Updates: relationshipsUpdate,
	})
	if err != nil {
		return "", err
	}

	logger.Step("relationships_wrote").WithString("token", resp.WrittenAt.Token).Log()

	return resp.WrittenAt.Token, nil
}

// DeleteRelationships removes all the relationships of the resource.
// If only one relationships needs to be remove use WriteRelationships and WithoutRelations generator.
func (d *Datastore) DeleteRelationships(ctx context.Context, resource entity.Resource) (string, error) {
	logger := logger.New("authz_store").
		WithContext(ctx).
		Operation("delete_relationships").
		Build()

	logger.Step("execute_relationship_deletions").Log()

	req := &v1pb.DeleteRelationshipsRequest{
		RelationshipFilter: &v1pb.RelationshipFilter{
			ResourceType: resource.Kind.String(),
		},
	}
	if resource.ID != "" {
		req.RelationshipFilter.OptionalResourceId = resource.ID
	}

	resp, err := d.client.DeleteRelationships(ctx, req)
	if err != nil {
		return "", err
	}

	logger.Step("write_zed_token").WithString("token", resp.DeletedAt.Token).Log()

	return resp.DeletedAt.Token, nil
}

func (d *Datastore) GetPermissions(ctx context.Context, token string, userID string, resources []entity.Resource) (map[entity.Resource][]entity.Permission, error) {
	logger := logger.New("authz_store").
		WithContext(ctx).
		Operation("get_permissions").
		Build()

	// Build bulk permission check request for all assessment-permission combinations
	var items []*v1pb.CheckBulkPermissionsRequestItem

	resourceIndex := []struct {
		resource   entity.Resource
		permission entity.Permission
	}{}

	for _, r := range resources {
		for _, perm := range entity.AllPermissions {
			items = append(items, &v1pb.CheckBulkPermissionsRequestItem{
				Resource: &v1pb.ObjectReference{
					ObjectType: r.Kind.String(),
					ObjectId:   r.ID,
				},
				Permission: perm.String(),
				Subject: &v1pb.SubjectReference{
					Object: &v1pb.ObjectReference{
						ObjectType: "user",
						ObjectId:   userID,
					},
				},
			})
			resourceIndex = append(resourceIndex, struct {
				resource   entity.Resource
				permission entity.Permission
			}{r, perm})
		}
	}

	req := &v1pb.CheckBulkPermissionsRequest{Items: items}
	req.Consistency = &v1pb.Consistency{
		Requirement: &v1pb.Consistency_AtLeastAsFresh{
			AtLeastAsFresh: &v1pb.ZedToken{Token: token},
		},
	}

	resp, err := d.client.CheckBulkPermissions(ctx, req)
	if err != nil {
		return nil, err
	}

	logger.Step("check_bulk_permissions").Log()

	// Build result map
	result := make(map[entity.Resource][]entity.Permission)
	for i, pair := range resp.Pairs {
		resource := resourceIndex[i].resource
		permission := resourceIndex[i].permission

		// Initialize slice if not exists
		if _, exists := result[resource]; !exists {
			result[resource] = []entity.Permission{}
		}

		// Check if the response is an item (not an error)
		if item := pair.GetItem(); item != nil {
			if item.Permissionship == v1pb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION {
				result[resource] = append(result[resource], permission)
			}
		}
		// If there's an error for this permission check, we skip it
	}

	return result, nil
}
