# Business Layer Logging & Error Management Improvement Plan

## Current State Analysis

After analyzing the Photos NG business layer (services), I've identified critical gaps in error handling and logging that make debugging difficult when errors occur.

### Current Logging Pattern Analysis

**Limited Logging Examples Found:**
```go
// Media Service - Only 2 log statements
zap.S().Debugw("update media skipped.same hash", "hash", oldMedia.Hash)
zap.S().Warnw("failed to get captured at timestamp", "error", err, "filename", media.Filepath())

// Album Service - NO logging found

// Sync Service - Basic logging
zap.S().Errorw("failed to create sync job", "path", albumPath, "error", err)
zap.S().Infow("sync job created and scheduled", "job_id", jobID, "path", albumPath)

// Processing Service - Only 1 log statement
zap.S().Warnw("failed to read exif metadata value", "error", "value is not a string", "value", v)
```

### Critical Problems Identified

#### 1. **Silent Failures Everywhere**
```go
// Current pattern - NO context when errors occur
func (a *AlbumService) GetAlbum(ctx context.Context, id string) (*entity.Album, error) {
    album, err := a.dt.QueryAlbum(ctx, pg.FilterByAlbumId(id))
    if err != nil {
        return nil, err  // Silent failure - no logging!
    }
    // ...
}

// When this fails, you have NO idea:
// - Which album ID was requested
// - What database error occurred  
// - Who made the request
// - What query was executed
```

#### 2. **Database Operations Have Zero Visibility**
```go
// Current pattern - Database operations are black boxes
func (m *MediaService) WriteMedia(ctx context.Context, media entity.Media) (*entity.Media, error) {
    err = m.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
        // Multiple operations happen here with NO logging:
        // 1. File processing
        // 2. Thumbnail generation  
        // 3. EXIF extraction
        // 4. Database write
        // 5. File system write
        
        // If ANY step fails, you have no idea which one!
        return writer.WriteMedia(ctx, media)
    })
    if err != nil {
        return nil, err  // Silent failure again!
    }
}
```

#### 3. **Transaction Failures Are Invisible**
```go
// WriteTx failures provide no context
err := a.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
    // Could fail at:
    // - Transaction start
    // - File system operation
    // - Database write
    // - Transaction commit
    // But you'll never know which!
})
```

#### 4. **Business Logic Errors Have No Context**
```go
// Example: Album creation failure
func (a *AlbumService) CreateAlbum(ctx context.Context, album entity.Album) (*entity.Album, error) {
    // No logging when:
    // - Parent album doesn't exist
    // - Path computation fails  
    // - Album already exists
    // - File system creation fails
}
```

## Comprehensive Improvement Plan

### Phase 1: Structured Service Logging Infrastructure

#### 1.1 Create Service-Specific Loggers

```go
// internal/services/logging.go
package services

import (
    "context"
    "time"
    
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// ServiceLogger provides structured logging for business services
type ServiceLogger struct {
    logger    *zap.Logger
    service   string
    requestID string
}

func NewServiceLogger(service string) *ServiceLogger {
    return &ServiceLogger{
        logger:  zap.L().Named(service),
        service: service,
    }
}

func (l *ServiceLogger) WithContext(ctx context.Context) *ServiceLogger {
    requestID := GetRequestID(ctx) // Extract from context
    return &ServiceLogger{
        logger:    l.logger.With(zap.String("request_id", requestID)),
        service:   l.service,
        requestID: requestID,
    }
}

// Operation logging methods
func (l *ServiceLogger) StartOperation(operation string, params map[string]interface{}) *OperationLogger {
    start := time.Now()
    
    l.logger.Info("Operation started",
        zap.String("operation", operation),
        zap.Any("params", params),
        zap.Time("started_at", start),
    )
    
    return &OperationLogger{
        ServiceLogger: l,
        operation:     operation,
        startTime:     start,
        params:        params,
    }
}

type OperationLogger struct {
    *ServiceLogger
    operation string
    startTime time.Time
    params    map[string]interface{}
}

func (ol *OperationLogger) Success(result interface{}) {
    duration := time.Since(ol.startTime)
    ol.logger.Info("Operation completed successfully",
        zap.String("operation", ol.operation),
        zap.Duration("duration", duration),
        zap.Any("result", result),
    )
}

func (ol *OperationLogger) Error(err error, context map[string]interface{}) {
    duration := time.Since(ol.startTime)
    
    fields := []zap.Field{
        zap.String("operation", ol.operation),
        zap.Duration("duration", duration),
        zap.Error(err),
        zap.Any("params", ol.params),
    }
    
    if context != nil {
        fields = append(fields, zap.Any("error_context", context))
    }
    
    ol.logger.Error("Operation failed", fields...)
}

func (ol *OperationLogger) Warn(message string, context map[string]interface{}) {
    ol.logger.Warn(message,
        zap.String("operation", ol.operation),
        zap.Any("context", context),
    )
}
```

