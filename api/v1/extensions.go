package v1

import (
	"fmt"
	"path"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"github.com/oapi-codegen/runtime/types"
)

// NewAlbum converts an entity.Album to a v1.Album for API responses
func NewAlbum(album entity.Album) Album {
	apiAlbum := Album{
		Id:          album.ID,
		Path:        album.Path,
		Description: album.Description,
		Href:        "/api/v1/albums/" + album.ID,
	}

	_, name := path.Split(album.Path)
	apiAlbum.Name = name

	// Convert children
	if len(album.Children) > 0 {
		children := make([]struct {
			Href string `json:"href"`
			Name string `json:"name"`
		}, 0, len(album.Children))

		for _, childID := range album.Children {
			children = append(children, struct {
				Href string `json:"href"`
				Name string `json:"name"`
			}{
				Href: "/api/v1/albums/" + childID.ID,
				Name: childID.ID, // Using ID as name for now
			})
		}
		apiAlbum.Children = &children
	}

	// Convert media references
	if len(album.Media) > 0 {
		mediaHrefs := make([]string, 0, len(album.Media))
		for _, media := range album.Media {
			mediaHrefs = append(mediaHrefs, "/api/v1/media/"+media.ID)
		}
		apiAlbum.Media = &mediaHrefs
	}

	// Set parent href if parent exists
	if album.ParentId != nil {
		parentHref := "/api/v1/albums/" + *album.ParentId
		apiAlbum.ParentHref = &parentHref
	}

	if album.Thumbnail != nil {
		thumbnail := fmt.Sprintf("/api/v1/media/%s/thumbnail", *album.Thumbnail)
		apiAlbum.Thumbnail = &thumbnail
	}

	return apiAlbum
}

// NewMedia converts an entity.Media to a v1.Media for API responses
func NewMedia(media entity.Media) Media {
	// Convert captured date
	capturedAt := types.Date{Time: media.CapturedAt}

	// Convert EXIF data
	exifHeaders := make([]ExifHeader, 0, len(media.Exif))
	for key, value := range media.Exif {
		exifHeaders = append(exifHeaders, ExifHeader{
			Key:   key,
			Value: value,
		})
	}

	return Media{
		Id:         media.ID,
		Filename:   media.Filename,
		AlbumHref:  "/api/v1/albums/" + media.Album.ID,
		CapturedAt: capturedAt,
		Type:       string(media.MediaType),
		Content:    "/api/v1/media/" + media.ID + "/content",
		Thumbnail:  "/api/v1/media/" + media.ID + "/thumbnail",
		Href:       "/api/v1/media/" + media.ID,
		Exif:       exifHeaders,
	}
}

// Entity converts a v1.CreateAlbumRequest to an entity.Album for business logic processing.
// This method transforms the HTTP request data into the internal domain model representation.
func (r CreateAlbumRequest) Entity() entity.Album {
	album := entity.Album{
		ID:          entity.GenerateId(r.Name),
		Path:        r.Name,
		ParentId:    r.ParentId,
		Description: r.Description,
		Children:    []entity.Album{},
		Media:       []entity.Media{},
		CreatedAt:   time.Now(),
	}

	return album
}

// Entity converts a v1.UpdateAlbumRequest to an entity.Album for business logic processing.
// This method applies the updates to an existing album entity.
func (r UpdateAlbumRequest) ApplyTo(album *entity.Album) {
	album.Description = r.Description
	album.Thumbnail = r.Thumbnail
}

// Entity converts a v1.UpdateMediaRequest to updates for an entity.Media.
// This method applies the updates to an existing media entity.
func (r UpdateMediaRequest) ApplyTo(media *entity.Media) {
	if r.CapturedAt != nil {
		media.CapturedAt = r.CapturedAt.Time
	}
	if r.Exif != nil {
		if media.Exif == nil {
			media.Exif = make(map[string]string)
		}
		for _, exif := range *r.Exif {
			media.Exif[exif.Key] = exif.Value
		}
	}
}
