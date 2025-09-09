package http

import (
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
)

// NewAlbum converts an entity.Album to a v1.Album for API responses
func NewAlbum(album entity.Album) Album {
	apiAlbum := Album{
		Id:          album.ID,
		Path:        album.Path,
		Description: album.Description,
		Href:        "/api/v1/albums/" + album.ID,
		MediaCount:  album.MediaCount,
	}

	_, name := path.Split(album.Path)
	apiAlbum.Name = name

	// Convert children
	if len(album.Children) > 0 {
		children := make([]struct {
			Href string `json:"href"`
			Name string `json:"name"`
		}, 0, len(album.Children))

		for _, childID := range album.Children {
			children = append(children, struct {
				Href string `json:"href"`
				Name string `json:"name"`
			}{
				Href: "/api/v1/albums/" + childID.ID,
				Name: childID.ID, // Using ID as name for now
			})
		}
		apiAlbum.Children = &children
	}

	// Convert media references
	if len(album.Media) > 0 {
		mediaHrefs := make([]string, 0, len(album.Media))
		for _, media := range album.Media {
			mediaHrefs = append(mediaHrefs, "/api/v1/media/"+media.ID)
		}
		apiAlbum.Media = &mediaHrefs
	}

	// Set parent href if parent exists
	if album.ParentId != nil {
		parentHref := "/api/v1/albums/" + *album.ParentId
		apiAlbum.ParentHref = &parentHref
	}

	if album.Thumbnail != nil {
		thumbnail := fmt.Sprintf("/api/v1/media/%s/thumbnail", *album.Thumbnail)
		apiAlbum.Thumbnail = &thumbnail
	}

	return apiAlbum
}

// NewMedia converts an entity.Media to a v1.Media for API responses
func NewMedia(media entity.Media) Media {
	// Convert EXIF data
	exifHeaders := make([]ExifHeader, 0, len(media.Exif))
	for key, value := range media.Exif {
		exifHeaders = append(exifHeaders, ExifHeader{
			Key:   key,
			Value: value,
		})
	}

	return Media{
		Id:         media.ID,
		Filename:   media.Filename,
		AlbumHref:  "/api/v1/albums/" + media.Album.ID,
		CapturedAt: media.CapturedAt,
		Type:       string(media.MediaType),
		Content:    "/api/v1/media/" + media.ID + "/content",
		Thumbnail:  "/api/v1/media/" + media.ID + "/thumbnail",
		Href:       "/api/v1/media/" + media.ID,
		Exif:       exifHeaders,
	}
}

// Entity converts a v1.CreateAlbumRequest to an entity.Album for business logic processing.
// This method transforms the HTTP request data into the internal domain model representation.
func (r CreateAlbumRequest) Entity() entity.Album {
	album := entity.Album{
		ID:          entity.GenerateId(r.Name),
		Path:        r.Name,
		ParentId:    r.ParentId,
		Description: r.Description,
		Children:    []entity.Album{},
		Media:       []entity.Media{},
		CreatedAt:   time.Now(),
	}

	return album
}

// Entity converts a v1.UpdateAlbumRequest to an entity.Album for business logic processing.
// This method applies the updates to an existing album entity.
func (r UpdateAlbumRequest) ApplyTo(album *entity.Album) {
	album.Description = r.Description
	album.Thumbnail = r.Thumbnail
}

// Entity converts a v1.UpdateMediaRequest to updates for an entity.Media.
// This method applies the updates to an existing media entity.
func (r UpdateMediaRequest) ApplyTo(media *entity.Media) {
	if r.CapturedAt != nil {
		media.CapturedAt = r.CapturedAt.Time
	}
	if r.Exif != nil {
		if media.Exif == nil {
			media.Exif = make(map[string]string)
		}
		for _, exif := range *r.Exif {
			media.Exif[exif.Key] = exif.Value
		}
	}
}

// ToMediaEntity converts upload request data to an entity.Media for business logic processing.
// This method transforms the multipart form data into the internal domain model representation.
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

// ConvertJobStatusToAPI converts internal job status to API status
func ConvertJobStatusToAPI(status services.JobStatus) SyncJobStatus {
	switch status {
	case services.StatusPending:
		return Pending
	case services.StatusRunning:
		return Running
	case services.StatusCompleted:
		return Completed
	case services.StatusFailed:
		return Failed
	case services.StatusStopped:
		return Stopped
	default:
		return Pending
	}
}