#### 1.2 Enhanced Album Service with Comprehensive Logging

```go
// internal/services/album.go (Enhanced)
package services

import (
    "context"
    "fmt"
    
    "git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
    "git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
    "git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

type AlbumService struct {
    dt     *pg.Datastore
    fs     *fs.Datastore
    logger *ServiceLogger
}

func NewAlbumService(dt *pg.Datastore, fsDatastore *fs.Datastore) *AlbumService {
    return &AlbumService{
        dt:     dt,
        fs:     fsDatastore,
        logger: NewServiceLogger("album_service"),
    }
}

func (a *AlbumService) GetAlbum(ctx context.Context, id string) (*entity.Album, error) {
    log := a.logger.WithContext(ctx).StartOperation("get_album", map[string]interface{}{
        "album_id": id,
    })
    
    // Input validation logging
    if id == "" {
        err := fmt.Errorf("album ID cannot be empty")
        log.Error(err, map[string]interface{}{
            "validation_error": "empty_album_id",
        })
        return nil, err
    }
    
    // Database query logging
    log.ServiceLogger.logger.Debug("Querying album from database",
        zap.String("album_id", id),
        zap.String("query_type", "single_album"),
    )
    
    album, err := a.dt.QueryAlbum(ctx, pg.FilterByAlbumId(id))
    if err != nil {
        log.Error(err, map[string]interface{}{
            "database_operation": "query_album",
            "album_id":          id,
            "error_type":        "database_error",
        })
        return nil, fmt.Errorf("failed to query album %s: %w", id, err)
    }
    
    if album == nil {
        err := NewErrAlbumNotFound(id)
        log.Error(err, map[string]interface{}{
            "album_id":    id,
            "error_type":  "not_found",
            "search_type": "by_id",
        })
        return nil, err
    }
    
    log.Success(map[string]interface{}{
        "album_id":   album.ID,
        "album_name": album.Name,
        "album_path": album.Path,
    })
    
    return album, nil
}

func (a *AlbumService) CreateAlbum(ctx context.Context, album entity.Album) (*entity.Album, error) {
    log := a.logger.WithContext(ctx).StartOperation("create_album", map[string]interface{}{
        "album_id":   album.ID,
        "album_name": album.Name,
        "album_path": album.Path,
        "parent_id":  album.ParentId,
    })
    
    // Check if album already exists
    log.ServiceLogger.logger.Debug("Checking if album already exists")
    isAlbumExists := true
    if _, err := a.GetAlbum(ctx, album.ID); err != nil && IsErrResourceNotFound(err) {
        isAlbumExists = false
        log.ServiceLogger.logger.Debug("Album does not exist, proceeding with creation")
    } else if err != nil {
        log.Error(err, map[string]interface{}{
            "check_operation": "album_exists",
            "album_id":       album.ID,
        })
        return nil, err
    } else {
        log.Warn("Album already exists", map[string]interface{}{
            "album_id":       album.ID,
            "existing_album": true,
        })
    }
    
    // Handle parent album relationship
    if album.ParentId != nil {
        log.ServiceLogger.logger.Debug("Processing parent album relationship",
            zap.String("parent_id", *album.ParentId),
        )
        
        parent, err := a.GetAlbum(ctx, *album.ParentId)
        if err != nil {
            if IsErrResourceNotFound(err) {
                err := fmt.Errorf("parent album %s does not exist", *album.ParentId)
                log.Error(err, map[string]interface{}{
                    "parent_id":    *album.ParentId,
                    "error_type":   "parent_not_found",
                    "album_path":   album.Path,
                })
                return nil, err
            }
            log.Error(err, map[string]interface{}{
                "parent_operation": "get_parent_album",
                "parent_id":       *album.ParentId,
            })
            return nil, err
        }
        
        // Path computation logging
        originalPath := album.Path
        if !strings.HasPrefix(album.Path, parent.Path+"/") && album.Path != parent.Path {
            album.Path = path.Join(parent.Path, album.Path)
            log.ServiceLogger.logger.Debug("Album path updated for parent relationship",
                zap.String("original_path", originalPath),
                zap.String("computed_path", album.Path),
                zap.String("parent_path", parent.Path),
            )
        }
        album.ID = entity.GenerateId(album.Path)
    }
    
    // Database transaction with detailed logging
    log.ServiceLogger.logger.Debug("Starting database transaction for album creation")
    
    err := a.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
        // Database write
        log.ServiceLogger.logger.Debug("Writing album to database")
        if err := writer.WriteAlbum(ctx, album); err != nil {
            log.ServiceLogger.logger.Error("Database write failed",
                zap.Error(err),
                zap.String("album_id", album.ID),
                zap.String("album_path", album.Path),
            )
            return fmt.Errorf("database write failed: %w", err)
        }
        
        // File system operation
        if !isAlbumExists {
            log.ServiceLogger.logger.Debug("Creating album folder on filesystem",
                zap.String("folder_path", album.Path),
            )
            
            if err := a.fs.CreateFolder(ctx, album.Path); err != nil {
                log.ServiceLogger.logger.Error("Filesystem folder creation failed",
                    zap.Error(err),
                    zap.String("folder_path", album.Path),
                )
                return fmt.Errorf("failed to create folder %s: %w", album.Path, err)
            }
            
            log.ServiceLogger.logger.Debug("Album folder created successfully",
                zap.String("folder_path", album.Path),
            )
        }
        
        return nil
    })
    
    if err != nil {
        log.Error(err, map[string]interface{}{
            "transaction_operation": "create_album_tx",
            "album_id":             album.ID,
            "album_path":           album.Path,
            "filesystem_operation": "create_folder",
        })
        return nil, err
    }
    
    log.Success(map[string]interface{}{
        "album_id":     album.ID,
        "album_name":   album.Name,
        "album_path":   album.Path,
        "was_existing": isAlbumExists,
    })
    
    return &album, nil
}

func (a *AlbumService) DeleteAlbum(ctx context.Context, id string) error {
    log := a.logger.WithContext(ctx).StartOperation("delete_album", map[string]interface{}{
        "album_id": id,
    })
    
    // Check if album exists first
    album, err := a.GetAlbum(ctx, id)
    if err != nil {
        log.Error(err, map[string]interface{}{
            "pre_check_operation": "verify_album_exists",
        })
        return err
    }
    
    log.ServiceLogger.logger.Debug("Album found, proceeding with deletion",
        zap.String("album_path", album.Path),
        zap.String("album_name", album.Name),
    )
    
    // Transaction with detailed logging
    err = a.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
        // File system deletion
        log.ServiceLogger.logger.Debug("Deleting album folder from filesystem",
            zap.String("folder_path", album.Path),
        )
        
        if err := a.fs.DeleteFolder(ctx, album.Path); err != nil {
            log.ServiceLogger.logger.Error("Filesystem folder deletion failed",
                zap.Error(err),
                zap.String("folder_path", album.Path),
            )
            return fmt.Errorf("failed to delete folder %s: %w", album.Path, err)
        }
        
        // Database deletion
        log.ServiceLogger.logger.Debug("Deleting album from database")
        if err := writer.DeleteAlbum(ctx, id); err != nil {
            log.ServiceLogger.logger.Error("Database deletion failed",
                zap.Error(err),
                zap.String("album_id", id),
            )
            return fmt.Errorf("failed to delete album from database: %w", err)
        }
        
        return nil
    })
    
    if err != nil {
        log.Error(err, map[string]interface{}{
            "transaction_operation": "delete_album_tx",
            "album_id":             id,
            "album_path":           album.Path,
        })
        return err
    }
    
    log.Success(map[string]interface{}{
        "album_id":   id,
        "album_path": album.Path,
        "deleted":    true,
    })
    
    return nil
}
```

