package services

import (
	"context"
	"path"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

// JobGenerator generates SyncJobs based on folder structure discovery
type JobGenerator struct {
	albumSrv *AlbumService
	mediaSrv *MediaService
	fs       *fs.Datastore
}

// NewJobGenerator creates a new JobGenerator instance
func NewJobGenerator(albumSrv *AlbumService, mediaSrv *MediaService, fsDatastore *fs.Datastore) *JobGenerator {
	return &JobGenerator{
		albumSrv: albumSrv,
		mediaSrv: mediaSrv,
		fs:       fsDatastore,
	}
}

// Generate generates SyncJobs for a given path
// For root path: creates a SyncJob for each folder found at all levels
// For specific album: creates a SyncJob for that album and its content
func (g *JobGenerator) Generate(ctx context.Context, albumPath string) ([]*SyncAlbumJob, error) {
	// First, discover the complete folder structure
	tree, err := g.discoverFolderStructure(ctx, albumPath)
	if err != nil {
		return nil, err
	}

	// Generate SyncJobs from the discovered tree
	jobs := g.createJobsFromTree(ctx, tree, albumPath)
	
	return jobs, nil
}

// discoverFolderStructure discovers the complete folder structure using WalkTree
func (g *JobGenerator) discoverFolderStructure(ctx context.Context, albumPath string) (*entity.FolderNode, error) {
	// Use the WalkTree method to get the complete structure as a tree
	tree, err := g.fs.WalkTree(ctx, albumPath)
	if err != nil {
		return nil, err
	}
	return tree, nil
}

// createJobsFromTree processes the folder tree and creates SyncJobs for each folder
func (g *JobGenerator) createJobsFromTree(ctx context.Context, tree *entity.FolderNode, rootPath string) []*SyncAlbumJob {
	var jobs []*SyncAlbumJob

	// Process the tree recursively to create a job for each folder
	g.processNodeForJobs(ctx, tree, rootPath, &jobs)

	return jobs
}

// processNodeForJobs recursively processes a folder node and creates SyncJobs
func (g *JobGenerator) processNodeForJobs(ctx context.Context, node *entity.FolderNode, rootPath string, jobs *[]*SyncAlbumJob) {
	// For each folder (including the root), create a SyncJob that handles only that folder's content
	if node.Path != "" || rootPath == "" {
		// Create album entity for this folder
		album := entity.NewAlbum(node.Path)
		
		// Generate tasks for this specific folder (album creation + its direct media files)
		albumTasks, mediaTasks := g.createTasksForSingleFolder(node, album)
		
		// Create a SyncJob for this folder
		job, err := NewSyncJob(album.ID, album.Path, albumTasks, mediaTasks)
		if err == nil {
			*jobs = append(*jobs, job)
		}
	}

	// Recursively process children to create jobs for subfolders
	for _, child := range node.Children {
		g.processNodeForJobs(ctx, child, rootPath, jobs)
	}
}

// createTasksForSingleFolder creates tasks for a single folder (album creation + media processing)
func (g *JobGenerator) createTasksForSingleFolder(node *entity.FolderNode, album entity.Album) ([]Task[entity.Album], []Task[entity.Media]) {
	albumTasks := []Task[entity.Album]{}
	mediaTasks := []Task[entity.Media]{}

	// Create album creation task (only if not root)
	if node.Path != "" {
		// Determine parent album based on the node's parent
		var parentAlbum *entity.Album
		if node.Parent != nil && node.Parent.Path != "" {
			parentEntity := entity.NewAlbum(node.Parent.Path)
			parentAlbum = &parentEntity
		}
		
		albumTasks = append(albumTasks, g.createAlbumTaskWithParent(node.Path, parentAlbum))
	}

	// Create media processing tasks for all media files in this folder
	for _, mediaFilePath := range node.MediaFiles {
		mediaTasks = append(mediaTasks, g.createMediaTask(mediaFilePath, album))
	}

	return albumTasks, mediaTasks
}

// createAlbumTaskWithParent creates a task to create an album with proper parent relationship
func (g *JobGenerator) createAlbumTaskWithParent(albumPath string, parent *entity.Album) Task[entity.Album] {
	return func(ctx context.Context) entity.Result[entity.Album] {
		album := entity.NewAlbum(albumPath)
		if parent != nil {
			album.ParentId = &parent.ID
		}

		createdAlbum, err := g.albumSrv.CreateAlbum(ctx, album)
		if err != nil {
			return entity.NewResultWithError[entity.Album](err)
		}
		return entity.NewResult(*createdAlbum)
	}
}

// createMediaTask creates a task to process a media file
func (g *JobGenerator) createMediaTask(mediaFilePath string, album entity.Album) Task[entity.Media] {
	return func(ctx context.Context) entity.Result[entity.Media] {
		// Extract filename from full path
		filename := path.Base(mediaFilePath)

		// Create media entity
		media := entity.NewMedia(filename, album)

		// Get content function
		contentFn := g.mediaSrv.GetContentFn(ctx, media)
		media.Content = contentFn

		// Process the media
		createdMedia, err := g.mediaSrv.WriteMedia(ctx, media)
		if err != nil {
			return entity.NewResultWithError[entity.Media](err)
		}
		return entity.NewResult(*createdMedia)
	}
}