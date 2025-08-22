# Error Handling Improvements for Photos NG

## Current Error Handling Analysis

After analyzing the Photos NG codebase, I found several areas where error handling could be significantly improved:

### Current State

**Strengths:**
- Basic custom error types in `internal/services/errors.go`
- Structured error responses in HTTP handlers
- Error wrapping with `fmt.Errorf` and `%w` verb
- Logging with zap structured logger

**Weaknesses:**
- Inconsistent error handling patterns across layers
- Limited error context and correlation IDs
- No proper error classification for API responses
- Missing validation error details
- Lack of error recovery mechanisms
- No centralized error handling middleware

## Specific Issues Found

### 1. **Inconsistent HTTP Error Responses**

**Current Pattern (from handlers):**
```go
func (s *Handler) ListAlbums(c *gin.Context, params v1.ListAlbumsParams) {
    albums, err := s.albumSrv.GetAlbums(c.Request.Context(), opts)
    if err != nil {
        zap.S().Errorw("failed to get albums", "error", err)
        c.JSON(http.StatusInternalServerError, v1.Error{
            Message: err.Error(),
        })
        return
    }
}
```

**Problems:**
- Always returns 500 Internal Server Error
- Exposes internal error messages to clients
- No error classification (400 vs 404 vs 500)
- No request correlation ID for debugging

### 2. **Limited Error Context**

**Current Pattern:**
```go
return fmt.Errorf("failed to create sync job: %w", err)
```

**Problems:**
- Missing contextual information (user ID, resource ID, operation details)
- No structured error data
- Hard to trace error through request lifecycle

### 3. **No Input Validation Error Details**

**Current Pattern:**
```go
type ErrInvalidInput struct {
    Field   string
    Value   interface{}
    Message string
}
```

**Problems:**
- Single field validation only
- No validation rule information
- No localization support

### 4. **Database Error Handling**

**Missing Patterns:**
- Connection pool exhaustion handling
- Transaction rollback error handling
- Constraint violation proper mapping
- Deadlock detection and retry

## Comprehensive Error Handling Improvements

### 1. **Enhanced Error Types & Classification**

```go
// internal/errors/types.go
package errors

import (
    "context"
    "fmt"
    "time"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
    // Client errors (4xx)
    ErrCodeBadRequest      ErrorCode = "BAD_REQUEST"
    ErrCodeUnauthorized    ErrorCode = "UNAUTHORIZED"
    ErrCodeForbidden       ErrorCode = "FORBIDDEN"
    ErrCodeNotFound        ErrorCode = "NOT_FOUND"
    ErrCodeConflict        ErrorCode = "CONFLICT"
    ErrCodeValidation      ErrorCode = "VALIDATION_ERROR"
    ErrCodeRateLimit       ErrorCode = "RATE_LIMIT"
    
    // Server errors (5xx)
    ErrCodeInternal        ErrorCode = "INTERNAL_ERROR"
    ErrCodeServiceUnavail  ErrorCode = "SERVICE_UNAVAILABLE"
    ErrCodeDatabaseError   ErrorCode = "DATABASE_ERROR"
    ErrCodeStorageError    ErrorCode = "STORAGE_ERROR"
    ErrCodeProcessingError ErrorCode = "PROCESSING_ERROR"
)

// AppError represents application-specific errors with rich context
type AppError struct {
    Code       ErrorCode              `json:"code"`
    Message    string                 `json:"message"`
    Details    map[string]interface{} `json:"details,omitempty"`
    Cause      error                  `json:"-"`
    StackTrace string                 `json:"-"`
    RequestID  string                 `json:"request_id,omitempty"`
    Timestamp  time.Time              `json:"timestamp"`
    
    // Context for debugging
    Operation  string `json:"-"`
    UserID     string `json:"-"`
    ResourceID string `json:"-"`
}

func (e *AppError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
    return e.Cause
}

// HTTP status mapping
func (e *AppError) HTTPStatus() int {
    switch e.Code {
    case ErrCodeBadRequest, ErrCodeValidation:
        return 400
    case ErrCodeUnauthorized:
        return 401
    case ErrCodeForbidden:
        return 403
    case ErrCodeNotFound:
        return 404
    case ErrCodeConflict:
        return 409
    case ErrCodeRateLimit:
        return 429
    case ErrCodeServiceUnavail:
        return 503
    default:
        return 500
    }
}

// Constructor functions
func NewNotFoundError(resource, id string) *AppError {
    return &AppError{
        Code:       ErrCodeNotFound,
        Message:    fmt.Sprintf("%s not found", resource),
        Details:    map[string]interface{}{"resource": resource, "id": id},
        Timestamp:  time.Now(),
        ResourceID: id,
    }
}

func NewValidationError(field, message string, value interface{}) *AppError {
    return &AppError{
        Code:    ErrCodeValidation,
        Message: "Validation failed",
        Details: map[string]interface{}{
            "field":   field,
            "message": message,
            "value":   value,
        },
        Timestamp: time.Now(),
    }
}

func NewInternalError(operation string, cause error) *AppError {
    return &AppError{
        Code:       ErrCodeInternal,
        Message:    "Internal server error",
        Cause:      cause,
        Operation:  operation,
        Timestamp:  time.Now(),
        StackTrace: getStackTrace(),
    }
}
```

