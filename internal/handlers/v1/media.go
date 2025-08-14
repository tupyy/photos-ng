package v1

import (
	"fmt"
	"net/http"

	v1 "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ListMedia handles GET /api/v1/media requests to retrieve a list of media items.
// It supports filtering by album, type, date range, and includes sorting and pagination.
// Returns HTTP 500 for server errors, or HTTP 200 with the media list on success.
func (s *ServerImpl) ListMedia(c *gin.Context, params v1.ListMediaParams) {
	// Build media opt from parameters
	opt := &services.MediaOptions{}

	// Set default values for pagination
	limit := 20
	if params.Limit != nil {
		opt.MediaLimit = *params.Limit
	}
	offset := 0
	if params.Offset != nil {
		opt.MediaOffset = *params.Offset
	}

	// Add album filter
	if params.AlbumId != nil {
		opt.AlbumID = params.AlbumId
	}

	// Add media type filter
	if params.Type != nil {
		mediaType := string(*params.Type)
		opt.MediaType = &mediaType
	}

	// Add date range filters
	if params.StartDate != nil {
		opt.StartDate = &params.StartDate.Time
	}
	if params.EndDate != nil {
		opt.EndDate = &params.EndDate.Time
	}

	// Add sorting
	if params.SortBy != nil {
		switch string(*params.SortBy) {
		case "capturedAt":
			opt.SortBy = services.SortByCapturedAt
		default:
			opt.SortBy = string(*params.SortBy)
		}
		if params.SortOrder != nil {
			opt.Descending = *params.SortOrder == v1.Desc
		}
	}

	// Create media service and get media
	mediaItems, err := s.mediaSrv.GetMedia(c.Request.Context(), opt)
	if err != nil {
		zap.S().Errorw("failed to get media", "error", err)
		c.JSON(http.StatusInternalServerError, v1.Error{
			Message: err.Error(),
		})
		return
	}

	// Convert entity media to API media
	apiMedia := make([]v1.Media, 0, len(mediaItems))
	for _, media := range mediaItems {
		apiMedia = append(apiMedia, v1.NewMedia(media))
	}

	// TODO: Get total count from service for proper pagination
	total := len(apiMedia)

	response := v1.ListMediaResponse{
		Media:  apiMedia,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	c.JSON(http.StatusOK, response)
}

// GetMedia handles GET /api/v1/media/{id} requests to retrieve a specific media item by ID.
// Returns HTTP 404 if the media is not found, HTTP 500 for server errors,
// or HTTP 200 with the media data on success.
func (s *ServerImpl) GetMedia(c *gin.Context, id string) {
	// Create media service and get the media
	media, err := s.mediaSrv.GetMediaByID(c.Request.Context(), id)
	if err != nil {
		switch err.(type) {
		case *services.ErrResourceNotFound:
			c.JSON(http.StatusNotFound, v1.Error{
				Message: err.Error(),
			})
			return
		default:
			zap.S().Errorw("failed to get media", "error", err, "media_id", id)
			c.JSON(http.StatusInternalServerError, v1.Error{
				Message: err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, v1.NewMedia(*media))
}

// UpdateMedia handles PUT /api/v1/media/{id} requests to update media metadata.
// Returns HTTP 400 for validation errors, HTTP 404 if media not found,
// HTTP 500 for server errors, or HTTP 200 with the updated media on success.
func (s *ServerImpl) UpdateMedia(c *gin.Context, id string) {
	var request v1.UpdateMediaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, v1.Error{
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	// Get the existing media and apply updates
	media, err := s.mediaSrv.GetMediaByID(c.Request.Context(), id)
	if err != nil {
		switch err.(type) {
		case *services.ErrResourceNotFound:
			c.JSON(http.StatusNotFound, v1.Error{
				Message: err.Error(),
			})
			return
		default:
			zap.S().Errorw("failed to get media for update", "error", err, "media_id", id)
			c.JSON(http.StatusInternalServerError, v1.Error{
				Message: err.Error(),
			})
			return
		}
	}

	// Apply updates from request to entity
	request.ApplyTo(media)

	// Update the media
	updatedMedia, err := s.mediaSrv.UpdateMedia(c.Request.Context(), *media)
	if err != nil {
		switch err.(type) {
		case *services.ErrResourceNotFound:
			c.JSON(http.StatusNotFound, v1.Error{
				Message: err.Error(),
			})
			return
		default:
			zap.S().Errorw("failed to update media", "error", err, "media_id", id)
			c.JSON(http.StatusInternalServerError, v1.Error{
				Message: err.Error(),
			})
			return
		}
	}

	zap.S().Infow("media updated", "media_id", id)
	c.JSON(http.StatusOK, v1.NewMedia(*updatedMedia))
}

// DeleteMedia handles DELETE /api/v1/media/{id} requests to delete a media item.
// Returns HTTP 404 if media not found, HTTP 500 for server errors,
// or HTTP 204 on successful deletion.
func (s *ServerImpl) DeleteMedia(c *gin.Context, id string) {
	// Create media service and delete the media
	err := s.mediaSrv.DeleteMedia(c.Request.Context(), id)
	if err != nil {
		switch err.(type) {
		case *services.ErrResourceNotFound:
			c.JSON(http.StatusNotFound, v1.Error{
				Message: err.Error(),
			})
			return
		default:
			zap.S().Errorw("failed to delete media", "error", err, "media_id", id)
			c.JSON(http.StatusInternalServerError, v1.Error{
				Message: err.Error(),
			})
			return
		}
	}

	zap.S().Infow("media deleted", "media_id", id)
	c.Status(http.StatusNoContent)
}

// GetMediaContent handles GET /api/v1/media/{id}/content requests to serve the full media content.
// Returns HTTP 404 if media not found, HTTP 500 for server errors,
// or the binary media content with appropriate content-type on success.
func (s *ServerImpl) GetMediaContent(c *gin.Context, id string) {
	// Get the media from service
	media, err := s.mediaSrv.GetMediaByID(c.Request.Context(), id)
	if err != nil {
		switch err.(type) {
		case *services.ErrResourceNotFound:
			c.JSON(http.StatusNotFound, v1.Error{
				Message: "Media not found: " + id,
			})
			return
		default:
			zap.S().Errorw("failed to get media", "error", err, "media_id", id)
			c.JSON(http.StatusInternalServerError, v1.Error{
				Message: err.Error(),
			})
			return
		}
	}

	// Check if content function is available
	if media.Content == nil {
		zap.S().Errorw("media content function not available", "media_id", id)
		c.JSON(http.StatusInternalServerError, v1.Error{
			Message: "Media content not available",
		})
		return
	}

	// Read the media content from filesystem
	contentReader, err := media.Content()
	if err != nil {
		zap.S().Errorw("failed to read media content", "error", err, "media_id", id, "filepath", media.Filepath())
		c.JSON(http.StatusInternalServerError, v1.Error{
			Message: "Failed to read media content: " + err.Error(),
		})
		return
	}

	// Set appropriate headers for content
	c.Header("Content-Type", media.ContentType())
	c.Header("Cache-Control", "public, max-age=86400") // Cache for 24 hours

	// Stream the content to the client
	c.DataFromReader(http.StatusOK, -1, media.ContentType(), contentReader, nil)
}

// GetMediaThumbnail handles GET /api/v1/media/{id}/thumbnail requests to serve media thumbnails.
// Returns HTTP 404 if media not found, HTTP 500 for server errors,
// or the binary thumbnail data on success.
func (s *ServerImpl) GetMediaThumbnail(c *gin.Context, id string) {
	// Get the media from service
	media, err := s.mediaSrv.GetMediaByID(c.Request.Context(), id)
	if err != nil {
		switch err.(type) {
		case *services.ErrResourceNotFound:
			c.JSON(http.StatusNotFound, v1.Error{
				Message: "Media not found: " + id,
			})
			return
		default:
			zap.S().Errorw("failed to get media", "error", err, "media_id", id)
			c.JSON(http.StatusInternalServerError, v1.Error{
				Message: err.Error(),
			})
			return
		}
	}

	// Check if media has thumbnail
	if !media.HasThumbnail() {
		c.JSON(http.StatusNotFound, v1.Error{
			Message: "Thumbnail not available for media: " + id,
		})
		return
	}

	// Set appropriate headers for thumbnail
	c.Header("Content-Type", "image/jpeg")             // Thumbnails are typically JPEG
	c.Header("Cache-Control", "public, max-age=86400") // Cache for 24 hours
	c.Header("Content-Length", fmt.Sprintf("%d", len(media.Thumbnail)))

	// Serve the thumbnail data
	c.Data(http.StatusOK, "image/jpeg", media.Thumbnail)
}

// UploadMedia handles POST /api/v1/media requests to upload a new media file.
// Returns HTTP 400 for validation errors, HTTP 404 if album not found,
// HTTP 500 for server errors, or HTTP 201 with the created media on success.
func (s *ServerImpl) UploadMedia(c *gin.Context) {
	// Parse multipart form
	err := c.Request.ParseMultipartForm(32 << 20) // 32 MB max
	if err != nil {
		zap.S().Errorw("failed to parse multipart form", "error", err)
		c.JSON(http.StatusBadRequest, v1.Error{
			Message: "Failed to parse multipart form: " + err.Error(),
		})
		return
	}

	// Get form values
	filename := c.Request.FormValue("filename")
	albumId := c.Request.FormValue("albumId")

	if filename == "" {
		c.JSON(http.StatusBadRequest, v1.Error{
			Message: "filename is required",
		})
		return
	}

	if albumId == "" {
		c.JSON(http.StatusBadRequest, v1.Error{
			Message: "albumId is required",
		})
		return
	}

	// Get the uploaded file
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		zap.S().Errorw("failed to get uploaded file", "error", err)
		c.JSON(http.StatusBadRequest, v1.Error{
			Message: "Failed to get uploaded file: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// Get the album to ensure it exists
	album, err := s.albumSrv.GetAlbum(c.Request.Context(), albumId)
	if err != nil {
		switch err.(type) {
		case *services.ErrResourceNotFound:
			c.JSON(http.StatusNotFound, v1.Error{
				Message: "Album not found: " + albumId,
			})
			return
		default:
			zap.S().Errorw("failed to get album", "error", err, "album_id", albumId)
			c.JSON(http.StatusInternalServerError, v1.Error{
				Message: err.Error(),
			})
			return
		}
	}

	// Convert request data to entity using the mapping function
	media := v1.ToMediaEntity(filename, albumId, file, *album)

	// Write the media using the service
	createdMedia, err := s.mediaSrv.WriteMedia(c.Request.Context(), media)
	if err != nil {
		zap.S().Errorw("failed to write media", "error", err, "filename", filename, "album_id", albumId)
		c.JSON(http.StatusInternalServerError, v1.Error{
			Message: err.Error(),
		})
		return
	}

	zap.S().Infow("media uploaded successfully", "media_id", createdMedia.ID, "filename", filename, "album_id", albumId)
	c.JSON(http.StatusCreated, v1.NewMedia(*createdMedia))
}
