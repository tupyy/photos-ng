package v1

import (
	"net/http"

	v1 "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ListAlbums handles GET /api/v1/albums requests to retrieve a list of albums.
// It supports pagination through limit and offset parameters.
// Returns HTTP 500 for server errors, or HTTP 200 with the album list on success.
func (s *ServerImpl) ListAlbums(c *gin.Context, params v1.ListAlbumsParams) {
	// Set default values for pagination
	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}
	offset := 0
	if params.Offset != nil {
		offset = *params.Offset
	}

	// Create album service and filter
	opts := services.NewAlbumOptionsWithOptions(services.WithLimit(limit), services.WithOffset(offset))

	albums, err := s.albumSrv.GetAlbums(c.Request.Context(), opts)
	if err != nil {
		zap.S().Errorw("failed to get albums", "error", err)
		c.JSON(http.StatusInternalServerError, v1.Error{
			Message: err.Error(),
		})
		return
	}

	// Convert entity albums to API albums
	apiAlbums := make([]v1.Album, 0, len(albums))
	for _, album := range albums {
		apiAlbums = append(apiAlbums, v1.NewAlbum(album))
	}

	// TODO: Get total count from service for proper pagination
	total := len(apiAlbums)

	response := v1.ListAlbumsResponse{
		Albums: apiAlbums,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	c.JSON(http.StatusOK, response)
}

// CreateAlbum handles POST /api/v1/albums requests to create a new album.
// It validates the request body and creates a new album in the database.
// Returns HTTP 400 for validation errors, HTTP 500 for server errors,
// or HTTP 201 with the created album on success.
func (s *ServerImpl) CreateAlbum(c *gin.Context) {
	var request v1.CreateAlbumRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, v1.Error{
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Create album service and create the album
	createdAlbum, err := s.albumSrv.CreateAlbum(c.Request.Context(), request)
	if err != nil {
		zap.S().Errorw("failed to create album", "error", err)
		c.JSON(http.StatusInternalServerError, v1.Error{
			Message: err.Error(),
		})
		return
	}

	zap.S().Infow("album created", "album_id", createdAlbum.ID, "name", request.Name)
	c.JSON(http.StatusCreated, v1.NewAlbum(*createdAlbum))
}

// GetAlbum handles GET /api/v1/albums/{id} requests to retrieve a specific album by ID.
// Returns HTTP 404 if the album is not found, HTTP 500 for server errors,
// or HTTP 200 with the album data on success.
func (s *ServerImpl) GetAlbum(c *gin.Context, id string) {
	// Create album service and get the album
	album, err := s.albumSrv.GetAlbum(c.Request.Context(), id)
	if err != nil {
		switch err.(type) {
		case *services.ErrResourceNotFound:
			c.JSON(http.StatusNotFound, v1.Error{
				Message: err.Error(),
			})
			return
		default:
			zap.S().Errorw("failed to get album", "error", err, "album_id", id)
			c.JSON(http.StatusInternalServerError, v1.Error{
				Message: err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, v1.NewAlbum(*album))
}

// UpdateAlbum handles PUT /api/v1/albums/{id} requests to update an album's metadata.
// Returns HTTP 400 for validation errors, HTTP 404 if album not found,
// HTTP 500 for server errors, or HTTP 200 with the updated album on success.
func (s *ServerImpl) UpdateAlbum(c *gin.Context, id string) {
	var request v1.UpdateAlbumRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, v1.Error{
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Create album service and update the album
	updatedAlbum, err := s.albumSrv.UpdateAlbum(c.Request.Context(), id, request)
	if err != nil {
		switch err.(type) {
		case *services.ErrResourceNotFound:
			c.JSON(http.StatusNotFound, v1.Error{
				Message: err.Error(),
			})
			return
		default:
			zap.S().Errorw("failed to update album", "error", err, "album_id", id)
			c.JSON(http.StatusInternalServerError, v1.Error{
				Message: err.Error(),
			})
			return
		}
	}

	zap.S().Infow("album updated", "album_id", id)
	c.JSON(http.StatusOK, v1.NewAlbum(*updatedAlbum))
}

// DeleteAlbum handles DELETE /api/v1/albums/{id} requests to delete an album.
// Returns HTTP 404 if album not found, HTTP 500 for server errors,
// or HTTP 204 on successful deletion.
func (s *ServerImpl) DeleteAlbum(c *gin.Context, id string) {
	// Create album service and delete the album
	err := s.albumSrv.DeleteAlbum(c.Request.Context(), id)
	if err != nil {
		switch err.(type) {
		case *services.ErrResourceNotFound:
			c.JSON(http.StatusNotFound, v1.Error{
				Message: err.Error(),
			})
			return
		default:
			zap.S().Errorw("failed to delete album", "error", err, "album_id", id)
			c.JSON(http.StatusInternalServerError, v1.Error{
				Message: err.Error(),
			})
			return
		}
	}

	zap.S().Infow("album deleted", "album_id", id)
	c.Status(http.StatusNoContent)
}

// SyncAlbum handles POST /api/v1/albums/{id}/sync requests to synchronize an album with the file system.
// Returns HTTP 404 if album not found, HTTP 500 for server errors,
// or HTTP 200 with sync results on success.
func (s *ServerImpl) SyncAlbum(c *gin.Context, id string) {
	syncedItems, err := s.albumSrv.SyncAlbum(c.Request.Context(), id)
	if err != nil {
		switch err.(type) {
		case *services.ErrResourceNotFound:
			c.JSON(http.StatusNotFound, v1.Error{
				Message: err.Error(),
			})
			return
		default:
			zap.S().Errorw("failed to sync album", "error", err, "album_id", id)
			c.JSON(http.StatusInternalServerError, v1.Error{
				Message: err.Error(),
			})
			return
		}
	}

	zap.S().Infow("album synced", "album_id", id, "synced_items", syncedItems)

	response := v1.SyncAlbumResponse{
		Message:     "Album sync completed",
		SyncedItems: syncedItems,
	}

	c.JSON(http.StatusOK, response)
}
