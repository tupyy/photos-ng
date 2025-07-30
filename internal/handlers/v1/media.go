package v1

import (
	"net/http"

	v1 "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
	dtContext "git.tls.tupangiu.ro/cosmin/photos-ng/pkg/context"
	"github.com/gin-gonic/gin"
	"github.com/oapi-codegen/runtime/types"
	"go.uber.org/zap"
)

// ListMedia handles GET /api/v1/media requests to retrieve a list of media items.
// It supports filtering by album, type, date range, and includes sorting and pagination.
// Returns HTTP 500 for server errors, or HTTP 200 with the media list on success.
func (s *ServerImpl) ListMedia(c *gin.Context, params v1.ListMediaParams) {
	dt := dtContext.MustFromContext(c)

	// Set default values for pagination
	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}
	offset := 0
	if params.Offset != nil {
		offset = *params.Offset
	}

	// Build media filter from parameters
	filter := &services.MediaFilter{
		Limit:  limit,
		Offset: offset,
	}

	// Add album filter
	if params.AlbumId != nil {
		filter.AlbumID = params.AlbumId
	}

	// Add media type filter
	if params.Type != nil {
		mediaType := string(*params.Type)
		filter.MediaType = &mediaType
	}

	// Add date range filters
	if params.StartDate != nil {
		filter.StartDate = &params.StartDate.Time
	}
	if params.EndDate != nil {
		filter.EndDate = &params.EndDate.Time
	}

	// Add sorting
	if params.SortBy != nil {
		filter.SortBy = string(*params.SortBy)
		if params.SortOrder != nil {
			filter.Descending = *params.SortOrder == v1.Desc
		}
	}

	// Create media service and get media
	mediaSrv := services.NewMediaService(dt)
	mediaItems, err := mediaSrv.GetMedia(c.Request.Context(), filter)
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
		apiMedia = append(apiMedia, v1.NewMediaFromEntity(media))
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
	dt := dtContext.MustFromContext(c)

	// Create media service and get the media
	mediaSrv := services.NewMediaService(dt)
	media, err := mediaSrv.GetMediaByID(c.Request.Context(), id)
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

	c.JSON(http.StatusOK, v1.NewMediaFromEntity(*media))
}

// UpdateMedia handles PUT /api/v1/media/{id} requests to update media metadata.
// Returns HTTP 400 for validation errors, HTTP 404 if media not found,
// HTTP 500 for server errors, or HTTP 200 with the updated media on success.
func (s *ServerImpl) UpdateMedia(c *gin.Context, id types.UUID) {
	dt := dtContext.MustFromContext(c)

	var request v1.UpdateMediaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, v1.Error{
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	idStr := id.String()

	// Create media service and update the media
	mediaSrv := services.NewMediaService(dt)
	updatedMedia, err := mediaSrv.UpdateMedia(c.Request.Context(), idStr, func(media *entity.Media) error {
		// Update captured at time
		if request.CapturedAt != nil {
			media.CapturedAt = request.CapturedAt.Time
		}

		// Update EXIF data
		if request.Exif != nil {
			for _, exif := range *request.Exif {
				media.ExifMetadata[exif.Key] = exif.Value
			}
		}

		return nil
	})
	if err != nil {
		switch err.(type) {
		case *services.ErrResourceNotFound:
			c.JSON(http.StatusNotFound, v1.Error{
				Message: err.Error(),
			})
			return
		default:
			zap.S().Errorw("failed to update media", "error", err, "media_id", idStr)
			c.JSON(http.StatusInternalServerError, v1.Error{
				Message: err.Error(),
			})
			return
		}
	}

	zap.S().Infow("media updated", "media_id", idStr)
	c.JSON(http.StatusOK, v1.NewMediaFromEntity(*updatedMedia))
}

// DeleteMedia handles DELETE /api/v1/media/{id} requests to delete a media item.
// Returns HTTP 404 if media not found, HTTP 500 for server errors,
// or HTTP 204 on successful deletion.
func (s *ServerImpl) DeleteMedia(c *gin.Context, id string) {
	dt := dtContext.MustFromContext(c)

	// Create media service and delete the media
	mediaSrv := services.NewMediaService(dt)
	err := mediaSrv.DeleteMedia(c.Request.Context(), id)
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
func (s *ServerImpl) GetMediaContent(c *gin.Context, id types.UUID) {
	dt := dtContext.MustFromContext(c)

	idStr := id.String()

	// Create media service and get media content
	mediaSrv := services.NewMediaService(dt)
	media, content, err := mediaSrv.GetMediaContent(c.Request.Context(), idStr)
	if err != nil {
		switch err.(type) {
		case *services.ErrResourceNotFound:
			c.JSON(http.StatusNotFound, v1.Error{
				Message: err.Error(),
			})
			return
		default:
			zap.S().Errorw("failed to get media content", "error", err, "media_id", idStr)
			c.JSON(http.StatusInternalServerError, v1.Error{
				Message: err.Error(),
			})
			return
		}
	}

	// Set appropriate headers
	c.Header("Content-Type", media.ContentType())
	c.Header("Content-Disposition", "inline; filename=\""+media.Filename+"\"")

	zap.S().Infow("serving media content", "media_id", idStr, "filename", media.Filename)
	c.Data(http.StatusOK, media.ContentType(), content)
}

// GetMediaThumbnail handles GET /api/v1/media/{id}/thumbnail requests to serve media thumbnails.
// Returns HTTP 404 if media not found, HTTP 500 for server errors,
// or the binary thumbnail data on success.
func (s *ServerImpl) GetMediaThumbnail(c *gin.Context, id string) {
	dt := dtContext.MustFromContext(c)

	// Create media service and get media thumbnail
	mediaSrv := services.NewMediaService(dt)
	media, thumbnail, err := mediaSrv.GetMediaThumbnail(c.Request.Context(), id)
	if err != nil {
		switch err.(type) {
		case *services.ErrResourceNotFound:
			c.JSON(http.StatusNotFound, v1.Error{
				Message: err.Error(),
			})
			return
		default:
			zap.S().Errorw("failed to get media thumbnail", "error", err, "media_id", id)
			c.JSON(http.StatusInternalServerError, v1.Error{
				Message: err.Error(),
			})
			return
		}
	}

	_ = media // Use media if needed for additional logic

	// Serve thumbnail
	c.Header("Content-Type", "image/jpeg")            // Assume JPEG thumbnails
	c.Header("Cache-Control", "public, max-age=3600") // Cache for 1 hour

	zap.S().Infow("serving media thumbnail", "media_id", id)
	c.Data(http.StatusOK, "image/jpeg", thumbnail)
}
