package entity

import "time"

type Album struct {
	ID        string
	CreatedAt time.Time
	Path      string
	Thumbnail *string
	Parent    *string
	Children  []string
	Media     []Media
}

func NewAlbum(folderPath string) Album {
	return Album{
		ID:       generateId(folderPath),
		Path:     folderPath,
		Children: make([]string, 0),
		Media:    make([]Media, 0),
	}
}
