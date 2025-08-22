# Business Layer Error Context Plan

## Overview

Design a structured error system for Photos NG business layer that adds operational context without using `fmt.Errorf`. The goal is to capture WHERE in the business flow errors occur and WHAT conditions caused them.

## Core Error Structure Design

### 1. Base Business Error

```go
// internal/services/errors/business_error.go
package errors

import (
    "context"
    "fmt"
    "time"
)

// BusinessError provides structured error context for business operations
type BusinessError struct {
    // Core identification
    Operation   string    `json:"operation"`    // "create_album", "write_media", etc.
    Step        string    `json:"step"`        // "database_write", "filesystem_create", etc.
    Condition   string    `json:"condition"`   // "parent_not_found", "permission_denied", etc.
    RequestID   string    `json:"request_id"`  // Request correlation
    Timestamp   time.Time `json:"timestamp"`   // When error occurred
    
    // Context data
    Context     map[string]interface{} `json:"context"`     // Business-specific data
    Cause       error                 `json:"-"`           // Original error (not serialized)
    
    // Message building
    message     string                // Cached formatted message
}

func (e *BusinessError) Error() string {
    if e.message == "" {
        e.message = e.buildMessage()
    }
    return e.message
}

func (e *BusinessError) Unwrap() error {
    return e.Cause
}

func (e *BusinessError) buildMessage() string {
    parts := []string{}
    
    if e.Operation != "" {
        parts = append(parts, fmt.Sprintf("operation=%s", e.Operation))
    }
    
    if e.Step != "" {
        parts = append(parts, fmt.Sprintf("step=%s", e.Step))
    }
    
    if e.Condition != "" {
        parts = append(parts, fmt.Sprintf("condition=%s", e.Condition))
    }
    
    // Add key context values to message
    for key, value := range e.Context {
        switch key {
        case "album_id", "media_id", "filename", "album_path", "filepath":
            parts = append(parts, fmt.Sprintf("%s=%v", key, value))
        }
    }
    
    if e.RequestID != "" {
        parts = append(parts, fmt.Sprintf("request_id=%s", e.RequestID))
    }
    
    contextStr := strings.Join(parts, " ")
    
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", contextStr, e.Cause)
    }
    
    return contextStr
}

// Add context data
func (e *BusinessError) WithContext(key string, value interface{}) *BusinessError {
    if e.Context == nil {
        e.Context = make(map[string]interface{})
    }
    e.Context[key] = value
    e.message = "" // Reset cached message
    return e
}

// Add multiple context values
func (e *BusinessError) WithContextMap(context map[string]interface{}) *BusinessError {
    if e.Context == nil {
        e.Context = make(map[string]interface{})
    }
    for k, v := range context {
        e.Context[k] = v
    }
    e.message = "" // Reset cached message
    return e
}
```

### 2. Error Builder Functions

```go
// internal/services/errors/builders.go
package errors

import (
    "context"
    "time"
)

// NewBusinessError creates a new business error with basic context
func NewBusinessError(operation string) *BusinessError {
    return &BusinessError{
        Operation: operation,
        Timestamp: time.Now(),
        Context:   make(map[string]interface{}),
    }
}

// NewBusinessErrorWithContext creates a business error and extracts request context
func NewBusinessErrorWithContext(ctx context.Context, operation string) *BusinessError {
    err := NewBusinessError(operation)
    
    // Extract request ID if available
    if requestID, ok := ctx.Value("request_id").(string); ok {
        err.RequestID = requestID
    }
    
    return err
}

// Step-specific builders
func (e *BusinessError) AtStep(step string) *BusinessError {
    e.Step = step
    e.message = ""
    return e
}

func (e *BusinessError) WithCondition(condition string) *BusinessError {
    e.Condition = condition
    e.message = ""
    return e
}

func (e *BusinessError) WithCause(cause error) *BusinessError {
    e.Cause = cause
    e.message = ""
    return e
}

// Common context builders
func (e *BusinessError) WithAlbumID(albumID string) *BusinessError {
    return e.WithContext("album_id", albumID)
}

func (e *BusinessError) WithMediaID(mediaID string) *BusinessError {
    return e.WithContext("media_id", mediaID)
}

func (e *BusinessError) WithFilename(filename string) *BusinessError {
    return e.WithContext("filename", filename)
}

func (e *BusinessError) WithAlbumPath(path string) *BusinessError {
    return e.WithContext("album_path", path)
}

func (e *BusinessError) WithFilepath(filepath string) *BusinessError {
    return e.WithContext("filepath", filepath)
}

func (e *BusinessError) WithParentID(parentID string) *BusinessError {
    return e.WithContext("parent_id", parentID)
}
```

