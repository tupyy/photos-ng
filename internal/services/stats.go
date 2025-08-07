package services

import (
	"context"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

// StatsService provides business logic for statistics operations
type StatsService struct {
	dt *pg.Datastore
}

// NewStatsService creates a new instance of StatsService with the provided datastore
func NewStatsService(dt *pg.Datastore) *StatsService {
	return &StatsService{dt: dt}
}

// GetStats retrieves statistics including album count, media count, and years with media
func (s *StatsService) GetStats(ctx context.Context) (entity.Stats, error) {
	return s.dt.Stats(ctx)
}
