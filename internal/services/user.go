package services

import (
	"context"
	"fmt"

	"github.com/Nerzal/gocloak/v13"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

const (
	photosClientName = "photos"
)

// UserService fetches users and their roles from Keycloak.
// Role bindings are retrieved from the client's role mappings, not realm roles.
type UserService struct {
	client       *gocloak.GoCloak
	clientID     string
	clientSecret string
	realm        string
}

func NewUserService(client *gocloak.GoCloak, clientID, clientSecret, realm string) *UserService {
	return &UserService{
		client:       client,
		clientID:     clientID,
		clientSecret: clientSecret,
		realm:        realm,
	}
}

func (u *UserService) GetUsers(ctx context.Context) ([]entity.User, error) {
	token, err := u.client.LoginClient(ctx, u.clientID, u.clientSecret, u.realm)
	if err != nil {
		return []entity.User{}, fmt.Errorf("faild to login: %v", err)
	}
	users, err := u.client.GetUsers(ctx, token.AccessToken, u.realm, gocloak.GetUsersParams{})
	if err != nil {
		return []entity.User{}, fmt.Errorf("failed to fetch users: %v", err)
	}

	myUsers := make([]entity.User, 0, len(users))
	for _, uu := range users {
		user := entity.User{
			Username: *uu.Username,
		}
		if uu.FirstName != nil {
			user.FirstName = *uu.FirstName
		}
		if uu.LastName != nil {
			user.LastName = *uu.LastName
		}
		roles, err := u.getUserRole(ctx, token.AccessToken, *uu.ID)
		if err != nil {
			return nil, err
		}
		if len(roles) > 0 {
			userRole, err := entity.RoleFromString(roles[0])
			if err != nil {
				return nil, err
			}
			user.Role = &userRole
		}
		myUsers = append(myUsers, user)
	}
	return myUsers, nil
}

func (u *UserService) getUserRole(ctx context.Context, accessToken string, userID string) ([]string, error) {
	roleMapping, err := u.client.GetRoleMappingByUserID(ctx, accessToken, u.realm, userID)
	if err != nil {
		return nil, err
	}
	roles := []string{}
	for _, m := range roleMapping.ClientMappings {
		if *m.Client == photosClientName {
			for _, rr := range *m.Mappings {
				roles = append(roles, *rr.Name)
			}
		}
	}

	return roles, nil
}
