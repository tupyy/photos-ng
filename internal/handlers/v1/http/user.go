package v1

import (
	"net/http"

	v1 "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1/http"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/context/user"
	"github.com/gin-gonic/gin"
)

// GetCurrentUser handles GET /api/v1/user requests to retrieve the current user profile.
func (h *Handler) GetCurrentUser(c *gin.Context) {
	u := user.FromGin(c)
	if u == nil {
		c.JSON(http.StatusUnauthorized, errorResponse(c, "user not found in context"))
		return
	}

	// Determine permissions based on user role
	canSync := v1.PermissionsCanSyncDenied
	canCreateAlbums := v1.PermissionsCanCreateAlbumsDenied

	// TODO: get the actual permissions from authz service
	if u.Role != nil {
		switch *u.Role {
		case entity.AdminRole:
			canSync = v1.PermissionsCanSyncAllowed
			canCreateAlbums = v1.PermissionsCanCreateAlbumsAllowed
		case entity.CreatorRole:
			canCreateAlbums = v1.PermissionsCanCreateAlbumsAllowed
		}
	}

	response := v1.User{
		User: u.Username,
		Name: u.FirstName + " " + u.LastName,
		Permissions: v1.Permissions{
			CanSync:         &canSync,
			CanCreateAlbums: &canCreateAlbums,
		},
	}

	c.JSON(http.StatusOK, response)
}
