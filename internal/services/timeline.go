package services

import (
	"context"
	"sort"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

// TimelineFilter represents filtering criteria for timeline queries
type TimelineFilter struct {
	StartDate time.Time
	Limit     int
	Offset    int
}

// TimelineService provides business logic for timeline operations
type TimelineService struct {
	dt *pg.Datastore
}

// NewTimelineService creates a new instance of TimelineService with the provided datastore
func NewTimelineService(dt *pg.Datastore) *TimelineService {
	return &TimelineService{dt: dt}
}

// GetTimeline retrieves media organized into timeline buckets
func (t *TimelineService) GetTimeline(ctx context.Context, filter *TimelineFilter) (entity.Buckets, []int, error) {
	// Get all media from the start date onwards
	// TODO: Add proper date filtering in the query
	queryOptions := []pg.QueryOption{
		pg.SortByColumn("captured_at", true), // Most recent first
	}

	allMedia, err := t.dt.QueryMedia(ctx, queryOptions...)
	if err != nil {
		return nil, nil, err
	}

	// Filter media by start date
	filteredMedia := make([]entity.Media, 0)
	for _, media := range allMedia {
		if media.CapturedAt.After(filter.StartDate) || media.CapturedAt.Equal(filter.StartDate) {
			filteredMedia = append(filteredMedia, media)
		}
	}

	// Organize media into buckets
	buckets := t.organizeBuckets(filteredMedia)

	// Get years from datastore stats
	stats, err := t.dt.Stats(ctx)
	if err != nil {
		return nil, nil, err
	}
	years := stats.TimelineYears

	// Apply pagination to buckets
	if filter.Offset >= len(buckets) {
		buckets = entity.Buckets{}
	} else {
		end := filter.Offset + filter.Limit
		if end > len(buckets) {
			end = len(buckets)
		}
		buckets = buckets[filter.Offset:end]
	}

	return buckets, years, nil
}

// organizeBuckets groups media items by year and month
func (t *TimelineService) organizeBuckets(media []entity.Media) entity.Buckets {
	bucketMap := make(map[string][]entity.Media)

	// Group media by year-month
	for _, item := range media {
		key := item.CapturedAt.Format("2006-01") // YYYY-MM format
		bucketMap[key] = append(bucketMap[key], item)
	}

	// Convert to bucket slice
	buckets := make(entity.Buckets, 0, len(bucketMap))
	for dateKey, mediaItems := range bucketMap {
		// Parse the date key
		date, err := time.Parse("2006-01", dateKey)
		if err != nil {
			continue // Skip invalid dates
		}

		bucket := entity.Bucket{
			Year:  date.Year(),
			Month: int(date.Month()),
			Media: mediaItems,
		}

		buckets = append(buckets, bucket)
	}

	// Sort buckets by date (most recent first)
	sort.Slice(buckets, func(i, j int) bool {
		if buckets[i].Year != buckets[j].Year {
			return buckets[i].Year > buckets[j].Year
		}
		return buckets[i].Month > buckets[j].Month
	})

	return buckets
}
