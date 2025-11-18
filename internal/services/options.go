package services

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
)

const (
	SortByCapturedAt string = "captured_at"
)

// PaginationCursor represents a cursor for pagination based on captured_at and id
type PaginationCursor struct {
	CapturedAt time.Time `json:"captured_at"`
	ID         string    `json:"id"`
}

// Encode converts the cursor to a base64 encoded string for URL usage
func (c *PaginationCursor) Encode() (string, error) {
	if c == nil {
		return "", nil
	}
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(data), nil
}

// DecodeCursor parses a base64 encoded cursor string
func DecodeCursor(encoded string) (*PaginationCursor, error) {
	if encoded == "" {
		return nil, nil
	}

	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	// Parse into temporary struct with string fields to handle date-only formats
	var temp struct {
		CapturedAt string `json:"captured_at"`
		ID         string `json:"id"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, err
	}

	// Parse the timestamp - try RFC3339 first, then date-only
	var capturedAt time.Time
	if capturedAt, err = time.Parse(time.RFC3339, temp.CapturedAt); err != nil {
		// Try date-only format as fallback
		if capturedAt, err = time.Parse("2006-01-02", temp.CapturedAt); err != nil {
			return nil, err
		}
	}

	return &PaginationCursor{
		CapturedAt: capturedAt,
		ID:         temp.ID,
	}, nil
}

//go:generate go run github.com/ecordell/optgen -output zz_generated.media_options.go . MediaOptions
type MediaOptions struct {
	MediaLimit int               `debugmap:"visible"`
	Cursor     *PaginationCursor `debugmap:"visible"`
	Direction  string            `debugmap:"visible"`
	SortBy     string            `debugmap:"visible"`
	AlbumID    *string           `debugmap:"visible"`
	MediaType  *string           `debugmap:"visible"`
	StartDate  *time.Time        `debugmap:"visible"`
	EndDate    *time.Time        `debugmap:"visible"`
}

// QueriesFn returns a slice of query options based on the media filter criteria
func (mf *MediaOptions) QueriesFn() []pg.QueryOption {
	qf := []pg.QueryOption{}

	// Add album filter
	if mf.AlbumID != nil {
		qf = append(qf, pg.FilterByColumnName("media.album_id", *mf.AlbumID))
	}

	// Add media type filter
	if mf.MediaType != nil {
		qf = append(qf, pg.FilterByColumnName("media_type", *mf.MediaType))
	}

	// Add date range filters
	// TODO: Implement date range filtering in query options
	// For now, we'll filter in the service layer after querying

	// Add cursor-based filtering
	if mf.Cursor != nil {
		direction := mf.Direction
		if direction == "" {
			direction = "forward" // default direction
		}
		qf = append(qf, pg.FilterByCursor(mf.Cursor.CapturedAt, mf.Cursor.ID, direction))
	}

	// Add direction-based sorting
	if mf.Direction == "backward" {
		// Backward: sort ASC to get newer photos, then reverse in service
		qf = append(qf, pg.SortByColumn("captured_at", false)) // ASC
		qf = append(qf, pg.SortByColumn("id", false))          // ASC
	} else {
		// Forward: sort DESC to get older photos (default behavior)
		qf = append(qf, pg.SortByColumn("captured_at", true)) // DESC
		qf = append(qf, pg.SortByColumn("id", true))          // DESC
	}

	// Add limit
	if mf.MediaLimit > 0 {
		qf = append(qf, pg.Limit(mf.MediaLimit))
	}

	return qf
}

// ListOptions represents optionsing criteria for album queries
//
//go:generate go run github.com/ecordell/optgen -output zz_generated.album_options.go . ListOptions
type ListOptions struct {
	Limit     int     `debugmap:"visible"`
	Offset    int     `debugmap:"visible"`
	ParentID  *string `debugmap:"visible"`
	HasParent bool    `debugmap:"visible"`
}

// QueriesFn returns a slice of query options based on the album filter criteria
// Note: Pagination (Limit/Offset) is removed because it interferes with JOIN queries
// and causes incorrect media counts. Album pagination should be handled at the application level.
func (af *ListOptions) QueriesFn() []pg.QueryOption {
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
