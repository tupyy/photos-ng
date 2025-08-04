package entity

import "time"

type Album struct {
	ID          string
	CreatedAt   time.Time
	Path        string
	Description *string
	Thumbnail   *string
	ParentId    *string
	Children    []Album
	Media       []Media
}

func NewAlbum(folderPath string) Album {
	return Album{
		ID:        GenerateId(folderPath),
		Path:      folderPath,
		CreatedAt: time.Now(),
		Children:  make([]Album, 0),
		Media:     make([]Media, 0),
	}
}
