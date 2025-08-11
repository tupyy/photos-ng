package job

import (
	"context"
	"path"
	"sync"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Task[R any] func(ctx context.Context) entity.Result[R]

// FileInfo represents information about a file to be processed
type FileInfo struct {
	Path     string // Relative path from the root
	FullPath string // Absolute path on filesystem
	Name     string // Filename
	Size     int64  // File size in bytes
}

// SyncAlbumJob represents a sync job that processes files in a directory
type SyncAlbumJob struct {
	ID          uuid.UUID
	createdAt   time.Time
	startedAt   *time.Time
	completedAt *time.Time
	rootAlbum   entity.Album
	mediaSrv    *services.MediaService
	albumSrv    *services.AlbumService
	fs          *fs.Datastore
	status      entity.JobStatus
	total       int
	remaining   int
	completed   []entity.TaskResult
	mu          sync.Mutex
}

// NewJob creates a new sync job
func NewSyncJob(rootAlbum entity.Album, albumService *services.AlbumService, mediaService *services.MediaService, fsDatastore *fs.Datastore) (*SyncAlbumJob, error) {
	job := &SyncAlbumJob{
		ID:        uuid.New(),
		createdAt: time.Now(),
		mediaSrv:  mediaService,
		albumSrv:  albumService,
		fs:        fsDatastore,
		rootAlbum: rootAlbum,
		status:    entity.StatusPending,
		completed: []entity.TaskResult{},
	}

	return job, nil
}

// Start begins the job execution
func (j *SyncAlbumJob) Start(ctx context.Context) error {
	j.status = entity.StatusRunning
	now := time.Now()
	j.startedAt = &now

	structureDiscoveryTask := j.createFolderStructureDiscoveryTask()

	zap.S().Debugw("starting folder structure discovery", "job_id", j.ID)

	result := structureDiscoveryTask(ctx)
	if result.Err != nil {
		j.status = entity.StatusFailed
		return result.Err
	}

	albumTasks, mediaTasks := j.createTasksFromTree(result.Data)
	j.total = len(albumTasks) + len(mediaTasks)
	j.remaining = j.total

	zap.S().Debugw("creating albums", "job_id", j.ID, "tasks_count", len(albumTasks))

	for _, t := range albumTasks {
		r := t(ctx)
		if r.Err != nil {
			zap.S().Warnw("failed to create album", "job_id", j.ID, "error", r.Err)
		}

		j.completed = append(j.completed, entity.TaskResult{
			ItemType: "album",
			Name:     r.Data.Path,
			Err:      r.Err,
		})

		j.remaining--

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	zap.S().Debugw("processing media", "job_id", j.ID, "tasks_count", len(mediaTasks))
	for _, t := range mediaTasks {
		r := t(ctx)
		if r.Err != nil {
			zap.S().Warnw("failed to process media", "job_id", j.ID, "error", r.Err)
		}

		j.remaining--

		j.completed = append(j.completed, entity.TaskResult{
			ItemType: "media",
			Name:     r.Data.Filename,
			Err:      r.Err,
		})

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	j.status = entity.StatusCompleted
	now = time.Now()
	j.completedAt = &now
	return nil
}

// Stop cancels the job execution
func (j *SyncAlbumJob) Stop() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.status == entity.StatusRunning {
		j.status = entity.StatusStopped
		now := time.Now()
		j.completedAt = &now
	}
	zap.S().Infow("Job stopped", "id", j.ID)
	return nil
}

// createFolderStructureDiscoveryTask creates a task that discovers the complete folder structure
// using WalkTree to return a hierarchical tree of folders and media files
func (j *SyncAlbumJob) createFolderStructureDiscoveryTask() Task[*entity.FolderNode] {
	return func(ctx context.Context) entity.Result[*entity.FolderNode] {
		// Use the new WalkTree method to get the complete structure as a tree
		tree, err := j.fs.WalkTree(ctx, j.rootAlbum.Path)
		if err != nil {
			return entity.NewResultWithError[*entity.FolderNode](err)
		}

		return entity.NewResult(tree)
	}
}

// createTasksFromTree processes the folder tree and creates album and media tasks
func (j *SyncAlbumJob) createTasksFromTree(tree *entity.FolderNode) ([]Task[entity.Album], []Task[entity.Media]) {
	albumTasks := []Task[entity.Album]{}
	mediaTasks := []Task[entity.Media]{}

	// Create a map to track created albums by path for parent lookup
	albumMap := make(map[string]*entity.Album)
	albumMap[j.rootAlbum.Path] = &j.rootAlbum

	// Process the tree recursively to maintain parent-child relationships
	j.processNodeForTasks(tree, albumMap, &albumTasks, &mediaTasks)

	return albumTasks, mediaTasks
}

// processNodeForTasks recursively processes a folder node and its children
func (j *SyncAlbumJob) processNodeForTasks(node *entity.FolderNode, albumMap map[string]*entity.Album, albumTasks *[]Task[entity.Album], mediaTasks *[]Task[entity.Media]) {
	// Handle root node logic
	if node.Path == j.rootAlbum.Path {
		// Root data folder never has media files - only process if root path is not empty
		if j.rootAlbum.Path != "" {
			// Non-empty root album - process media files
			for _, mediaFilePath := range node.MediaFiles {
				*mediaTasks = append(*mediaTasks, j.createMediaTask(mediaFilePath, j.rootAlbum))
			}
		}
		// Empty root path (data folder itself) - no media files to process by design
	} else {
		// Non-root node - create album
		var parentAlbum *entity.Album
		if node.Parent != nil {
			// If parent is root with empty path, parent album is nil (top-level album)
			if node.Parent.Path == "" {
				parentAlbum = nil
			} else {
				parentAlbum = albumMap[node.Parent.Path]
			}
		}

		// Create album task for this folder
		*albumTasks = append(*albumTasks, j.createAlbumTaskWithParent(node.Path, parentAlbum))

		// Add this album to the map for its children to reference
		albumEntity := entity.NewAlbum(node.Path)
		if parentAlbum != nil {
			albumEntity.ParentId = &parentAlbum.ID
		}
		albumMap[node.Path] = &albumEntity

		// Process media files in this folder
		for _, mediaFilePath := range node.MediaFiles {
			*mediaTasks = append(*mediaTasks, j.createMediaTask(mediaFilePath, albumEntity))
		}
	}

	// Recursively process children
	for _, child := range node.Children {
		j.processNodeForTasks(child, albumMap, albumTasks, mediaTasks)
	}
}

// createAlbumTaskWithParent creates a task to create an album with proper parent relationship
func (j *SyncAlbumJob) createAlbumTaskWithParent(albumPath string, parent *entity.Album) Task[entity.Album] {
	return func(ctx context.Context) entity.Result[entity.Album] {
		album := entity.NewAlbum(albumPath)
		if parent != nil {
			album.ParentId = &parent.ID
		}

		createdAlbum, err := j.albumSrv.CreateAlbum(ctx, album)
		if err != nil {
			return entity.NewResultWithError[entity.Album](err)
		}
		return entity.NewResult(*createdAlbum)
	}
}

// createMediaTask creates a task to process a media file
func (j *SyncAlbumJob) createMediaTask(mediaFilePath string, album entity.Album) Task[entity.Media] {
	return func(ctx context.Context) entity.Result[entity.Media] {
		// Extract filename from full path
		filename := path.Base(mediaFilePath)

		// Create media entity
		media := entity.NewMedia(filename, album)

		// Get content function
		contentFn := j.mediaSrv.GetContentFn(ctx, media)
		media.Content = contentFn

		// Process the media
		createdMedia, err := j.mediaSrv.WriteMedia(ctx, media)
		if err != nil {
			return entity.NewResultWithError[entity.Media](err)
		}
		return entity.NewResult(*createdMedia)
	}
}

func (j *SyncAlbumJob) Status() entity.JobProgress {
	j.mu.Lock()
	defer j.mu.Unlock()
	progress := entity.JobProgress{
		Id:          j.ID,
		Status:      j.status,
		CreatedAt:   j.createdAt,
		StartedAt:   j.startedAt,
		CompletedAt: j.completedAt,
		Total:       j.total,
		Remaining:   j.remaining,
		Completed:   j.completed,
	}

	return progress
}
