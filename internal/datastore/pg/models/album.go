package models

import (
	"encoding/json"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

// Albums stores albums indexed by ID for efficient grouping of related records
type Albums map[string][]Album

// Add appends an album to the collection, grouping by album ID.
func (aa Albums) Add(a Album) {
	rows, ok := aa[a.ID]
	if !ok {
		aa[a.ID] = []Album{a}
		return
	}
	aa[a.ID] = append(rows, a)
}

// Entity converts the database model albums to entity albums.
func (aa Albums) Entity() []entity.Album {
	albums := []entity.Album{}
	for _, rows := range aa {
		album := entity.Album{
			Children: []entity.Album{},
			Media:    []entity.Media{},
		}
		for _, row := range rows {
			album.ID = row.ID
			album.CreatedAt = row.CreatedAt
			album.Path = row.Path
			album.Description = row.Description
			album.Parent = row.ParentID
			if row.ThumbnailID != nil {
				album.Thumbnail = row.ThumbnailID
			}
			// Child albums are added through join scenarios
			if row.ChildID != nil {
				childAlbum := entity.Album{
					ID:          *row.ChildID,
					Path:        fromPtr(row.ChildPath),
					Description: row.ChildDescription,
					Thumbnail:   row.ChildThumbnailID,
					Children:    []entity.Album{},
					Media:       []entity.Media{},
				}
				if row.ChildCreatedAt != nil {
					childAlbum.CreatedAt = *row.ChildCreatedAt
				}
				album.Children = append(album.Children, childAlbum)
			}

			// Media are added through join scenarios
			if row.MediaID != nil {
				// Parse EXIF metadata from JSON
				var exifMetadata map[string]string
				if row.MediaExif != nil {
					if err := json.Unmarshal(*row.MediaExif, &exifMetadata); err != nil {
						exifMetadata = make(map[string]string)
					}
				} else {
					exifMetadata = make(map[string]string)
				}

				mediaItem := entity.Media{
					ID:        *row.MediaID,
					Filename:  fromPtr(row.MediaFileName),
					Thumbnail: row.MediaThumbnail,
					Exif:      exifMetadata,
					MediaType: entity.MediaType(fromPtr(row.MediaMediaType)),
				}
				if row.MediaCapturedAt != nil {
					mediaItem.CapturedAt = *row.MediaCapturedAt
				}
				album.Media = append(album.Media, mediaItem)
			}
		}
		albums = append(albums, album)
	}
	return albums
}

// Album represents the database model for albums table
type Album struct {
	ID          string    `db:"id"`
	CreatedAt   time.Time `db:"created_at"`
	Path        string    `db:"path"`
	Description *string   `db:"description"`
	ParentID    *string   `db:"parent_id"`
	ThumbnailID *string   `db:"thumbnail_id"`

	// Fields for join scenarios (child albums)
	ChildID          *string    `db:"child_id"`
	ChildCreatedAt   *time.Time `db:"child_created_at"`
	ChildPath        *string    `db:"child_path"`
	ChildDescription *string    `db:"child_description"`
	ChildThumbnailID *string    `db:"child_thumbnail_id"`

	// Fields for join scenarios (media)
	MediaID         *string          `db:"media_id"`
	MediaCreatedAt  *time.Time       `db:"media_created_at"`
	MediaCapturedAt *time.Time       `db:"media_captured_at"`
	MediaAlbumID    *string          `db:"media_album_id"`
	MediaFileName   *string          `db:"media_file_name"`
	MediaThumbnail  []byte           `db:"media_thumbnail"`
	MediaExif       *json.RawMessage `db:"media_exif"`
	MediaMediaType  *string          `db:"media_media_type"`
}

// fromPtr safely converts *string to string, returning empty string if nil
func fromPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