#### 1.3 Enhanced Media Service with Transaction Logging

```go
// internal/services/media.go (Enhanced)
func (m *MediaService) WriteMedia(ctx context.Context, media entity.Media) (*entity.Media, error) {
    log := m.logger.WithContext(ctx).StartOperation("write_media", map[string]interface{}{
        "media_id":   media.ID,
        "filename":   media.Filename,
        "album_id":   media.Album.ID,
        "album_path": media.Album.Path,
    })
    
    // Check for existing media
    oldMedia, err := m.GetMediaByID(ctx, media.ID)
    if err != nil && !IsErrResourceNotFound(err) {
        log.Error(err, map[string]interface{}{
            "check_operation": "existing_media",
        })
        return nil, err
    }
    
    // Content reading with logging
    log.ServiceLogger.logger.Debug("Reading media content for processing")
    content, err := media.Content()
    if err != nil {
        log.Error(err, map[string]interface{}{
            "operation": "read_content",
            "filename":  media.Filename,
        })
        return nil, err
    }
    
    contentBytes, err := io.ReadAll(content)
    if err != nil {
        log.Error(err, map[string]interface{}{
            "operation":    "read_content_bytes",
            "filename":     media.Filename,
            "content_size": len(contentBytes),
        })
        return nil, fmt.Errorf("failed to read media content: %w", err)
    }
    
    log.ServiceLogger.logger.Debug("Content read successfully",
        zap.String("filename", media.Filename),
        zap.Int("content_size", len(contentBytes)),
    )
    
    // Hash computation with logging
    hash := sha256.Sum256(contentBytes)
    hashStr := fmt.Sprintf("%x", hash)
    
    if oldMedia != nil && hashStr == oldMedia.Hash {
        log.ServiceLogger.logger.Debug("Media content unchanged, skipping processing",
            zap.String("hash", hashStr),
            zap.String("filename", media.Filename),
        )
        log.Success(map[string]interface{}{
            "media_id":  oldMedia.ID,
            "filename":  oldMedia.Filename,
            "hash":      hashStr,
            "skipped":   true,
            "reason":    "unchanged_content",
        })
        return oldMedia, nil
    }
    
    // Transaction with detailed step logging
    log.ServiceLogger.logger.Debug("Starting media processing transaction")
    
    err = m.dt.WriteTx(ctx, func(ctx context.Context, writer *pg.Writer) error {
        media.Hash = hashStr
        
        // Step 1: Initialize processing service
        log.ServiceLogger.logger.Debug("Initializing media processing service")
        processingSrv, err := NewProcessingMediaService()
        if err != nil {
            log.ServiceLogger.logger.Error("Failed to initialize processing service", zap.Error(err))
            return fmt.Errorf("processing service initialization failed: %w", err)
        }
        
        // Step 2: Process media (thumbnail + EXIF)
        log.ServiceLogger.logger.Debug("Processing media content",
            zap.String("filename", media.Filename),
            zap.Int("content_size", len(contentBytes)),
        )
        
        r, exif, err := processingSrv.Process(ctx, bytes.NewReader(contentBytes))
        if err != nil {
            log.ServiceLogger.logger.Error("Media processing failed",
                zap.Error(err),
                zap.String("filename", media.Filename),
                zap.Int("content_size", len(contentBytes)),
            )
            return fmt.Errorf("media processing failed: %w", err)
        }
        
        log.ServiceLogger.logger.Debug("Media processing completed",
            zap.String("filename", media.Filename),
            zap.Int("exif_fields", len(exif)),
        )
        
        // Step 3: Read thumbnail
        thumbnail, err := io.ReadAll(r)
        if err != nil {
            log.ServiceLogger.logger.Error("Failed to read thumbnail data", zap.Error(err))
            return fmt.Errorf("failed to read thumbnail: %w", err)
        }
        
        media.Thumbnail = thumbnail
        media.Exif = exif
        
        log.ServiceLogger.logger.Debug("Thumbnail generated",
            zap.String("filename", media.Filename),
            zap.Int("thumbnail_size", len(thumbnail)),
        )
        
        // Step 4: Extract capture time
        if captureAt, err := media.GetCapturedTime(); err != nil {
            log.ServiceLogger.logger.Warn("Failed to extract capture time from EXIF",
                zap.Error(err),
                zap.String("filename", media.Filename),
            )
        } else {
            media.CapturedAt = captureAt
            log.ServiceLogger.logger.Debug("Capture time extracted",
                zap.String("filename", media.Filename),
                zap.Time("captured_at", captureAt),
            )
        }
        
        // Step 5: Write file to disk
        log.ServiceLogger.logger.Debug("Writing media file to disk",
            zap.String("filepath", media.Filepath()),
        )
        
        if err := m.fs.Write(ctx, media.Filepath(), bytes.NewReader(contentBytes)); err != nil {
            log.ServiceLogger.logger.Error("Failed to write media file to disk",
                zap.Error(err),
                zap.String("filepath", media.Filepath()),
                zap.Int("content_size", len(contentBytes)),
            )
            return fmt.Errorf("failed to write file %s: %w", media.Filepath(), err)
        }
        
        // Step 6: Write to database
        log.ServiceLogger.logger.Debug("Writing media metadata to database")
        if err := writer.WriteMedia(ctx, media); err != nil {
            log.ServiceLogger.logger.Error("Failed to write media to database",
                zap.Error(err),
                zap.String("media_id", media.ID),
                zap.String("filename", media.Filename),
            )
            return fmt.Errorf("database write failed: %w", err)
        }
        
        return nil
    })
    
    if err != nil {
        log.Error(err, map[string]interface{}{
            "transaction_operation": "write_media_tx",
            "media_id":             media.ID,
            "filename":             media.Filename,
            "content_size":         len(contentBytes),
            "hash":                 hashStr,
        })
        return nil, err
    }
    
    log.Success(map[string]interface{}{
        "media_id":       media.ID,
        "filename":       media.Filename,
        "hash":           hashStr,
        "content_size":   len(contentBytes),
        "thumbnail_size": len(media.Thumbnail),
        "exif_fields":    len(media.Exif),
        "captured_at":    media.CapturedAt,
    })
    
    return &media, nil
}
```

