package grpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	v1grpc "git.tls.tupangiu.ro/cosmin/photos-ng/api/v1/grpc"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Handler implements the gRPC PhotosNGService
type Handler struct {
	v1grpc.UnimplementedPhotosNGServiceServer
	albumSrv *services.AlbumService
	mediaSrv *services.MediaService
	statsSrv *services.StatsService
	syncSrv  *services.SyncService
}

// NewHandler creates a new gRPC server implementation
func NewHandler(dt *pg.Datastore, fsDatastore *fs.Datastore) *Handler {
	albumSrv := services.NewAlbumService(dt, fsDatastore)
	mediaSrv := services.NewMediaService(dt, fsDatastore)
	syncSrv := services.NewSyncService(albumSrv, mediaSrv, fsDatastore)

	return &Handler{
		albumSrv: albumSrv,
		mediaSrv: mediaSrv,
		statsSrv: services.NewStatsService(dt),
		syncSrv:  syncSrv,
	}
}

// Album operations implementation
func (s *Handler) ListAlbums(ctx context.Context, req *v1grpc.ListAlbumsRequest) (*v1grpc.ListAlbumsResponse, error) {
	// Set default values for pagination
	limit := 20
	if req.Pagination != nil && req.Pagination.Limit > 0 {
		limit = int(req.Pagination.Limit)
	}
	offset := 0
	if req.Pagination != nil && req.Pagination.Offset >= 0 {
		offset = int(req.Pagination.Offset)
	}

	hasParent := false
	if req.WithParent != nil {
		hasParent = *req.WithParent
	}

	// Create album service options
	opts := services.NewListOptionsWithOptions(
		services.WithLimit(limit),
		services.WithOffset(offset),
		services.WithHasParent(hasParent),
	)

	albums, err := s.albumSrv.List(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Get total count for pagination (without limit/offset)
	totalOpts := services.NewListOptionsWithOptions()
	allAlbums, err := s.albumSrv.List(ctx, totalOpts)
	if err != nil {
		return nil, err
	}
	total := len(allAlbums)

	// Convert entity albums to gRPC albums
	grpcAlbums := make([]*v1grpc.Album, 0, len(albums))
	for _, album := range albums {
		// Check if album has sync job in progress
		syncInProgress := s.syncSrv.IsAlbumSyncing(album.ID)
		grpcAlbums = append(grpcAlbums, v1grpc.NewAlbum(album, syncInProgress))
	}

	return &v1grpc.ListAlbumsResponse{
		Albums: grpcAlbums,
		Pagination: &v1grpc.PaginationResponse{
			Total:  int32(total),
			Limit:  int32(limit),
			Offset: int32(offset),
		},
	}, nil
}

func (s *Handler) GetAlbum(ctx context.Context, req *v1grpc.GetAlbumRequest) (*v1grpc.Album, error) {
	album, err := s.albumSrv.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// Check if album has sync job in progress
	syncInProgress := s.syncSrv.IsAlbumSyncing(album.ID)
	return v1grpc.NewAlbum(*album, syncInProgress), nil
}

func (s *Handler) CreateAlbum(ctx context.Context, req *v1grpc.CreateAlbumRequest) (*v1grpc.Album, error) {
	// Convert gRPC request to entity
	album := req.Entity()

	// Create album using service
	createdAlbum, err := s.albumSrv.Create(ctx, album)
	if err != nil {
		return nil, err
	}

	// Check if album has sync job in progress
	syncInProgress := s.syncSrv.IsAlbumSyncing(createdAlbum.ID)
	return v1grpc.NewAlbum(*createdAlbum, syncInProgress), nil
}

func (s *Handler) UpdateAlbum(ctx context.Context, req *v1grpc.UpdateAlbumByIdRequest) (*v1grpc.Album, error) {
	// Get existing album
	album, err := s.albumSrv.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	req.Update.ApplyTo(album)

	// Update album using service
	updatedAlbum, err := s.albumSrv.Update(ctx, *album)
	if err != nil {
		return nil, err
	}

	// Check if album has sync job in progress
	syncInProgress := s.syncSrv.IsAlbumSyncing(updatedAlbum.ID)
	return v1grpc.NewAlbum(*updatedAlbum, syncInProgress), nil
}

func (s *Handler) DeleteAlbum(ctx context.Context, req *v1grpc.DeleteAlbumRequest) (*emptypb.Empty, error) {
	err := s.albumSrv.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Handler) SyncAlbum(ctx context.Context, req *v1grpc.SyncAlbumRequest) (*v1grpc.SyncAlbumResponse, error) {
	// Note: SyncAlbum is handled via StartSync in this implementation
	_, err := s.syncSrv.StartSync(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &v1grpc.SyncAlbumResponse{
		Message:     "Album sync started",
		SyncedItems: 0, // Will be populated as job progresses
	}, nil
}

// Media operations implementation
func (s *Handler) ListMedia(req *v1grpc.ListMediaRequest, stream v1grpc.PhotosNGService_ListMediaServer) error {
	// Build media options from parameters
	opt := &services.MediaOptions{}

	// Set default values for pagination
	limit := 50 // Default batch size for streaming
	if req.Pagination != nil && req.Pagination.Limit > 0 {
		opt.MediaLimit = int(req.Pagination.Limit)
		limit = int(req.Pagination.Limit)
	} else {
		opt.MediaLimit = limit
	}

	// Parse cursor if provided
	if req.Pagination != nil && req.Pagination.Cursor != nil {
		cursor, err := services.DecodeCursor(*req.Pagination.Cursor)
		if err != nil {
			return err
		}
		opt.Cursor = cursor
	}

	// Parse direction parameter (default to "forward")
	if req.Pagination != nil && req.Pagination.Direction != nil {
		opt.Direction = *req.Pagination.Direction
	} else {
		opt.Direction = "forward"
	}

	// Add album filter
	if req.AlbumId != nil {
		opt.AlbumID = req.AlbumId
	}

	// Add media type filter
	if req.Type != nil && *req.Type != v1grpc.MediaType_MEDIA_TYPE_UNSPECIFIED {
		var mediaType string
		switch *req.Type {
		case v1grpc.MediaType_MEDIA_TYPE_PHOTO:
			mediaType = "photo"
		case v1grpc.MediaType_MEDIA_TYPE_VIDEO:
			mediaType = "video"
		}
		opt.MediaType = &mediaType
	}

	// Note: Sorting is fixed to captured_at DESC, id DESC for cursor pagination

	// Fetch media using the service
	mediaItems, nextCursor, err := s.mediaSrv.List(stream.Context(), opt)
	if err != nil {
		return err
	}

	// Stream the media items
	for _, mediaItem := range mediaItems {
		grpcMedia := v1grpc.NewMedia(mediaItem)
		if err := stream.Send(grpcMedia); err != nil {
			return err
		}
	}

	// Send next cursor in trailer if available
	if nextCursor != nil {
		encodedCursor, err := nextCursor.Encode()
		if err == nil {
			stream.SetTrailer(metadata.Pairs("next-cursor", encodedCursor))
		}
	}

	return nil
}

func (s *Handler) GetMedia(ctx context.Context, req *v1grpc.GetMediaRequest) (*v1grpc.Media, error) {
	media, err := s.mediaSrv.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return v1grpc.NewMedia(*media), nil
}

func (s *Handler) UpdateMedia(ctx context.Context, req *v1grpc.UpdateMediaByIdRequest) (*v1grpc.Media, error) {
	// Get existing media
	media, err := s.mediaSrv.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	req.Update.ApplyTo(media)

	// Update media using service
	updatedMedia, err := s.mediaSrv.Update(ctx, *media)
	if err != nil {
		return nil, err
	}

	return v1grpc.NewMedia(*updatedMedia), nil
}

func (s *Handler) DeleteMedia(ctx context.Context, req *v1grpc.DeleteMediaRequest) (*emptypb.Empty, error) {
	err := s.mediaSrv.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Handler) GetMediaThumbnail(ctx context.Context, req *v1grpc.GetMediaThumbnailRequest) (*v1grpc.BinaryDataResponse, error) {
	// Get media to access thumbnail
	media, err := s.mediaSrv.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// Check if media has thumbnail
	if !media.HasThumbnail() {
		return nil, fmt.Errorf("thumbnail not available for media: %s", req.Id)
	}

	return &v1grpc.BinaryDataResponse{
		Data:        media.Thumbnail,
		ContentType: "image/jpeg", // Thumbnails are typically JPEG
		Filename:    media.Filename,
	}, nil
}

func (s *Handler) GetMediaContent(req *v1grpc.GetMediaContentRequest, stream v1grpc.PhotosNGService_GetMediaContentServer) error {
	// Get media
	media, err := s.mediaSrv.Get(stream.Context(), req.Id)
	if err != nil {
		return err
	}

	// Get media content
	contentFn, err := media.Content()
	if err != nil {
		return err
	}

	// Stream content in chunks
	const chunkSize = 32 * 1024 // 32KB chunks
	buffer := make([]byte, chunkSize)
	chunkIndex := int64(0)
	totalSize := int64(0) // Would need to get actual file size

	// Determine content type based on file extension
	contentType := "application/octet-stream"
	if strings.HasSuffix(strings.ToLower(media.Filename), ".jpg") || strings.HasSuffix(strings.ToLower(media.Filename), ".jpeg") {
		contentType = "image/jpeg"
	} else if strings.HasSuffix(strings.ToLower(media.Filename), ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(strings.ToLower(media.Filename), ".mp4") {
		contentType = "video/mp4"
	}

	for {
		n, err := contentFn.Read(buffer)
		if n > 0 {
			chunk := &v1grpc.BinaryDataChunk{
				Chunk:       buffer[:n],
				ChunkIndex:  chunkIndex,
				IsLastChunk: false,
			}

			// Include metadata in first chunk
			if chunkIndex == 0 {
				chunk.ContentType = contentType
				chunk.Filename = media.Filename
				chunk.TotalSize = totalSize
			}

			if err := stream.Send(chunk); err != nil {
				return err
			}
			chunkIndex++
		}

		if err == io.EOF {
			// Send final chunk to indicate end
			finalChunk := &v1grpc.BinaryDataChunk{
				Chunk:       []byte{},
				ChunkIndex:  chunkIndex,
				IsLastChunk: true,
			}
			return stream.Send(finalChunk)
		}
		if err != nil {
			return err
		}
	}
}

func (s *Handler) UploadMedia(ctx context.Context, req *v1grpc.UploadMediaRequest) (*v1grpc.Media, error) {
	// Get album for the upload
	album, err := s.albumSrv.Get(ctx, req.AlbumId)
	if err != nil {
		return nil, err
	}

	// Create media entity from upload request
	fileContent := bytes.NewReader(req.FileContent)
	media := v1grpc.ToMediaEntity(req.Filename, req.AlbumId, fileContent, *album)

	// Create media using service
	createdMedia, err := s.mediaSrv.WriteMedia(ctx, media)
	if err != nil {
		return nil, err
	}

	return v1grpc.NewMedia(*createdMedia), nil
}

// Sync operations implementation
func (s *Handler) StartSyncJob(ctx context.Context, req *v1grpc.StartSyncRequest) (*v1grpc.StartSyncResponse, error) {
	jobID, err := s.syncSrv.StartSync(ctx, req.Path)
	if err != nil {
		return nil, err
	}

	return &v1grpc.StartSyncResponse{
		Id: jobID,
	}, nil
}

func (s *Handler) ListSyncJobs(ctx context.Context, req *v1grpc.ListSyncJobsRequest) (*v1grpc.ListSyncJobsResponse, error) {
	jobs := s.syncSrv.ListJobStatuses()

	grpcJobs := make([]*v1grpc.SyncJob, 0, len(jobs))
	for _, job := range jobs {
		grpcJobs = append(grpcJobs, v1grpc.NewSyncJob(job))
	}

	return &v1grpc.ListSyncJobsResponse{
		Jobs: grpcJobs,
	}, nil
}

func (s *Handler) GetSyncJob(ctx context.Context, req *v1grpc.GetSyncJobRequest) (*v1grpc.SyncJob, error) {
	job, err := s.syncSrv.GetJobStatus(req.Id)
	if err != nil {
		return nil, err
	}

	return v1grpc.NewSyncJob(*job), nil
}

func (s *Handler) StopSyncJob(ctx context.Context, req *v1grpc.StopSyncJobRequest) (*v1grpc.StopSyncJobResponse, error) {
	err := s.syncSrv.StopJob(req.Id)
	if err != nil {
		return nil, err
	}

	return &v1grpc.StopSyncJobResponse{
		Message: "Sync job stopped successfully",
		JobId:   req.Id,
	}, nil
}

func (s *Handler) ActionAllSyncJobs(ctx context.Context, req *v1grpc.ActionAllSyncJobsRequest) (*v1grpc.ActionAllSyncJobsResponse, error) {
	var affectedCount int32
	var message string

	switch req.Action {
	case v1grpc.SyncJobAction_SYNC_JOB_ACTION_STOP:
		// Get count of active jobs before stopping
		runningJobs := s.syncSrv.ListJobStatusesByStatus(entity.StatusRunning)
		pendingJobs := s.syncSrv.ListJobStatusesByStatus(entity.StatusPending)
		affectedCount = int32(len(runningJobs) + len(pendingJobs))

		err := s.syncSrv.StopAllJobs()
		if err != nil {
			return nil, err
		}
		message = "All sync jobs stopped successfully"

	case v1grpc.SyncJobAction_SYNC_JOB_ACTION_RESUME:
		// Resume functionality not yet implemented
		return nil, fmt.Errorf("resume functionality not yet implemented")

	default:
		return nil, fmt.Errorf("invalid action: %v", req.Action)
	}

	return &v1grpc.ActionAllSyncJobsResponse{
		Message:       message,
		Action:        req.Action,
		AffectedCount: affectedCount,
	}, nil
}

func (s *Handler) ActionSyncJob(ctx context.Context, req *v1grpc.ActionSyncJobRequest) (*v1grpc.ActionSyncJobResponse, error) {
	var affectedCount int32
	var message string

	switch req.Action {
	case v1grpc.SyncJobAction_SYNC_JOB_ACTION_STOP:
		err := s.syncSrv.StopJob(req.Id)
		if err != nil {
			return nil, err
		}
		affectedCount = 1
		message = "Sync job stopped successfully"

	case v1grpc.SyncJobAction_SYNC_JOB_ACTION_RESUME:
		// Resume functionality not yet implemented
		return nil, fmt.Errorf("resume functionality not yet implemented")

	default:
		return nil, fmt.Errorf("invalid action: %v", req.Action)
	}

	return &v1grpc.ActionSyncJobResponse{
		Message:       message,
		Action:        req.Action,
		AffectedCount: affectedCount,
	}, nil
}

func (s *Handler) ClearFinishedSyncJobs(ctx context.Context, req *v1grpc.ClearFinishedSyncJobsRequest) (*v1grpc.ClearFinishedSyncJobsResponse, error) {
	// Get count of finished jobs before clearing
	completedJobs := s.syncSrv.ListJobStatusesByStatus(entity.StatusCompleted)
	stoppedJobs := s.syncSrv.ListJobStatusesByStatus(entity.StatusStopped)
	failedJobs := s.syncSrv.ListJobStatusesByStatus(entity.StatusFailed)
	clearedCount := int32(len(completedJobs) + len(stoppedJobs) + len(failedJobs))

	err := s.syncSrv.ClearFinishedJobs()
	if err != nil {
		return nil, err
	}

	return &v1grpc.ClearFinishedSyncJobsResponse{
		Message:      "Finished sync jobs cleared",
		ClearedCount: clearedCount,
	}, nil
}

func (s *Handler) StopAllSyncJobs(ctx context.Context, req *v1grpc.StopAllSyncJobsRequest) (*v1grpc.StopAllSyncJobsResponse, error) {
	err := s.syncSrv.StopAllJobs()
	if err != nil {
		return nil, err
	}

	return &v1grpc.StopAllSyncJobsResponse{
		Message:      "All sync jobs stopped",
		StoppedCount: 0, // StopAllSyncJobs doesn't return count
	}, nil
}

func (s *Handler) GetStats(ctx context.Context, req *v1grpc.GetStatsRequest) (*v1grpc.StatsResponse, error) {
	stats, err := s.statsSrv.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	years := make([]int32, 0, len(stats.Years))
	for _, year := range stats.Years {
		years = append(years, int32(year))
	}

	return &v1grpc.StatsResponse{
		Years:      years,
		CountMedia: int32(stats.CountMedia),
		CountAlbum: int32(stats.CountAlbum),
	}, nil
}
