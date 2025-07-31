package entity

import "time"

type Album struct {
	ID          string
	CreatedAt   time.Time
	Path        string
	Description *string
	Thumbnail   *string
	Parent      *string
	Children    []Album
	Media       []Media
}

func NewAlbum(folderPath string) Album {
	return Album{
		ID:       generateId(folderPath),
		Path:     folderPath,
		Children: make([]Album, 0),
		Media:    make([]Media, 0),
	}
}
