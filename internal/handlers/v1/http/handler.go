package v1

import (
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
)

// Handler implements the V1 API handlers for the photos-ng application.
// It contains the business logic for handling HTTP requests and responses
// for all V1 endpoints including albums and media.
type Handler struct {
	albumSrv *services.AlbumService
	mediaSrv *services.MediaService
	statsSrv *services.StatsService
	syncSrv  *services.SyncService
}

func NewHandler(dt *pg.Datastore, fs *fs.Datastore) *Handler {
	albumSrv := services.NewAlbumService(dt, fs)
	mediaSrv := services.NewMediaService(dt, fs)
	syncSrv := services.NewSyncService(albumSrv, mediaSrv, fs)
	statsSrv := services.NewStatsService(dt)

	return &Handler{
		albumSrv: albumSrv,
		mediaSrv: mediaSrv,
		statsSrv: statsSrv,
		syncSrv:  syncSrv,
	}
}
