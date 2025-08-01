package entity

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"
)

type MediaContentFn func() (io.Reader, error)

type Thumbnail struct {
	Path string
}

type MediaType string

var (
	exifCapturedTime = map[int][]string{
		1: {"ModifyDate", "2006:01:02 15:04:05"},
		2: {"CreateDate", "2006:01:02 15:04:05"},
		3: {"FileModifyDate", "2006:01:02 15:04:05+02:00"},
	}
)

const (
	Photo MediaType = "photo"
	Video MediaType = "Video"
)

type Media struct {
	ID         string
	Album      Album
	CapturedAt time.Time
	MediaType  MediaType
	Filename   string
	Thumbnail  []byte
	Content    MediaContentFn
	Exif       map[string]string
}

func NewMedia(filename string, album Album) Media {
	return Media{
		ID:        generateId(fmt.Sprintf("%s%s", filename, album.ID)),
		Album:     album,
		Filename:  filename,
		MediaType: Photo,
		Exif:      make(map[string]string),
	}
}

func (m Media) ContentType() string {
	ext := path.Ext(strings.ToLower(m.Filename))
	switch strings.TrimLeft(ext, ".") {
	case "jpg":
		fallthrough
	case "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	default:
		return "image/unknown"
	}
}

func (m Media) GetCapturedTime() (time.Time, error) {
	for i := 1; i <= len(exifCapturedTime); i++ {
		val := exifCapturedTime[i]
		if capturedAt, ok := m.Exif[val[0]]; ok {
			return time.Parse(val[1], capturedAt)
		}
	}

	return time.Now(), errors.New("failed to find captured at value in exif meta data")
}

func (m Media) Filepath() string {
	return path.Join(m.Album.Path, m.Filename)
}

func (m Media) HasThumbnail() bool {
	return len(m.Thumbnail) > 0
}

func (m Media) String() string {
	mm := struct {
		ID         string
		Album      Album
		CapturedAt time.Time
		MediaType  MediaType
		Filename   string
	}{
		ID:         m.ID,
		Album:      m.Album,
		CapturedAt: m.CapturedAt,
		MediaType:  m.MediaType,
		Filename:   m.Filename,
	}

	data, _ := json.Marshal(mm)
	return string(data)
}
