package v1

import (
	"net/http"

	v1 "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1/http"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/context/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// getHTTPStatusFromError maps ServiceError types to appropriate HTTP status codes
func getHTTPStatusFromError(err error) int {
	switch err.(type) {
	case *services.NotFoundError:
		return http.StatusNotFound
	case *services.ConflictError:
		return http.StatusConflict
	case *services.ValidationError:
		return http.StatusBadRequest
	case *services.InternalError:
		return http.StatusInternalServerError
	case *services.ForbiddenAccessError:
		return http.StatusForbidden
	default:
		return http.StatusInternalServerError
	}
}

// errorResponse creates a standardized error response with requestId included in details
func errorResponse(c *gin.Context, message string) v1.Error {
	requestID := requestid.FromGin(c)

	details := map[string]any{
		"request_id": requestID,
	}

	return v1.Error{
		Message: message,
		Details: &details,
	}
}

// logError logs the error with structured context from ServiceError and requestId
func logError(requestID, message string, err error, additionalFields ...any) {
	fields := []any{}

	// Always add requestId
	if requestID != "" {
		fields = append(fields, "request_id", requestID)
	}

	// Extract structured context from ServiceError
	if serviceErr, ok := err.(*services.ServiceError); ok {
		if serviceErr.Operation != "" {
			fields = append(fields, "operation", serviceErr.Operation)
		}
		if serviceErr.Step != "" {
			fields = append(fields, "step", serviceErr.Step)
		}
		if serviceErr.Condition != "" {
			fields = append(fields, "condition", serviceErr.Condition)
		}

		// Add context fields
		for key, value := range serviceErr.Context {
			switch key {
			case "album_id", "media_id", "filename", "album_path", "filepath", "parent_id", "job_id":
				fields = append(fields, key, value)
			}
		}
	}

	// Add any additional fields provided by caller
	fields = append(fields, additionalFields...)

	zap.S().Errorw(message, fields...)
}