### 2. **Centralized Error Handling Middleware**

```go
// internal/server/middlewares/error_handler.go
package middlewares

import (
    "context"
    "net/http"
    
    "git.tls.tupangiu.ro/cosmin/photos-ng/internal/errors"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "go.uber.org/zap"
)

func ErrorHandler(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Generate request ID
        requestID := uuid.New().String()
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)
        
        // Continue with request processing
        c.Next()
        
        // Handle any errors that occurred
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            handleError(c, err, requestID, logger)
        }
    }
}

func handleError(c *gin.Context, err error, requestID string, logger *zap.Logger) {
    var appErr *errors.AppError
    
    // Check if it's already an AppError
    if !errors.As(err, &appErr) {
        // Convert unknown errors to internal errors
        appErr = errors.NewInternalError("unknown_operation", err)
    }
    
    // Add request context
    appErr.RequestID = requestID
    
    // Log error with appropriate level
    logFields := []zap.Field{
        zap.String("request_id", requestID),
        zap.String("error_code", string(appErr.Code)),
        zap.String("operation", appErr.Operation),
        zap.Any("details", appErr.Details),
        zap.String("method", c.Request.Method),
        zap.String("path", c.Request.URL.Path),
    }
    
    if appErr.HTTPStatus() >= 500 {
        logger.Error("Server error", append(logFields, zap.Error(appErr.Cause))...)
    } else {
        logger.Warn("Client error", logFields...)
    }
    
    // Return appropriate response
    response := gin.H{
        "error": gin.H{
            "code":       appErr.Code,
            "message":    getClientMessage(appErr),
            "request_id": requestID,
            "timestamp":  appErr.Timestamp,
        },
    }
    
    // Add details for client errors (not server errors)
    if appErr.HTTPStatus() < 500 && appErr.Details != nil {
        response["error"].(gin.H)["details"] = appErr.Details
    }
    
    c.JSON(appErr.HTTPStatus(), response)
}

// getClientMessage returns sanitized messages for clients
func getClientMessage(err *errors.AppError) string {
    switch err.Code {
    case errors.ErrCodeNotFound:
        return "The requested resource was not found"
    case errors.ErrCodeValidation:
        return "Validation failed"
    case errors.ErrCodeConflict:
        return "Resource already exists"
    case errors.ErrCodeUnauthorized:
        return "Authentication required"
    case errors.ErrCodeForbidden:
        return "Access denied"
    default:
        return "An internal error occurred"
    }
}
```

### 3. **Enhanced Service Layer Error Handling**

```go
// internal/services/album.go (enhanced)
package services

import (
    "context"
    "database/sql"
    
    "git.tls.tupangiu.ro/cosmin/photos-ng/internal/errors"
)

func (s *AlbumService) GetAlbum(ctx context.Context, id string) (*Album, error) {
    // Add operation context
    const operation = "album.get"
    
    // Validate input
    if id == "" {
        return nil, errors.NewValidationError("id", "album ID is required", id)
    }
    
    // Add timeout context
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    album, err := s.datastore.GetAlbum(ctx, id)
    if err != nil {
        return nil, s.handleDatabaseError(operation, "get_album", err, map[string]interface{}{
            "album_id": id,
        })
    }
    
    if album == nil {
        return nil, errors.NewNotFoundError("album", id)
    }
    
    return album, nil
}

func (s *AlbumService) handleDatabaseError(operation, query string, err error, context map[string]interface{}) error {
    switch {
    case errors.Is(err, sql.ErrNoRows):
        return errors.NewNotFoundError("resource", "")
    case errors.Is(err, context.DeadlineExceeded):
        return &errors.AppError{
            Code:      errors.ErrCodeServiceUnavail,
            Message:   "Database timeout",
            Operation: operation,
            Details:   context,
            Cause:     err,
        }
    case isConstraintViolation(err):
        return &errors.AppError{
            Code:      errors.ErrCodeConflict,
            Message:   "Resource constraint violation",
            Operation: operation,
            Details:   context,
            Cause:     err,
        }
    default:
        return &errors.AppError{
            Code:      errors.ErrCodeDatabaseError,
            Message:   "Database operation failed",
            Operation: operation,
            Details:   context,
            Cause:     err,
        }
    }
}
```

### 4. **Input Validation with Detailed Errors**

