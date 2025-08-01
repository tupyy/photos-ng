package services

import (
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
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

	// Add pagination
	if mf.MediaLimit > 0 {
		qf = append(qf, pg.Limit(mf.MediaLimit))
	}
	if mf.MediaOffset > 0 {
		qf = append(qf, pg.Offset(mf.MediaOffset))
	}

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

	return qf
}

// AlbumOptions represents optionsing criteria for album queries
//
//go:generate go run github.com/ecordell/optgen -output zz_generated.album_options.go . AlbumOptions
type AlbumOptions struct {
	Limit    int     `debugmap:"visible"`
	Offset   int     `debugmap:"visible"`
	ParentID *string `debugmap:"visible"`
}

// QueriesFn returns a slice of query options based on the album filter criteria
func (af *AlbumOptions) QueriesFn() []pg.QueryOption {
	qf := []pg.QueryOption{}

	// Add pagination
	if af.Limit > 0 {
		qf = append(qf, pg.Limit(af.Limit))
	}
	if af.Offset > 0 {
		qf = append(qf, pg.Offset(af.Offset))
	}

	// Add parent filter
	if af.ParentID != nil {
		qf = append(qf, pg.FilterByColumnName("parent", *af.ParentID))
	}

	return qf
}
