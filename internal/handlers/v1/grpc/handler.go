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
	opts := services.NewAlbumOptionsWithOptions(
		services.WithLimit(limit),
		services.WithOffset(offset),
		services.WithHasParent(hasParent),
	)

	albums, err := s.albumSrv.GetAlbums(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Get total count for pagination (without limit/offset)
	totalOpts := services.NewAlbumOptionsWithOptions(
		services.WithHasParent(hasParent),
	)
	allAlbums, err := s.albumSrv.GetAlbums(ctx, totalOpts)
	if err != nil {
		return nil, err
	}
	total := len(allAlbums)

	// Convert entity albums to gRPC albums
	grpcAlbums := make([]*v1grpc.Album, 0, len(albums))
	for _, album := range albums {
		grpcAlbums = append(grpcAlbums, v1grpc.NewAlbum(album))
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
	album, err := s.albumSrv.GetAlbum(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return v1grpc.NewAlbum(*album), nil
}

func (s *Handler) CreateAlbum(ctx context.Context, req *v1grpc.CreateAlbumRequest) (*v1grpc.Album, error) {
	// Convert gRPC request to entity
	album := req.Entity()

	// Create album using service
	createdAlbum, err := s.albumSrv.CreateAlbum(ctx, album)
	if err != nil {
		return nil, err
	}

	return v1grpc.NewAlbum(*createdAlbum), nil
}

func (s *Handler) UpdateAlbum(ctx context.Context, req *v1grpc.UpdateAlbumByIdRequest) (*v1grpc.Album, error) {
	// Get existing album
	album, err := s.albumSrv.GetAlbum(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	req.Update.ApplyTo(album)

	// Update album using service
	updatedAlbum, err := s.albumSrv.UpdateAlbum(ctx, *album)
	if err != nil {
		return nil, err
	}

	return v1grpc.NewAlbum(*updatedAlbum), nil
}

func (s *Handler) DeleteAlbum(ctx context.Context, req *v1grpc.DeleteAlbumRequest) (*emptypb.Empty, error) {
	err := s.albumSrv.DeleteAlbum(ctx, req.Id)
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
	mediaItems, err := s.mediaSrv.GetMedia(stream.Context(), opt)
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

	// Get next cursor for pagination and send in trailer
	if len(mediaItems) == limit {
		// Check if there's a next page by requesting one more item
		nextOpt := &services.MediaOptions{
			MediaLimit: 1,
			Direction:  opt.Direction,
			AlbumID:    opt.AlbumID,
			MediaType:  opt.MediaType,
		}

		// Set cursor based on last item
		if len(mediaItems) > 0 {
			lastItem := mediaItems[len(mediaItems)-1]
			nextOpt.Cursor = &services.PaginationCursor{
				CapturedAt: lastItem.CapturedAt,
				ID:         lastItem.ID,
			}
		}

		nextItems, err := s.mediaSrv.GetMedia(stream.Context(), nextOpt)
		if err == nil && len(nextItems) > 0 {
			// There are more items, encode cursor for next page
			lastItem := mediaItems[len(mediaItems)-1]
			nextCursor := &services.PaginationCursor{
				CapturedAt: lastItem.CapturedAt,
				ID:         lastItem.ID,
			}
			encodedCursor, err := nextCursor.Encode()
			if err == nil {
				// Send cursor in trailer metadata
				stream.SetTrailer(metadata.Pairs("next-cursor", encodedCursor))
			}
		}
	}

	return nil
}

func (s *Handler) GetMedia(ctx context.Context, req *v1grpc.GetMediaRequest) (*v1grpc.Media, error) {
	media, err := s.mediaSrv.GetMediaByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return v1grpc.NewMedia(*media), nil
}

func (s *Handler) UpdateMedia(ctx context.Context, req *v1grpc.UpdateMediaByIdRequest) (*v1grpc.Media, error) {
	// Get existing media
	media, err := s.mediaSrv.GetMediaByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	req.Update.ApplyTo(media)

	// Update media using service
	updatedMedia, err := s.mediaSrv.UpdateMedia(ctx, *media)
	if err != nil {
		return nil, err
	}

	return v1grpc.NewMedia(*updatedMedia), nil
}

func (s *Handler) DeleteMedia(ctx context.Context, req *v1grpc.DeleteMediaRequest) (*emptypb.Empty, error) {
	err := s.mediaSrv.DeleteMedia(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Handler) GetMediaThumbnail(ctx context.Context, req *v1grpc.GetMediaThumbnailRequest) (*v1grpc.BinaryDataResponse, error) {
	// Get media to access thumbnail
	media, err := s.mediaSrv.GetMediaByID(ctx, req.Id)
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
	media, err := s.mediaSrv.GetMediaByID(stream.Context(), req.Id)
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
	album, err := s.albumSrv.GetAlbum(ctx, req.AlbumId)
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
	jobs := s.syncSrv.ListSyncJobStatuses()

	grpcJobs := make([]*v1grpc.SyncJob, 0, len(jobs))
	for _, job := range jobs {
		grpcJobs = append(grpcJobs, v1grpc.NewSyncJob(job))
	}

	return &v1grpc.ListSyncJobsResponse{
		Jobs: grpcJobs,
	}, nil
}

func (s *Handler) GetSyncJob(ctx context.Context, req *v1grpc.GetSyncJobRequest) (*v1grpc.SyncJob, error) {
	job, err := s.syncSrv.GetSyncJobStatus(req.Id)
	if err != nil {
		return nil, err
	}

	return v1grpc.NewSyncJob(*job), nil
}

func (s *Handler) StopSyncJob(ctx context.Context, req *v1grpc.StopSyncJobRequest) (*v1grpc.StopSyncJobResponse, error) {
	err := s.syncSrv.StopSyncJob(req.Id)
	if err != nil {
		return nil, err
	}

	return &v1grpc.StopSyncJobResponse{
		Message: "Sync job stopped successfully",
		JobId:   req.Id,
	}, nil
}

func (s *Handler) StopAllSyncJobs(ctx context.Context, req *v1grpc.StopAllSyncJobsRequest) (*v1grpc.StopAllSyncJobsResponse, error) {
	err := s.syncSrv.StopAllSyncJobs()
	if err != nil {
		return nil, err
	}

	return &v1grpc.StopAllSyncJobsResponse{
		Message:      "All sync jobs stopped",
		StoppedCount: 0, // StopAllSyncJobs doesn't return count
	}, nil
}

// Stats operations implementation
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
