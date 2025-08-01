package v1

import (
	"net/http"

	v1 "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetTimeline handles GET /api/v1/timeline requests to retrieve a timeline of media organized in buckets.
// It supports pagination and returns media organized by date buckets.
// Returns HTTP 400 for invalid parameters, HTTP 500 for server errors,
// or HTTP 200 with the timeline data on success.
func (s *ServerImpl) GetTimeline(c *gin.Context, params v1.GetTimelineParams) {
	// Set default values for pagination
	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}
	offset := 0
	if params.Offset != nil {
		offset = *params.Offset
	}

	// Create timeline service and filter
	filter := &services.TimelineFilter{
		StartDate: params.StartDate.Time,
		Limit:     limit,
		Offset:    offset,
	}

	// Get timeline from service
	buckets, years, err := s.timelineSrv.GetTimeline(c.Request.Context(), filter)
	if err != nil {
		zap.S().Errorw("failed to get timeline", "error", err)
		c.JSON(http.StatusInternalServerError, v1.Error{
			Message: err.Error(),
		})
		return
	}

	// Convert service buckets to API buckets
	apiBuckets := make([]v1.Bucket, 0, len(buckets))
	for _, bucket := range buckets {
		// Convert entity media to href strings
		mediaHrefs := make([]string, 0, len(bucket.Media))
		for _, media := range bucket.Media {
			mediaHrefs = append(mediaHrefs, "/api/v1/media/"+media.ID)
		}

		apiBucket := v1.Bucket{
			Year:  &bucket.Year,
			Month: &bucket.Month,
			Media: &mediaHrefs,
		}
		apiBuckets = append(apiBuckets, apiBucket)
	}

	response := v1.GetTimelineResponse{
		Buckets: apiBuckets,
		Years:   years,
		Total:   len(buckets), // Total before pagination was applied
		Limit:   limit,
		Offset:  offset,
	}

	zap.S().Infow("timeline retrieved", "start_date", params.StartDate.Time, "total_buckets", len(buckets), "years_count", len(years))
	c.JSON(http.StatusOK, response)
}