### 3. Pre-defined Error Constructors

```go
// internal/services/errors/constructors.go
package errors

import "context"

// Album operation errors
func NewAlbumNotFoundError(ctx context.Context, albumID string) *BusinessError {
    return NewBusinessErrorWithContext(ctx, "get_album").
        WithCondition("album_not_found").
        WithAlbumID(albumID)
}

func NewAlbumExistsError(ctx context.Context, albumID, albumPath string) *BusinessError {
    return NewBusinessErrorWithContext(ctx, "create_album").
        WithCondition("album_already_exists").
        WithAlbumID(albumID).
        WithAlbumPath(albumPath)
}

func NewParentAlbumNotFoundError(ctx context.Context, parentID string) *BusinessError {
    return NewBusinessErrorWithContext(ctx, "create_album").
        WithCondition("parent_album_not_found").
        WithParentID(parentID)
}

func NewDatabaseWriteError(ctx context.Context, operation string, cause error) *BusinessError {
    return NewBusinessErrorWithContext(ctx, operation).
        AtStep("database_write").
        WithCondition("database_operation_failed").
        WithCause(cause)
}

func NewFilesystemError(ctx context.Context, operation, step, filepath string, cause error) *BusinessError {
    return NewBusinessErrorWithContext(ctx, operation).
        AtStep(step).
        WithCondition("filesystem_operation_failed").
        WithFilepath(filepath).
        WithCause(cause)
}

// Media operation errors
func NewMediaNotFoundError(ctx context.Context, mediaID string) *BusinessError {
    return NewBusinessErrorWithContext(ctx, "get_media").
        WithCondition("media_not_found").
        WithMediaID(mediaID)
}

func NewMediaProcessingError(ctx context.Context, step, filename string, cause error) *BusinessError {
    return NewBusinessErrorWithContext(ctx, "write_media").
        AtStep(step).
        WithCondition("media_processing_failed").
        WithFilename(filename).
        WithCause(cause)
}

// Sync operation errors
func NewSyncJobError(ctx context.Context, step, jobID string, cause error) *BusinessError {
    return NewBusinessErrorWithContext(ctx, "sync_job").
        AtStep(step).
        WithCondition("sync_operation_failed").
        WithContext("job_id", jobID).
        WithCause(cause)
}
```

## Implementation in Services

### 1. Album Service Implementation

