package services

// Logging field constants for consistent tracer.Success() calls across services
const (
	// Common fields
	AlbumID      = "album_id"
	AlbumPath    = "album_path"
	MediaID      = "media_id"
	Filename     = "filename"
	Filepath     = "filepath"
	JobID        = "job_id"
	Hash         = "hash"
	
	// Album service specific
	AlbumsReturned = "albums_returned"
	TotalAlbums    = "total_albums"
	StartIndex     = "start_index"
	EndIndex       = "end_index"
	WasExisting    = "was_existing"
	DescriptionUpdated = "description_updated"
	ThumbnailUpdated = "thumbnail_updated"
	FilesystemDeleted = "filesystem_deleted"
	DatabaseDeleted = "database_deleted"
	
	// Media service specific
	MediaReturned   = "media_returned"
	DateFiltered    = "date_filtered"
	ContentSize     = "content_size"
	ThumbnailSize   = "thumbnail_size"
	ExifFields      = "exif_fields"
	CapturedAt      = "captured_at"
	Skipped         = "skipped"
	Reason          = "reason"
	
	// Sync service specific
	Status         = "status"
	Total          = "total"
	Remaining      = "remaining"
	StoppedCount   = "stopped_count"
	RunningCount   = "running_count"
	ShutdownCompleted = "shutdown_completed"
)