// ConvertJobResultsToTaskResults converts internal job results to API task results
func ConvertJobResultsToTaskResults(jobResults []services.JobResult) []TaskResult {
	taskResults := make([]TaskResult, 0, len(jobResults))

	for _, jobResult := range jobResults {
		taskResult := ConvertJobResultToTaskResult(jobResult)
		taskResults = append(taskResults, taskResult)
	}

	return taskResults
}

// ConvertJobResultToTaskResult converts a single JobResult to TaskResult
func ConvertJobResultToTaskResult(jobResult services.JobResult) TaskResult {
	// Calculate duration in seconds
	duration := int(jobResult.CompletedAt.Sub(jobResult.StartedAt).Seconds())

	// Extract item name and determine type from the result string
	// JobResult.Result format: "Album photos/2023 processed" or "Media photos/2023/IMG_001.jpg processed"
	item, itemType := extractItemAndType(jobResult.Result)

	// Create the task result
	taskResult := TaskResult{
		Item:     item,
		ItemType: itemType,
		Duration: duration,
	}

	// Set result based on error status
	if jobResult.Err != nil {
		// Error case - set error message
		var result TaskResult_Result
		result.FromTaskResultResult1(jobResult.Err.Error())
		taskResult.Result = result
	} else {
		// Success case
		var result TaskResult_Result
		result.FromTaskResultResult0(Ok)
		taskResult.Result = result
	}

	return taskResult
}

// extractItemAndType extracts the item name and determines if it's a file or folder
// from JobResult.Result strings like "Album photos/2023 processed" or "Media photos/2023/IMG_001.jpg processed"
func extractItemAndType(resultStr string) (string, TaskResultItemType) {
	// Parse the result string to extract item name
	if strings.HasPrefix(resultStr, "Album ") && strings.HasSuffix(resultStr, " processed") {
		// Extract album path: "Album photos/2023 processed" -> "photos/2023"
		item := strings.TrimPrefix(resultStr, "Album ")
		item = strings.TrimSuffix(item, " processed")
		return item, Folder
	} else if strings.HasPrefix(resultStr, "Media ") && strings.HasSuffix(resultStr, " processed") {
		// Extract media path: "Media photos/2023/IMG_001.jpg processed" -> "photos/2023/IMG_001.jpg"
		item := strings.TrimPrefix(resultStr, "Media ")
		item = strings.TrimSuffix(item, " processed")
		return item, File
	}

	// Fallback - try to determine from the string content
	if strings.Contains(resultStr, ".") {
		// Likely a file if it contains a dot (extension)
		return resultStr, File
	}

	// Default to folder
	return resultStr, Folder
}

// NewSyncJob converts a services.JobProgress to a v1.SyncJob for API responses
func NewSyncJob(jobProgress services.JobProgress) SyncJob {
	// Convert job results to API task results
	taskResults := ConvertJobResultsToTaskResults(jobProgress.Results)

	apiJob := SyncJob{
		Id:             jobProgress.Id.String(),
		Status:         ConvertJobStatusToAPI(jobProgress.Status),
		RemainingTasks: jobProgress.Remaining,
		TotalTasks:     jobProgress.Total,
		CompletedTasks: taskResults,
		CreatedAt:      jobProgress.CreatedAt,
		Path:           jobProgress.Path,
	}

	// Set timing fields based on job status
	if jobProgress.StartedAt != nil {
		apiJob.StartedAt = jobProgress.StartedAt
	} else {
		apiJob.StartedAt = &jobProgress.CreatedAt // Use created time if not started yet
	}

	if jobProgress.CompletedAt != nil {
		apiJob.FinishedAt = jobProgress.CompletedAt
	}
	// Don't set FinishedAt if job hasn't completed - let it be omitted

	// Calculate duration in seconds
	if jobProgress.StartedAt != nil {
		var durationSeconds int
		if jobProgress.CompletedAt != nil {
			// Job completed - calculate total duration
			durationSeconds = int(jobProgress.CompletedAt.Sub(*jobProgress.StartedAt).Seconds())
		} else {
			// Job still running - calculate elapsed time
			durationSeconds = int(time.Now().Sub(*jobProgress.StartedAt).Seconds())
		}
		apiJob.Duration = &durationSeconds
	}

	// Set error message if job failed
	if jobProgress.Err != nil {
		errorMessage := jobProgress.Err.Error()
		apiJob.Error = &errorMessage
	}

	return apiJob
}
