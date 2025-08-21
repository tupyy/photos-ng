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
	albumSrv   *services.AlbumService
	mediaSrv   *services.MediaService
	statsSrv   *services.StatsService
	syncSrv    *services.SyncService
	rootFolder string
}

func NewHandler(dt *pg.Datastore, rootFolder string) *Handler {
	fsDatastore := fs.NewFsDatastore(rootFolder)
	albumSrv := services.NewAlbumService(dt, fsDatastore)
	mediaSrv := services.NewMediaService(dt, fsDatastore)
	syncSrv := services.NewSyncService(albumSrv, mediaSrv, fsDatastore)

	return &Handler{
		albumSrv:   albumSrv,
		mediaSrv:   mediaSrv,
		statsSrv:   services.NewStatsService(dt),
		syncSrv:    syncSrv,
		rootFolder: rootFolder,
	}
}
