package v1

import (
	"net/http"

	v1 "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1/http"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ListAlbums handles GET /api/v1/albums requests to retrieve a list of albums.
// It supports pagination through limit and offset parameters.
// Returns HTTP 500 for server errors, or HTTP 200 with the album list on success.
func (s *Handler) ListAlbums(c *gin.Context, params v1.ListAlbumsParams) {
	// Set default values for pagination
	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}
	offset := 0
	if params.Offset != nil {
		offset = *params.Offset
	}

	hasParent := false
	if params.WithParent != nil {
		hasParent = *params.WithParent
	}

	// Create album service and filter
	opts := services.NewAlbumOptionsWithOptions(
		services.WithLimit(limit),
		services.WithOffset(offset),
		services.WithHasParent(hasParent),
	)

	albums, err := s.albumSrv.GetAlbums(c.Request.Context(), opts)
	if err != nil {
		logErrorWithContext("failed to get albums", err)
		c.JSON(getHTTPStatusFromError(err), v1.Error{
			Message: err.Error(),
		})
		return
	}

	// Get total count for pagination (without limit/offset)
	totalOpts := services.NewAlbumOptionsWithOptions(
		services.WithHasParent(hasParent),
		// No limit/offset for total count
	)
	allAlbums, err := s.albumSrv.GetAlbums(c.Request.Context(), totalOpts)
	if err != nil {
		logErrorWithContext("failed to get total albums count", err)
		c.JSON(getHTTPStatusFromError(err), v1.Error{
			Message: err.Error(),
		})
		return
	}
	total := len(allAlbums)

	// Convert entity albums to API albums
	apiAlbums := make([]v1.Album, 0, len(albums))
	for _, album := range albums {
		apiAlbums = append(apiAlbums, v1.NewAlbum(album))
	}

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
func (s *Handler) CreateAlbum(c *gin.Context) {
	var request v1.CreateAlbumRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, v1.Error{
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Convert request to entity and create the album
	createdAlbum, err := s.albumSrv.CreateAlbum(c.Request.Context(), request.Entity())
	if err != nil {
		logErrorWithContext("failed to create album", err)
		c.JSON(getHTTPStatusFromError(err), v1.Error{
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
func (s *Handler) GetAlbum(c *gin.Context, id string) {
	// Create album service and get the album
	album, err := s.albumSrv.GetAlbum(c.Request.Context(), id)
	if err != nil {
		logErrorWithContext("failed to get album", err)
		c.JSON(getHTTPStatusFromError(err), v1.Error{
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, v1.NewAlbum(*album))
}

// UpdateAlbum handles PUT /api/v1/albums/{id} requests to update an album's metadata.
// Returns HTTP 400 for validation errors, HTTP 404 if album not found,
// HTTP 500 for server errors, or HTTP 200 with the updated album on success.
func (s *Handler) UpdateAlbum(c *gin.Context, id string) {
	var request v1.UpdateAlbumRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, v1.Error{
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Get the existing album and apply updates
	album, err := s.albumSrv.GetAlbum(c.Request.Context(), id)
	if err != nil {
		logErrorWithContext("failed to get album for update", err)
		c.JSON(getHTTPStatusFromError(err), v1.Error{
			Message: err.Error(),
		})
		return
	}

	// Apply updates from request to entity
	request.ApplyTo(album)

	// Update the album
	updatedAlbum, err := s.albumSrv.UpdateAlbum(c.Request.Context(), *album)
	if err != nil {
		logErrorWithContext("failed to update album", err)
		c.JSON(getHTTPStatusFromError(err), v1.Error{
			Message: err.Error(),
		})
		return
	}

	zap.S().Infow("album updated", "album_id", id)
	c.JSON(http.StatusOK, v1.NewAlbum(*updatedAlbum))
}

// DeleteAlbum handles DELETE /api/v1/albums/{id} requests to delete an album.
// Returns HTTP 404 if album not found, HTTP 500 for server errors,
// or HTTP 204 on successful deletion.
func (s *Handler) DeleteAlbum(c *gin.Context, id string) {
	// Create album service and delete the album
	err := s.albumSrv.DeleteAlbum(c.Request.Context(), id)
	if err != nil {
		logErrorWithContext("failed to delete album", err)
		c.JSON(getHTTPStatusFromError(err), v1.Error{
			Message: err.Error(),
		})
		return
	}

	zap.S().Infow("album deleted", "album_id", id)
	c.Status(http.StatusNoContent)
}

// SyncAlbum handles POST /api/v1/albums/{id}/sync requests to synchronize an album with the file system.
// Returns HTTP 404 if album not found, HTTP 500 for server errors,
// or HTTP 200 with sync results on success.
func (s *Handler) SyncAlbum(c *gin.Context, id string) {
	c.JSON(http.StatusOK, gin.H{})
}
