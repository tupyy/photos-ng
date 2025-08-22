package services

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"time"
)

// ServiceError provides structured error context for business operations
type ServiceError struct {
	// Core identification
	Operation string    `json:"operation"`  // "create_album", "write_media", etc.
	Step      string    `json:"step"`       // "database_write", "filesystem_create", etc.
	Condition string    `json:"condition"`  // "parent_not_found", "permission_denied", etc.
	RequestID string    `json:"request_id"` // Request correlation
	Timestamp time.Time `json:"timestamp"`  // When error occurred

	// Context data
	Context map[string]any `json:"context"` // Business-specific data
	Cause   error          `json:"-"`       // Original error (not serialized)

	// Message building
	message string // Cached formatted message
}

func (e *ServiceError) Error() string {
	if e.message == "" {
		e.message = strings.Join(e.buildMessage(), " ")
	}
	return e.message
}

func (e *ServiceError) Unwrap() error {
	return e.Cause
}

func (e *ServiceError) buildMessage() []string {
	parts := []string{}

	if e.Operation != "" {
		parts = append(parts, "operation", e.Operation)
	}

	if e.Step != "" {
		parts = append(parts, "step", e.Step)
	}

	if e.Condition != "" {
		parts = append(parts, "condition", e.Condition)
	}

	// Add key context values to message
	for key, value := range e.Context {
		switch key {
		case "album_id", "media_id", "filename", "album_path", "filepath", "parent_id", "job_id":
			parts = append(parts, key, fmt.Sprintf("%v", value))
		}
	}

	if e.RequestID != "" {
		parts = append(parts, "request_id", e.RequestID)
	}

	if e.Cause != nil {
		parts = append(parts, "cause", e.Cause.Error())
	}

	return parts
}

// Add context data
func (e *ServiceError) WithContext(key string, value any) *ServiceError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	e.message = "" // Reset cached message
	return e
}

// Add multiple context values
func (e *ServiceError) WithContextMap(context map[string]any) *ServiceError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	maps.Copy(e.Context, context)
	e.message = "" // Reset cached message

	return e
}

// NewServiceError creates a new service error with basic context
func NewServiceError(operation string) *ServiceError {
	return &ServiceError{
		Operation: operation,
		Timestamp: time.Now(),
		Context:   make(map[string]any),
	}
}

// NewServiceErrorWithContext creates a service error and extracts request context
func NewServiceErrorWithContext(ctx context.Context, operation string) *ServiceError {
	err := NewServiceError(operation)

	// Extract request ID if available
	if requestID, ok := ctx.Value("request_id").(string); ok {
		err.RequestID = requestID
	}

	return err
}

// Step-specific builders
func (e *ServiceError) AtStep(step string) *ServiceError {
	e.Step = step
	e.message = ""
	return e
}

func (e *ServiceError) WithCondition(condition string) *ServiceError {
	e.Condition = condition
	e.message = ""
	return e
}

func (e *ServiceError) WithCause(cause error) *ServiceError {
	e.Cause = cause
	e.message = ""
	return e
}

// Common context builders
func (e *ServiceError) WithAlbumID(albumID string) *ServiceError {
	return e.WithContext("album_id", albumID)
}

func (e *ServiceError) WithMediaID(mediaID string) *ServiceError {
	return e.WithContext("media_id", mediaID)
}

func (e *ServiceError) WithFilename(filename string) *ServiceError {
	return e.WithContext("filename", filename)
}

func (e *ServiceError) WithAlbumPath(path string) *ServiceError {
	return e.WithContext("album_path", path)
}

func (e *ServiceError) WithFilepath(filepath string) *ServiceError {
	return e.WithContext("filepath", filepath)
}

func (e *ServiceError) WithParentID(parentID string) *ServiceError {
	return e.WithContext("parent_id", parentID)
}

// Pre-defined error constructors
func NewAlbumNotFoundError(ctx context.Context, albumID string) *NotFoundError {
	return &NotFoundError{
		ServiceError: NewServiceErrorWithContext(ctx, "get_album").
			WithCondition("album_not_found").
			WithAlbumID(albumID),
	}
}

func NewAlbumExistsError(ctx context.Context, albumID, albumPath string) *ConflictError {
	return &ConflictError{
		ServiceError: NewServiceErrorWithContext(ctx, "create_album").
			WithCondition("album_already_exists").
			WithAlbumID(albumID).
			WithAlbumPath(albumPath),
	}
}

func NewParentAlbumNotFoundError(ctx context.Context, parentID string) *NotFoundError {
	return &NotFoundError{
		ServiceError: NewServiceErrorWithContext(ctx, "create_album").
			WithCondition("parent_album_not_found").
			WithParentID(parentID),
	}
}

func NewDatabaseWriteError(ctx context.Context, operation string, cause error) *InternalError {
	return &InternalError{
		ServiceError: NewServiceErrorWithContext(ctx, operation).
			AtStep("database_write").
			WithCondition("database_operation_failed").
			WithCause(cause),
	}
}

func NewFilesystemError(ctx context.Context, operation, step, filepath string, cause error) *InternalError {
	return &InternalError{
		ServiceError: NewServiceErrorWithContext(ctx, operation).
			AtStep(step).
			WithCondition("filesystem_operation_failed").
			WithFilepath(filepath).
			WithCause(cause),
	}
}

func NewMediaNotFoundError(ctx context.Context, mediaID string) *NotFoundError {
	return &NotFoundError{
		ServiceError: NewServiceErrorWithContext(ctx, "get_media").
			WithCondition("media_not_found").
			WithMediaID(mediaID),
	}
}

func NewMediaProcessingError(ctx context.Context, step, filename string, cause error) *InternalError {
	return &InternalError{
		ServiceError: NewServiceErrorWithContext(ctx, "write_media").
			AtStep(step).
			WithCondition("media_processing_failed").
			WithFilename(filename).
			WithCause(cause),
	}
}

func NewSyncJobError(ctx context.Context, step, jobID string, cause error) *InternalError {
	return &InternalError{
		ServiceError: NewServiceErrorWithContext(ctx, "sync_job").
			AtStep(step).
			WithCondition("sync_operation_failed").
			WithContext("job_id", jobID).
			WithCause(cause),
	}
}

// Specific error types for HTTP status mapping
type NotFoundError struct {
	*ServiceError
}

type ConflictError struct {
	*ServiceError
}

type ValidationError struct {
	*ServiceError
}

type InternalError struct {
	*ServiceError
}

// Type-specific constructors
func NewNotFoundError(ctx context.Context, operation, condition string) *NotFoundError {
	return &NotFoundError{
		ServiceError: NewServiceErrorWithContext(ctx, operation).WithCondition(condition),
	}
}

func NewConflictError(ctx context.Context, operation, condition string) *ConflictError {
	return &ConflictError{
		ServiceError: NewServiceErrorWithContext(ctx, operation).WithCondition(condition),
	}
}

func NewValidationError(ctx context.Context, operation, condition string) *ValidationError {
	return &ValidationError{
		ServiceError: NewServiceErrorWithContext(ctx, operation).WithCondition(condition),
	}
}

func NewInternalError(ctx context.Context, operation, condition string, cause error) *InternalError {
	return &InternalError{
		ServiceError: NewServiceErrorWithContext(ctx, operation).WithCondition(condition).WithCause(cause),
	}
}
