package v1

import (
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
)

// ServerImpl implements the V1 API handlers for the photos-ng application.
// It contains the business logic for handling HTTP requests and responses
// for all V1 endpoints including albums, media, and timeline.
type ServerImpl struct {
	albumSrv    *services.AlbumService
	mediaSrv    *services.MediaService
	timelineSrv *services.TimelineService
}

func NewServerV1(dt *pg.Datastore, rootFolder string) *ServerImpl {
	return &ServerImpl{
		albumSrv:    services.NewAlbumService(dt, fs.NewFsDatastore(rootFolder)),
		mediaSrv:    services.NewMediaService(dt, fs.NewFsDatastore(rootFolder)),
		timelineSrv: services.NewTimelineService(dt),
	}
}