```go
// internal/services/album.go (Updated with BusinessError)
package services

import (
    "context"
    
    "git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
    "git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
    "git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
    "git.tls.tupangiu.ro/cosmin/photos-ng/internal/services/errors"
)

func (a *AlbumService) GetAlbum(ctx context.Context, id string) (*entity.Album, error) {
    // Input validation
    if id == "" {
        return nil, errors.NewBusinessErrorWithContext(ctx, "get_album").
            WithCondition("invalid_input").
            WithContext("validation_error", "empty_album_id")
    }
    
    // Database query
    album, err := a.dt.QueryAlbum(ctx, pg.FilterByAlbumId(id))
    if err != nil {
        return nil, errors.NewDatabaseWriteError(ctx, "get_album", err).
            WithAlbumID(id).
            AtStep("query_album")
    }
    
    if album == nil {
        return nil, errors.NewAlbumNotFoundError(ctx, id)
    }
    
    return album, nil
}

func (a *AlbumService) CreateAlbum(ctx context.Context, album entity.Album) (*entity.Album, error) {
    // Check if album already exists
    if existingAlbum, err := a.GetAlbum(ctx, album.ID); err == nil {
        return nil, errors.NewAlbumExistsError(ctx, album.ID, album.Path)
    } else {
        // Check if error was something other than "not found"
        var businessErr *errors.BusinessError
        if errors.As(err, &businessErr) && businessErr.Condition != "album_not_found" {
            return nil, errors.NewBusinessErrorWithContext(ctx, "create_album").
                AtStep("check_album_exists").
                WithCondition("validation_check_failed").
                WithAlbumID(album.ID).
                WithCause(err)
        }
    }
    
    // Handle parent album relationship
    if album.ParentId != nil {
        parent, err := a.GetAlbum(ctx, *album.ParentId)
        if err != nil {
            var businessErr *errors.BusinessError
            if errors.As(err, &businessErr) && businessErr.Condition == "album_not_found" {
                return nil, errors.NewParentAlbumNotFoundError(ctx, *album.ParentId)
            }
            return nil, errors.NewBusinessErrorWithContext(ctx, "create_album").
                AtStep("validate_parent").
                WithCondition("parent_lookup_failed").
                WithParentID(*album.ParentId).
                WithCause(err)
        }
        
        // Path computation
        originalPath := album.Path
        if !strings.HasPrefix(album.Path, parent.Path+"/") && album.Path != parent.Path {
            album.Path = path.Join(parent.Path, album.Path)
        }
        album.ID = entity.GenerateId(album.Path)
    }
    
    // Database transaction
    err := a.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
        // Write album to database
        if err := writer.WriteAlbum(ctx, album); err != nil {
            return errors.NewDatabaseWriteError(ctx, "create_album", err).
                WithAlbumID(album.ID).
                WithAlbumPath(album.Path)
        }
        
        // Create folder on filesystem
        if err := a.fs.CreateFolder(ctx, album.Path); err != nil {
            return errors.NewFilesystemError(ctx, "create_album", "filesystem_create", album.Path, err)
        }
        
        return nil
    })
    
    if err != nil {
        return nil, errors.NewBusinessErrorWithContext(ctx, "create_album").
            AtStep("transaction").
            WithCondition("transaction_failed").
            WithAlbumID(album.ID).
            WithAlbumPath(album.Path).
            WithCause(err)
    }
    
    return &album, nil
}

func (a *AlbumService) DeleteAlbum(ctx context.Context, id string) error {
    // Check if album exists
    album, err := a.GetAlbum(ctx, id)
    if err != nil {
        return errors.NewBusinessErrorWithContext(ctx, "delete_album").
            AtStep("validate_exists").
            WithCondition("album_lookup_failed").
            WithAlbumID(id).
            WithCause(err)
    }
    
    // Delete transaction
    err = a.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
        // Delete folder from filesystem
        if err := a.fs.DeleteFolder(ctx, album.Path); err != nil {
            return errors.NewFilesystemError(ctx, "delete_album", "filesystem_delete", album.Path, err)
        }
        
        // Delete from database
        if err := writer.DeleteAlbum(ctx, id); err != nil {
            return errors.NewDatabaseWriteError(ctx, "delete_album", err).
                WithAlbumID(id)
        }
        
        return nil
    })
    
    if err != nil {
        return errors.NewBusinessErrorWithContext(ctx, "delete_album").
            AtStep("transaction").
            WithCondition("transaction_failed").
            WithAlbumID(id).
            WithAlbumPath(album.Path).
            WithCause(err)
    }
    
    return nil
}
```

### 2. Media Service Implementation

