package services

import (
	"context"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

type Task[R any] func(ctx context.Context) entity.Result[R]

// SyncAlbumJob represents a sync job that processes files in a directory
type SyncAlbumJob struct {
	ID            uuid.UUID
	rootAlbum     entity.Album
	mediaSrv      *MediaService
	albumSrv      *AlbumService
	fs            *fs.Datastore
	progressMeter *jobProgressMeter
	doneCh        chan bool
}

// NewJob creates a new sync job
func NewSyncJob(rootAlbum entity.Album, albumService *AlbumService, mediaService *MediaService, fsDatastore *fs.Datastore) (*SyncAlbumJob, error) {
	job := &SyncAlbumJob{
		ID:            uuid.New(),
		mediaSrv:      mediaService,
		albumSrv:      albumService,
		fs:            fsDatastore,
		rootAlbum:     rootAlbum,
		progressMeter: newJobProgressMeter(),
	}

	return job, nil
}

// Start begins the job execution
func (j *SyncAlbumJob) Start(ctx context.Context) error {
	j.doneCh = make(chan bool)
	j.progressMeter.Start()

	discoveryTask := j.createFolderStructureDiscoveryTask()

	zap.S().Debugw("starting folder structure discovery", "job_id", j.ID)

	result := discoveryTask(ctx)
	if result.Err != nil {
		j.progressMeter.Failed()
		return result.Err
	}

	defer func() {
		j.progressMeter.Stop()
	}()

	albumTasks, mediaTasks := j.createTasks(result.Data)
	j.progressMeter.Total(len(albumTasks) + len(mediaTasks))

	zap.S().Debugw("creating albums", "job_id", j.ID, "tasks_count", len(albumTasks))

	for _, t := range albumTasks {
		start := time.Now()
		r := t(ctx)
		if r.Err != nil {
			zap.S().Warnw("failed to create album", "job_id", j.ID, "error", r.Err)
		}
		end := time.Now()

		j.progressMeter.Result(JobResult{
			Result:      fmt.Sprintf("Album %s processed", r.Data.Path),
			Err:         r.Err,
			StartedAt:   start,
			CompletedAt: end,
		})

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-j.doneCh:
			j.doneCh <- true
			return nil
		default:
		}
	}

	zap.S().Debugw("processing media", "job_id", j.ID, "tasks_count", len(mediaTasks))
	for _, t := range mediaTasks {
		start := time.Now()
		r := t(ctx)
		if r.Err != nil {
			zap.S().Warnw("failed to process media", "job_id", j.ID, "error", r.Err)
		}
		end := time.Now()

		j.progressMeter.Result(JobResult{
			Result:      fmt.Sprintf("Media %s processed", r.Data.Filepath()),
			Err:         r.Err,
			StartedAt:   start,
			CompletedAt: end,
		})

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-j.doneCh:
			j.doneCh <- true
			return nil
		default:
		}
	}

	return nil
}

// GetId returns the job's unique identifier
func (j *SyncAlbumJob) GetID() uuid.UUID {
	return j.ID
}

// Stop cancels the job execution
func (j *SyncAlbumJob) Stop() error {
	if j.doneCh == nil {
		return nil
	}
	j.doneCh <- true
	<-j.doneCh

	j.progressMeter.Stop()

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
func (j *SyncAlbumJob) createTasks(tree *entity.FolderNode) ([]Task[entity.Album], []Task[entity.Media]) {
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

func (j *SyncAlbumJob) Status() JobProgress {
	p := j.progressMeter.Status()
	p.Id = j.ID
	return p
}

// JobStatus represents the current status of a job
type JobStatus string

const (
	StatusPending   JobStatus = "pending"
	StatusRunning   JobStatus = "running"
	StatusCompleted JobStatus = "completed"
	StatusFailed    JobStatus = "failed"
	StatusStopped   JobStatus = "stopped"
)

// JobProgress tracks the progress of a job
type JobProgress struct {
	Id          uuid.UUID
	Status      JobStatus
	CreatedAt   time.Time
	StartedAt   *time.Time
	CompletedAt *time.Time
	Total       int
	Remaining   int
	Results     []JobResult
}

type JobResult struct {
	Result      string
	Err         error
	StartedAt   time.Time
	CompletedAt time.Time
}

type jobProgressMeter struct {
	createdAt   time.Time
	startedAt   *time.Time
	completedAt *time.Time
	status      JobStatus
	total       int
	remaining   int
	results     []JobResult
	mu          sync.Mutex
}

func newJobProgressMeter() *jobProgressMeter {
	return &jobProgressMeter{
		createdAt: time.Now(),
		status:    StatusPending,
		results:   []JobResult{},
	}
}

func (j *jobProgressMeter) Start() {
	now := time.Now()
	j.startedAt = &now
	j.status = StatusRunning
}

func (j *jobProgressMeter) Total(total int) {
	j.total, j.remaining = total, total
}

func (j *jobProgressMeter) Stop() {
	now := time.Now()
	j.completedAt = &now
	j.status = StatusStopped
}

func (j *jobProgressMeter) Failed() {
	now := time.Now()
	j.completedAt = &now
	j.status = StatusFailed
}

func (j *jobProgressMeter) Result(r JobResult) {
	j.remaining--
	j.results = append(j.results, r)
}

func (j *jobProgressMeter) Status() JobProgress {
	j.mu.Lock()
	defer j.mu.Unlock()

	p := JobProgress{
		Status:      j.status,
		CreatedAt:   j.createdAt,
		Total:       j.total,
		Remaining:   j.remaining,
		StartedAt:   j.startedAt,
		CompletedAt: j.completedAt,
		Results:     j.results,
	}

	return p
}