### Phase 2: Database Operation Instrumentation

#### 2.1 Database Query Logging Wrapper

```go
// internal/datastore/pg/instrumented.go
package pg

import (
    "context"
    "time"
    
    "go.uber.org/zap"
)

type InstrumentedDatastore struct {
    *Datastore
    logger *zap.Logger
}

func NewInstrumentedDatastore(ds *Datastore) *InstrumentedDatastore {
    return &InstrumentedDatastore{
        Datastore: ds,
        logger:    zap.L().Named("datastore"),
    }
}

func (i *InstrumentedDatastore) QueryAlbum(ctx context.Context, queries ...Query) (*entity.Album, error) {
    start := time.Now()
    requestID := GetRequestID(ctx)
    
    i.logger.Debug("Database query started",
        zap.String("request_id", requestID),
        zap.String("operation", "query_album"),
        zap.Int("filter_count", len(queries)),
    )
    
    album, err := i.Datastore.QueryAlbum(ctx, queries...)
    duration := time.Since(start)
    
    if err != nil {
        i.logger.Error("Database query failed",
            zap.String("request_id", requestID),
            zap.String("operation", "query_album"),
            zap.Duration("duration", duration),
            zap.Error(err),
            zap.Int("filter_count", len(queries)),
        )
        return nil, err
    }
    
    i.logger.Debug("Database query completed",
        zap.String("request_id", requestID),
        zap.String("operation", "query_album"),
        zap.Duration("duration", duration),
        zap.Bool("found", album != nil),
    )
    
    return album, nil
}

func (i *InstrumentedDatastore) WriteTx(ctx context.Context, fn func(context.Context, *Writer) error) error {
    start := time.Now()
    requestID := GetRequestID(ctx)
    
    i.logger.Debug("Database transaction started",
        zap.String("request_id", requestID),
        zap.String("operation", "write_transaction"),
    )
    
    err := i.Datastore.WriteTx(ctx, func(ctx context.Context, writer *Writer) error {
        // Wrap writer with instrumentation
        instrumentedWriter := &InstrumentedWriter{
            Writer: writer,
            logger: i.logger.With(zap.String("request_id", requestID)),
        }
        
        return fn(ctx, instrumentedWriter)
    })
    
    duration := time.Since(start)
    
    if err != nil {
        i.logger.Error("Database transaction failed",
            zap.String("request_id", requestID),
            zap.String("operation", "write_transaction"),
            zap.Duration("duration", duration),
            zap.Error(err),
        )
        return err
    }
    
    i.logger.Debug("Database transaction completed",
        zap.String("request_id", requestID),
        zap.String("operation", "write_transaction"),
        zap.Duration("duration", duration),
    )
    
    return nil
}

type InstrumentedWriter struct {
    *Writer
    logger *zap.Logger
}

func (w *InstrumentedWriter) WriteAlbum(ctx context.Context, album entity.Album) error {
    start := time.Now()
    
    w.logger.Debug("Writing album to database",
        zap.String("album_id", album.ID),
        zap.String("album_name", album.Name),
        zap.String("album_path", album.Path),
    )
    
    err := w.Writer.WriteAlbum(ctx, album)
    duration := time.Since(start)
    
    if err != nil {
        w.logger.Error("Album write failed",
            zap.String("album_id", album.ID),
            zap.Duration("duration", duration),
            zap.Error(err),
        )
        return err
    }
    
    w.logger.Debug("Album write completed",
        zap.String("album_id", album.ID),
        zap.Duration("duration", duration),
    )
    
    return nil
}
```

