package grpc

import (
	"io"
	"path"
	"strings"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// NewAlbum converts an entity.Album to a gRPC Album for API responses
func NewAlbum(album entity.Album, syncInProgress bool) *Album {
	_, name := path.Split(album.Path)

	grpcAlbum := &Album{
		Id:             album.ID,
		Path:           album.Path,
		Description:    album.Description,
		MediaCount:     int32(album.MediaCount),
		Name:           name,
		ParentId:       album.ParentId,
		Thumbnail:      album.Thumbnail,
		SyncInProgress: &syncInProgress,
	}

	// Convert children
	if len(album.Children) > 0 {
		children := make([]*AlbumChild, 0, len(album.Children))
		for _, childID := range album.Children {
			children = append(children, &AlbumChild{
				Id:   childID.ID,
				Name: childID.ID, // Using ID as name for now
			})
		}
		grpcAlbum.Children = children
	}

	// Convert media references
	if len(album.Media) > 0 {
		mediaIds := make([]string, 0, len(album.Media))
		for _, media := range album.Media {
			mediaIds = append(mediaIds, media.ID)
		}
		grpcAlbum.MediaIds = mediaIds
	}

	return grpcAlbum
}

// NewMedia converts an entity.Media to a gRPC Media for API responses
func NewMedia(media entity.Media) *Media {
	// Convert captured date to timestamp
	capturedAt := timestamppb.New(media.CapturedAt)

	// Convert EXIF data
	exifHeaders := make([]*ExifHeader, 0, len(media.Exif))
	for key, value := range media.Exif {
		exifHeaders = append(exifHeaders, &ExifHeader{
			Key:   key,
			Value: value,
		})
	}

	// Convert media type
	var mediaType MediaType
	switch strings.ToLower(string(media.MediaType)) {
	case "photo":
		mediaType = MediaType_MEDIA_TYPE_PHOTO
	case "video":
		mediaType = MediaType_MEDIA_TYPE_VIDEO
	default:
		mediaType = MediaType_MEDIA_TYPE_UNSPECIFIED
	}

	return &Media{
		Id:            media.ID,
		Filename:      media.Filename,
		AlbumId:       media.Album.ID,
		CapturedAt:    capturedAt,
		Type:          mediaType,
		Exif:          exifHeaders,
		ThumbnailData: media.Thumbnail,
	}
}

// Entity converts a gRPC CreateAlbumRequest to an entity.Album for business logic processing
func (r *CreateAlbumRequest) Entity() entity.Album {
	album := entity.Album{
		ID:          entity.GenerateId(r.Name),
		Path:        r.Name,
		Description: r.Description,
		Children:    []entity.Album{},
		Media:       []entity.Media{},
		CreatedAt:   time.Now(),
	}

	if r.ParentId != nil {
		album.ParentId = r.ParentId
	}

	return album
}

// ApplyTo applies a gRPC UpdateAlbumRequest to an existing entity.Album
func (r *UpdateAlbumRequest) ApplyTo(album *entity.Album) {
	if r.Description != nil {
		album.Description = r.Description
	}
	if r.Thumbnail != nil {
		album.Thumbnail = r.Thumbnail
	}
}

// ApplyTo applies a gRPC UpdateMediaRequest to an existing entity.Media
func (r *UpdateMediaRequest) ApplyTo(media *entity.Media) {
	if r.CapturedAt != nil {
		media.CapturedAt = r.CapturedAt.AsTime()
	}
	if len(r.Exif) > 0 {
		if media.Exif == nil {
			media.Exif = make(map[string]string)
		}
		for _, exif := range r.Exif {
			media.Exif[exif.Key] = exif.Value
		}
	}
}

// ToMediaEntity converts upload request data to an entity.Media for business logic processing
func ToMediaEntity(filename, albumId string, fileContent io.Reader, album entity.Album) entity.Media {
	// Create new media using the entity constructor
	media := entity.NewMedia(filename, album)

	// Set the content function to return the file reader
	media.Content = func() (io.Reader, error) {
		return fileContent, nil
	}

	// Set current time as captured time (can be updated later from EXIF)
	media.CapturedAt = time.Now()

	return media
}

// ConvertJobStatusToAPI converts internal job status to gRPC status
func ConvertJobStatusToAPI(status entity.JobStatus) SyncJobStatus {
	switch status {
	case entity.StatusPending:
		return SyncJobStatus_SYNC_JOB_STATUS_PENDING
	case entity.StatusRunning:
		return SyncJobStatus_SYNC_JOB_STATUS_RUNNING
	case entity.StatusCompleted:
		return SyncJobStatus_SYNC_JOB_STATUS_COMPLETED
	case entity.StatusFailed:
		return SyncJobStatus_SYNC_JOB_STATUS_FAILED
	case entity.StatusStopped:
		return SyncJobStatus_SYNC_JOB_STATUS_STOPPED
	default:
		return SyncJobStatus_SYNC_JOB_STATUS_UNSPECIFIED
	}
}

// ConvertJobResultsToTaskResults converts internal job results to gRPC task results
func ConvertJobResultsToTaskResults(jobResults []entity.JobResult) []*TaskResult {
	taskResults := make([]*TaskResult, 0, len(jobResults))

	for _, jobResult := range jobResults {
		taskResult := ConvertJobResultToTaskResult(jobResult)
		taskResults = append(taskResults, taskResult)
	}

	return taskResults
}

// ConvertJobResultToTaskResult converts a single JobResult to gRPC TaskResult
func ConvertJobResultToTaskResult(jobResult entity.JobResult) *TaskResult {
	// Calculate duration in milliseconds
	duration := int32(jobResult.CompletedAt.Sub(jobResult.StartedAt).Milliseconds())

	// Extract item name and determine type from the result string
	item, itemType := extractItemAndType(jobResult.Result)

	// Create the task result
	taskResult := &TaskResult{
		Item:     item,
		ItemType: itemType,
		Duration: duration,
	}

	// Set result based on error status
	if jobResult.Err != nil {
		// Error case - set error message
		taskResult.Result = &TaskResultStatus{
			Result: &TaskResultStatus_Error{
				Error: jobResult.Err.Error(),
			},
		}
	} else {
		// Success case
		taskResult.Result = &TaskResultStatus{
			Result: &TaskResultStatus_Success{
				Success: "ok",
			},
		}
	}

	return taskResult
}

// extractItemAndType extracts the item name and determines if it's a file or folder
func extractItemAndType(resultStr string) (string, TaskResultItemType) {
	// Parse the result string to extract item name
	if strings.HasPrefix(resultStr, "Album ") && strings.HasSuffix(resultStr, " processed") {
		// Extract album path: "Album photos/2023 processed" -> "photos/2023"
		item := strings.TrimPrefix(resultStr, "Album ")
		item = strings.TrimSuffix(item, " processed")
		return item, TaskResultItemType_TASK_RESULT_ITEM_TYPE_FOLDER
	} else if strings.HasPrefix(resultStr, "Media ") && strings.HasSuffix(resultStr, " processed") {
		// Extract media path: "Media photos/2023/IMG_001.jpg processed" -> "photos/2023/IMG_001.jpg"
		item := strings.TrimPrefix(resultStr, "Media ")
		item = strings.TrimSuffix(item, " processed")
		return item, TaskResultItemType_TASK_RESULT_ITEM_TYPE_FILE
	}

	// Fallback - try to determine from the string content
	if strings.Contains(resultStr, ".") {
		// Likely a file if it contains a dot (extension)
		return resultStr, TaskResultItemType_TASK_RESULT_ITEM_TYPE_FILE
	}

	// Default to folder
	return resultStr, TaskResultItemType_TASK_RESULT_ITEM_TYPE_FOLDER
}

// NewSyncJob converts an internal JobProgress to a gRPC SyncJob
func NewSyncJob(job entity.JobProgress) *SyncJob {
	grpcJob := &SyncJob{
		Id:             job.Id.String(),
		Status:         ConvertJobStatusToAPI(job.Status),
		RemainingTasks: int32(job.Remaining),
		TotalTasks:     int32(job.Total),
		CompletedTasks: ConvertJobResultsToTaskResults(job.Results),
		CreatedAt:      timestamppb.New(job.CreatedAt),
		Path:           job.Path,
	}

	if job.StartedAt != nil {
		grpcJob.StartedAt = timestamppb.New(*job.StartedAt)
	}

	if job.CompletedAt != nil {
		grpcJob.FinishedAt = timestamppb.New(*job.CompletedAt)
	}

	// Calculate duration in seconds
	if job.StartedAt != nil {
		var durationSeconds int32
		if job.CompletedAt != nil {
			// Job completed - calculate total duration
			durationSeconds = int32(job.CompletedAt.Sub(*job.StartedAt).Seconds())
		} else {
			// Job still running - calculate elapsed time
			durationSeconds = int32(time.Now().Sub(*job.StartedAt).Seconds())
		}
		grpcJob.Duration = &durationSeconds
	}

	// Set message if available (provides additional info about job status/reason)
	if job.Reason != "" {
		grpcJob.Error = &job.Reason
	}

	// Note: RemainingTime calculation would need to be added to JobProgress struct
	// For now, we'll skip this field

	return grpcJob
}
