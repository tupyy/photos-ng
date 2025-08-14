package services

import (
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
)

const (
	SortByCapturedAt string = "captured_at"
)

//go:generate go run github.com/ecordell/optgen -output zz_generated.media_options.go . MediaOptions
type MediaOptions struct {
	MediaLimit  int        `debugmap:"visible"`
	MediaOffset int        `debugmap:"visible"`
	SortBy      string     `debugmap:"visible"`
	Descending  bool       `debugmap:"visible"`
	AlbumID     *string    `debugmap:"visible"`
	MediaType   *string    `debugmap:"visible"`
	StartDate   *time.Time `debugmap:"visible"`
	EndDate     *time.Time `debugmap:"visible"`
}

// QueriesFn returns a slice of query options based on the media filter criteria
func (mf *MediaOptions) QueriesFn() []pg.QueryOption {
	qf := []pg.QueryOption{}

	// Add album filter
	if mf.AlbumID != nil {
		qf = append(qf, pg.FilterByColumnName("album_id", *mf.AlbumID))
	}

	// Add media type filter
	if mf.MediaType != nil {
		qf = append(qf, pg.FilterByColumnName("media_type", *mf.MediaType))
	}

	// Add date range filters
	// TODO: Implement date range filtering in query options
	// For now, we'll filter in the service layer after querying

	// Add sorting
	if mf.SortBy != "" {
		qf = append(qf, pg.SortByColumn(mf.SortBy, mf.Descending))
	} else {
		// Default sort by captured_at descending
		qf = append(qf, pg.SortByColumn("captured_at", true))
	}

	// Add pagination
	if mf.MediaLimit > 0 {
		qf = append(qf, pg.Limit(mf.MediaLimit))
	}
	if mf.MediaOffset > 0 {
		qf = append(qf, pg.Offset(mf.MediaOffset))
	}

	return qf
}

// AlbumOptions represents optionsing criteria for album queries
//
//go:generate go run github.com/ecordell/optgen -output zz_generated.album_options.go . AlbumOptions
type AlbumOptions struct {
	Limit     int     `debugmap:"visible"`
	Offset    int     `debugmap:"visible"`
	ParentID  *string `debugmap:"visible"`
	HasParent bool    `debugmap:"visible"`
}

// QueriesFn returns a slice of query options based on the album filter criteria
// Note: Pagination (Limit/Offset) is removed because it interferes with JOIN queries
// and causes incorrect media counts. Album pagination should be handled at the application level.
func (af *AlbumOptions) QueriesFn() []pg.QueryOption {
	qf := []pg.QueryOption{}

	// Note: Pagination removed due to JOIN query issues
	// TODO: Implement pagination at application level after query execution

	qf = append(qf, pg.FilterAlbumWithParents(af.HasParent))

	// Add parent filter
	if af.ParentID != nil {
		qf = append(qf, pg.FilterByColumnName("parent", *af.ParentID))
	}

	return qf
}
