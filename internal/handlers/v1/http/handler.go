package v1

import (
	"github.com/authzed/authzed-go/v1"

	authzStore "git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/authz"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	v1 "git.tls.tupangiu.ro/cosmin/photos-ng/internal/handlers/v1"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
)

// Handler implements the V1 API handlers for the photos-ng application.
// It contains the business logic for handling HTTP requests and responses
// for all V1 endpoints including albums and media.
type Handler struct {
	albumSrv v1.AlbumService
	mediaSrv v1.MediaService
	statsSrv *services.StatsService
	syncSrv  v1.SyncService
}

func NewHandler(dt *pg.Datastore, fs *fs.Datastore) *Handler {
	baseAlbumSrv := services.NewAlbumService(dt, fs)
	mediaSrv := services.NewMediaService(dt, fs)
	syncSrv := services.NewSyncService(baseAlbumSrv, mediaSrv, fs)
	statsSrv := services.NewStatsService(dt)

	return &Handler{
		albumSrv: baseAlbumSrv,
		mediaSrv: mediaSrv,
		statsSrv: statsSrv,
		syncSrv:  syncSrv,
	}
}

func NewHandlerWithAuthorization(spiceDBClient *authzed.Client, dt *pg.Datastore, fs *fs.Datastore) *Handler {
	authzSrv := services.NewAuthzService(authzStore.NewAuthzDatastore(spiceDBClient), dt)
	albumSrv := services.NewAlbumService(dt, fs)
	authzAlbumSrv := services.NewAuthzAlbumService(authzSrv, dt, fs)
	mediaSrv := services.NewMediaService(dt, fs)
	authzMediaSrv := services.NewAuthzMediaService(authzSrv, dt, fs)
	syncSrv := services.NewAuthzSyncService(authzSrv, albumSrv, mediaSrv, fs)
	statsSrv := services.NewStatsService(dt)

	return &Handler{
		albumSrv: authzAlbumSrv,
		mediaSrv: authzMediaSrv,
		statsSrv: statsSrv,
		syncSrv:  syncSrv,
	}
}
