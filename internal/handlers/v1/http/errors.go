package v1

import (
	"net/http"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
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
	default:
		return http.StatusInternalServerError
	}
}

// logErrorWithContext logs the error with structured context from ServiceError
func logErrorWithContext(message string, err error, additionalFields ...any) {
	fields := []any{}

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
		if serviceErr.RequestID != "" {
			fields = append(fields, "request_id", serviceErr.RequestID)
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