```go
// internal/services/media.go (Updated with BusinessError)
func (m *MediaService) GetMediaByID(ctx context.Context, id string) (*entity.Media, error) {
    if id == "" {
        return nil, errors.NewBusinessErrorWithContext(ctx, "get_media").
            WithCondition("invalid_input").
            WithContext("validation_error", "empty_media_id")
    }
    
    media, err := m.dt.QueryMedia(ctx, pg.FilterByMediaId(id), pg.Limit(1))
    if err != nil {
        return nil, errors.NewDatabaseWriteError(ctx, "get_media", err).
            WithMediaID(id).
            AtStep("query_media")
    }
    
    if len(media) == 0 {
        return nil, errors.NewMediaNotFoundError(ctx, id)
    }
    
    processedMedia := media[0]
    processedMedia.Content = m.fs.Read(ctx, processedMedia.Filepath())
    
    return &processedMedia, nil
}

func (m *MediaService) WriteMedia(ctx context.Context, media entity.Media) (*entity.Media, error) {
    // Check if media already exists
    oldMedia, err := m.GetMediaByID(ctx, media.ID)
    if err != nil {
        var businessErr *errors.BusinessError
        if !errors.As(err, &businessErr) || businessErr.Condition != "media_not_found" {
            return nil, errors.NewBusinessErrorWithContext(ctx, "write_media").
                AtStep("check_existing").
                WithCondition("existence_check_failed").
                WithMediaID(media.ID).
                WithCause(err)
        }
    }
    
    // Read content
    content, err := media.Content()
    if err != nil {
        return nil, errors.NewMediaProcessingError(ctx, "read_content", media.Filename, err)
    }
    
    contentBytes, err := io.ReadAll(content)
    if err != nil {
        return nil, errors.NewMediaProcessingError(ctx, "read_content_bytes", media.Filename, err)
    }
    
    // Compute hash
    hash := sha256.Sum256(contentBytes)
    hashStr := fmt.Sprintf("%x", hash)
    
    if oldMedia != nil && hashStr == oldMedia.Hash {
        return oldMedia, nil // No change needed
    }
    
    // Transaction with detailed step tracking
    err = m.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
        media.Hash = hashStr
        
        // Step 1: Initialize processing service
        processingSrv, err := NewProcessingMediaService()
        if err != nil {
            return errors.NewMediaProcessingError(ctx, "init_processing", media.Filename, err)
        }
        
        // Step 2: Process media (thumbnail + EXIF)
        r, exif, err := processingSrv.Process(ctx, bytes.NewReader(contentBytes))
        if err != nil {
            return errors.NewMediaProcessingError(ctx, "generate_thumbnail", media.Filename, err)
        }
        
        // Step 3: Read thumbnail
        thumbnail, err := io.ReadAll(r)
        if err != nil {
            return errors.NewMediaProcessingError(ctx, "read_thumbnail", media.Filename, err)
        }
        
        media.Thumbnail = thumbnail
        media.Exif = exif
        
        // Step 4: Extract capture time
        if captureAt, err := media.GetCapturedTime(); err == nil {
            media.CapturedAt = captureAt
        }
        
        // Step 5: Write to filesystem
        if err := m.fs.Write(ctx, media.Filepath(), bytes.NewReader(contentBytes)); err != nil {
            return errors.NewFilesystemError(ctx, "write_media", "filesystem_write", media.Filepath(), err)
        }
        
        // Step 6: Write to database
        if err := writer.WriteMedia(ctx, media); err != nil {
            return errors.NewDatabaseWriteError(ctx, "write_media", err).
                WithMediaID(media.ID).
                WithFilename(media.Filename)
        }
        
        return nil
    })
    
    if err != nil {
        return nil, errors.NewBusinessErrorWithContext(ctx, "write_media").
            AtStep("transaction").
            WithCondition("transaction_failed").
            WithMediaID(media.ID).
            WithFilename(media.Filename).
            WithCause(err)
    }
    
    return &media, nil
}
```

### 3. Sync Service Implementation

```go
// internal/services/sync.go (Updated with BusinessError)
func (s *SyncService) StartSync(ctx context.Context, albumPath string) (string, error) {
    rootAlbum := entity.NewAlbum(albumPath)
    
    // Create sync job
    syncJob, err := NewSyncJob(rootAlbum, s.albumService, s.mediaService, s.fsDatastore)
    if err != nil {
        return "", errors.NewSyncJobError(ctx, "create_job", "", err).
            WithContext("album_path", albumPath)
    }
    
    // Add to scheduler
    if err := s.scheduler.Add(syncJob); err != nil {
        return "", errors.NewSyncJobError(ctx, "schedule_job", syncJob.GetId().String(), err).
            WithContext("album_path", albumPath)
    }
    
    return syncJob.GetId().String(), nil
}

func (s *SyncService) StopSyncJob(jobID string) error {
    syncJob := s.scheduler.Get(jobID)
    if syncJob == nil {
        return errors.NewBusinessError("stop_sync_job").
            WithCondition("job_not_found").
            WithContext("job_id", jobID)
    }
    
    if err := syncJob.Stop(); err != nil {
        return errors.NewSyncJobError(context.Background(), "stop_job", jobID, err)
    }
    
    return nil
}
```

## Implementation Timeline

### Week 1: Core Error Infrastructure
1. **Create error package** with BusinessError struct
2. **Implement builder functions** and constructors
3. **Add request ID middleware** for context extraction

### Week 2: Service Integration
1. **Update AlbumService** with structured errors
2. **Update MediaService** with step-by-step error tracking
3. **Update SyncService** with job-specific errors

### Week 3: Error Handling Enhancement
1. **Update HTTP handlers** to use BusinessError information
2. **Add error classification** for proper HTTP status codes
3. **Test error scenarios** and verify context information

## Expected Error Output

### Before:
```
Error: pq: duplicate key value violates unique constraint "albums_pkey"
```

### After:
```
Error: operation=create_album step=database_write condition=database_operation_failed album_id=123 album_path=/photos/vacation request_id=req_abc123: pq: duplicate key value violates unique constraint "albums_pkey"
```

### Multi-step failure:
```
Error: operation=write_media step=filesystem_write condition=filesystem_operation_failed media_id=456 filename=photo.jpg filepath=/photos/vacation/photo.jpg request_id=req_def456: permission denied
```

This provides clear operational context while preserving the detailed PostgreSQL error information, making debugging significantly faster and more precise.