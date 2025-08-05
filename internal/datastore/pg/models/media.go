package models

import (
	"encoding/json"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

// MediaList stores media records indexed by ID for efficient grouping
type MediaList map[string]Media

// Add appends a media to the collection.
func (mm MediaList) Add(m Media) {
	mm[m.ID] = m
}

// Entity converts the database model media to entity media.
func (mm MediaList) Entity() []entity.Media {
	mediaList := []entity.Media{}
	for _, m := range mm {
		// Parse EXIF metadata from JSON
		var exifMetadata map[string]string
		if m.Exif != nil {
			if err := json.Unmarshal(*m.Exif, &exifMetadata); err != nil {
				exifMetadata = make(map[string]string)
			}
		} else {
			exifMetadata = make(map[string]string)
		}

		media := entity.Media{
			ID:         m.ID,
			CapturedAt: m.CapturedAt,
			Filename:   m.FileName,
			Thumbnail:  m.Thumbnail,
			Hash:       m.Hash,
			Exif:       exifMetadata,
			MediaType:  entity.MediaType(m.MediaType),
			// Album will be populated separately through joins or additional queries
		}
		mediaList = append(mediaList, media)
	}
	return mediaList
}

// Media represents the database model for media table
type Media struct {
	ID         string           `db:"id"`
	CreatedAt  time.Time        `db:"created_at"`
	CapturedAt time.Time        `db:"captured_at"`
	AlbumID    string           `db:"album_id"`
	FileName   string           `db:"file_name"`
	Hash       string           `db:"hash"`
	Thumbnail  []byte           `db:"thumbnail"`
	Exif       *json.RawMessage `db:"exif"`
	MediaType  string           `db:"media_type"`
}