### Phase 3: Error Context Enhancement

#### 3.1 Rich Error Context for Services

```go
// internal/services/errors.go (Enhanced)
package services

import (
    "context"
    "fmt"
    "time"
)

type ServiceError struct {
    Code        string                 `json:"code"`
    Message     string                 `json:"message"`
    Service     string                 `json:"service"`
    Operation   string                 `json:"operation"`
    RequestID   string                 `json:"request_id"`
    Timestamp   time.Time              `json:"timestamp"`
    Context     map[string]interface{} `json:"context"`
    Cause       error                  `json:"-"`
}

func (e *ServiceError) Error() string {
    return fmt.Sprintf("[%s:%s] %s", e.Service, e.Operation, e.Message)
}

func (e *ServiceError) Unwrap() error {
    return e.Cause
}

func NewServiceError(ctx context.Context, service, operation, code, message string) *ServiceError {
    return &ServiceError{
        Code:      code,
        Message:   message,
        Service:   service,
        Operation: operation,
        RequestID: GetRequestID(ctx),
        Timestamp: time.Now(),
        Context:   make(map[string]interface{}),
    }
}

func (e *ServiceError) WithContext(key string, value interface{}) *ServiceError {
    e.Context[key] = value
    return e
}

func (e *ServiceError) WithCause(err error) *ServiceError {
    e.Cause = err
    return e
}

// Error constructors for common scenarios
func NewAlbumNotFoundError(ctx context.Context, albumID string) *ServiceError {
    return NewServiceError(ctx, "album_service", "get_album", "ALBUM_NOT_FOUND", "Album not found").
        WithContext("album_id", albumID).
        WithContext("search_type", "by_id")
}

func NewDatabaseError(ctx context.Context, service, operation string, cause error) *ServiceError {
    return NewServiceError(ctx, service, operation, "DATABASE_ERROR", "Database operation failed").
        WithCause(cause).
        WithContext("error_type", "database").
        WithContext("retryable", isDatabaseErrorRetryable(cause))
}

func NewFileSystemError(ctx context.Context, service, operation, filepath string, cause error) *ServiceError {
    return NewServiceError(ctx, service, operation, "FILESYSTEM_ERROR", "File system operation failed").
        WithCause(cause).
        WithContext("filepath", filepath).
        WithContext("error_type", "filesystem")
}
```

