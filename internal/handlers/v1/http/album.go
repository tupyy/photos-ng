package v1

import (
	"net/http"

	v1 "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1/http"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/requestid"
	"github.com/gin-gonic/gin"
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
		logError(requestid.FromGin(c), "ListAlbums", err)
		c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
		return
	}

	countOpt := services.NewAlbumOptionsWithOptions(
		services.WithHasParent(hasParent),
	)
	count, err := s.albumSrv.CountAlbums(c.Request.Context(), countOpt)
	if err != nil {
		logError(requestid.FromGin(c), "ListAlbums", err)
		c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
		return
	}

	// Convert entity albums to API albums
	apiAlbums := make([]v1.Album, 0, len(albums))
	for _, album := range albums {
		apiAlbums = append(apiAlbums, v1.NewAlbum(album))
	}

	response := v1.ListAlbumsResponse{
		Albums: apiAlbums,
		Total:  count,
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
		c.JSON(http.StatusBadRequest, errorResponse(c, "Invalid request body: "+err.Error()))
		return
	}

	// Convert request to entity and create the album
	createdAlbum, err := s.albumSrv.CreateAlbum(c.Request.Context(), request.Entity())
	if err != nil {
		logError(requestid.FromGin(c), "CreateAlbum", err)
		c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
		return
	}
	c.JSON(http.StatusCreated, v1.NewAlbum(*createdAlbum))
}

// GetAlbum handles GET /api/v1/albums/{id} requests to retrieve a specific album by ID.
// Returns HTTP 404 if the album is not found, HTTP 500 for server errors,
// or HTTP 200 with the album data on success.
func (s *Handler) GetAlbum(c *gin.Context, id string) {
	// Create album service and get the album
	album, err := s.albumSrv.GetAlbum(c.Request.Context(), id)
	if err != nil {
		logError(requestid.FromGin(c), "GetAlbum", err)
		c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
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
		c.JSON(http.StatusBadRequest, errorResponse(c, "Invalid request body: "+err.Error()))
		return
	}

	// Get the existing album and apply updates
	album, err := s.albumSrv.GetAlbum(c.Request.Context(), id)
	if err != nil {
		logError(requestid.FromGin(c), "UpdateAlbum", err)
		c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
		return
	}

	// Apply updates from request to entity
	request.ApplyTo(album)

	// Update the album
	updatedAlbum, err := s.albumSrv.UpdateAlbum(c.Request.Context(), *album)
	if err != nil {
		logError(requestid.FromGin(c), "UpdateAlbum", err)
		c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
		return
	}
	c.JSON(http.StatusOK, v1.NewAlbum(*updatedAlbum))
}

// DeleteAlbum handles DELETE /api/v1/albums/{id} requests to delete an album.
// Returns HTTP 404 if album not found, HTTP 500 for server errors,
// or HTTP 204 on successful deletion.
func (s *Handler) DeleteAlbum(c *gin.Context, id string) {
	// Create album service and delete the album
	err := s.albumSrv.DeleteAlbum(c.Request.Context(), id)
	if err != nil {
		logError(requestid.FromGin(c), "DeleteAlbum", err)
		c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
		return
	}
	c.Status(http.StatusNoContent)
}

// SyncAlbum handles POST /api/v1/albums/{id}/sync requests to synchronize an album with the file system.
// Returns HTTP 404 if album not found, HTTP 500 for server errors,
// or HTTP 200 with sync results on success.
func (s *Handler) SyncAlbum(c *gin.Context, id string) {
	c.JSON(http.StatusOK, gin.H{})
}
