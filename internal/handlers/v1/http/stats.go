package v1

import (
	"net/http"

	v1 "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1/http"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/requestid"
	"github.com/gin-gonic/gin"
)

// GetStats handles GET /api/v1/stats requests to retrieve application statistics.
// Returns HTTP 500 for server errors, or HTTP 200 with the stats on success.
func (s *Handler) GetStats(c *gin.Context) {
	// Get stats from the service
	stats, err := s.statsSrv.GetStats(c.Request.Context())
	if err != nil {
		logError(requestid.FromGin(c), "GetStats", err)
		c.JSON(getHTTPStatusFromError(err), errorResponse(c, err.Error()))
		return
	}

	// Create response structure using the generated type
	response := v1.StatsResponse{
		Years:      stats.Years,
		CountMedia: stats.CountMedia,
		CountAlbum: stats.CountAlbum,
	}

	c.JSON(http.StatusOK, response)
}