### Phase 4: Implementation Timeline

#### Week 1: Foundation
1. **Implement ServiceLogger infrastructure**
2. **Create instrumented datastore wrapper**
3. **Update AlbumService with comprehensive logging**

#### Week 2: Core Services
1. **Update MediaService with transaction logging**
2. **Enhance SyncService with detailed job logging**
3. **Add ProcessingService instrumentation**

#### Week 3: Error Enhancement
1. **Implement ServiceError types**
2. **Add error context to all service methods**
3. **Create error aggregation for monitoring**

#### Week 4: Integration & Testing
1. **Add performance metrics to logging**
2. **Create log analysis tools**
3. **Test error scenarios and logging output**

### Expected Debugging Improvements

#### Before (Current State):
```
ERROR: failed to create album
// That's it. No context, no details, no way to debug.
```

#### After (Improved State):
```json
{
  "level": "error",
  "timestamp": "2024-01-15T10:30:00Z",
  "message": "Operation failed",
  "service": "album_service",
  "operation": "create_album",
  "request_id": "req_abc123",
  "duration": "150ms",
  "params": {
    "album_id": "123e4567-e89b-12d3-a456-426614174000",
    "album_name": "Vacation Photos",
    "album_path": "/photos/2024/vacation",
    "parent_id": "parent_123"
  },
  "error_context": {
    "transaction_operation": "create_album_tx",
    "filesystem_operation": "create_folder",
    "folder_path": "/photos/2024/vacation",
    "error_type": "permission_denied"
  },
  "error": "mkdir /photos/2024/vacation: permission denied"
}
```

#### Log Correlation Example:
```
10:30:00.100 [req_abc123] INFO  album_service: Operation started operation=create_album album_path=/photos/2024/vacation
10:30:00.105 [req_abc123] DEBUG album_service: Checking if album already exists album_id=123e4567
10:30:00.110 [req_abc123] DEBUG datastore: Database query started operation=query_album
10:30:00.125 [req_abc123] DEBUG datastore: Database query completed operation=query_album duration=15ms found=false
10:30:00.130 [req_abc123] DEBUG album_service: Album does not exist, proceeding with creation
10:30:00.135 [req_abc123] DEBUG album_service: Starting database transaction for album creation
10:30:00.140 [req_abc123] DEBUG album_service: Writing album to database
10:30:00.155 [req_abc123] DEBUG album_service: Creating album folder on filesystem folder_path=/photos/2024/vacation
10:30:00.250 [req_abc123] ERROR album_service: Filesystem folder creation failed error="mkdir /photos/2024/vacation: permission denied"
10:30:00.255 [req_abc123] ERROR album_service: Operation failed operation=create_album duration=155ms
```

This comprehensive logging will make debugging **10x faster** by providing complete request context and detailed operation flow visibility.