```go
// internal/validation/validator.go
package validation

import (
    "fmt"
    "reflect"
    "strings"
    
    "git.tls.tupangiu.ro/cosmin/photos-ng/internal/errors"
    "github.com/go-playground/validator/v10"
)

type Validator struct {
    validator *validator.Validate
}

func NewValidator() *Validator {
    v := validator.New()
    
    // Register custom validators
    v.RegisterValidation("album_name", validateAlbumName)
    v.RegisterValidation("media_type", validateMediaType)
    
    return &Validator{validator: v}
}

func (v *Validator) ValidateStruct(s interface{}) error {
    err := v.validator.Struct(s)
    if err == nil {
        return nil
    }
    
    var validationErrors []map[string]interface{}
    
    for _, err := range err.(validator.ValidationErrors) {
        validationErrors = append(validationErrors, map[string]interface{}{
            "field":   getJSONFieldName(err),
            "value":   err.Value(),
            "rule":    err.Tag(),
            "message": getValidationMessage(err),
        })
    }
    
    return &errors.AppError{
        Code:    errors.ErrCodeValidation,
        Message: "Validation failed",
        Details: map[string]interface{}{
            "validation_errors": validationErrors,
        },
    }
}

func getValidationMessage(err validator.FieldError) string {
    switch err.Tag() {
    case "required":
        return "This field is required"
    case "min":
        return fmt.Sprintf("Minimum length is %s", err.Param())
    case "max":
        return fmt.Sprintf("Maximum length is %s", err.Param())
    case "album_name":
        return "Album name contains invalid characters"
    case "media_type":
        return "Media type must be 'photo' or 'video'"
    default:
        return "Invalid value"
    }
}
```

### 5. **Retry Logic for Transient Errors**

```go
// internal/utils/retry.go
package utils

import (
    "context"
    "time"
    
    "git.tls.tupangiu.ro/cosmin/photos-ng/internal/errors"
)

type RetryConfig struct {
    MaxAttempts int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
    Multiplier  float64
}

func DefaultRetryConfig() RetryConfig {
    return RetryConfig{
        MaxAttempts: 3,
        BaseDelay:   100 * time.Millisecond,
        MaxDelay:    5 * time.Second,
        Multiplier:  2.0,
    }
}

func WithRetry(ctx context.Context, config RetryConfig, operation func() error) error {
    var lastErr error
    
    for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
        err := operation()
        if err == nil {
            return nil
        }
        
        lastErr = err
        
        // Don't retry client errors
        var appErr *errors.AppError
        if errors.As(err, &appErr) && appErr.HTTPStatus() < 500 {
            return err
        }
        
        // Don't retry on last attempt
        if attempt == config.MaxAttempts {
            break
        }
        
        // Calculate delay with exponential backoff
        delay := time.Duration(float64(config.BaseDelay) * 
                              math.Pow(config.Multiplier, float64(attempt-1)))
        if delay > config.MaxDelay {
            delay = config.MaxDelay
        }
        
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(delay):
            // Continue to next attempt
        }
    }
    
    return lastErr
}

// Usage in services
func (s *MediaService) ProcessMedia(ctx context.Context, mediaID string) error {
    return WithRetry(ctx, DefaultRetryConfig(), func() error {
        return s.processMediaInternal(ctx, mediaID)
    })
}
```

### 6. **Circuit Breaker for External Dependencies**

```go
// internal/utils/circuit_breaker.go
package utils

import (
    "context"
    "sync"
    "time"
    
    "git.tls.tupangiu.ro/cosmin/photos-ng/internal/errors"
)

type CircuitBreaker struct {
    mu           sync.RWMutex
    state        State
    failures     int
    lastFailTime time.Time
    
    maxFailures  int
    timeout      time.Duration
    resetTimeout time.Duration
}

type State int

const (
    StateClosed State = iota
    StateOpen
    StateHalfOpen
)

func NewCircuitBreaker(maxFailures int, timeout, resetTimeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        state:        StateClosed,
        maxFailures:  maxFailures,
        timeout:      timeout,
        resetTimeout: resetTimeout,
    }
}

func (cb *CircuitBreaker) Execute(ctx context.Context, operation func() error) error {
    if !cb.canExecute() {
        return &errors.AppError{
            Code:    errors.ErrCodeServiceUnavail,
            Message: "Circuit breaker is open",
        }
    }
    
    err := operation()
    cb.recordResult(err)
    return err
}
```

## Implementation Priority

### Phase 1: Foundation (1-2 weeks)
1. **Implement AppError type** with proper classification
2. **Add error handling middleware** for consistent responses
3. **Update service layer** to use new error types
4. **Add request ID tracking** for debugging

### Phase 2: Enhanced Validation (1 week)
1. **Implement comprehensive input validation** with detailed errors
2. **Add validation middleware** for request bodies
3. **Update API responses** to include validation details

### Phase 3: Resilience (2-3 weeks)
1. **Add retry logic** for transient failures
2. **Implement circuit breakers** for external dependencies
3. **Add timeout handling** for long-running operations
4. **Implement graceful degradation** for non-critical features

### Phase 4: Observability (1-2 weeks)
1. **Enhanced error logging** with structured data
2. **Error metrics and monitoring** integration
3. **Error aggregation** for trend analysis
4. **Alert configuration** for critical errors

## Expected Benefits

- **Better User Experience**: Clear, actionable error messages
- **Improved Debugging**: Request correlation and structured logging
- **Enhanced Reliability**: Retry logic and circuit breakers
- **Better Monitoring**: Standardized error codes and metrics
- **Faster Development**: Consistent error handling patterns
- **Reduced Support Burden**: Better error documentation and context