package encoder

import (
	"image"
	"io"
)

const (
	defaultWidth = 640
)

type PhotoDimensions struct {
	Width  int
	Height int
}

func GetPhotoDimensions(photo io.Reader) (*PhotoDimensions, error) {
	config, _, err := image.DecodeConfig(photo)
	if err != nil {
		return nil, err
	}

	return &PhotoDimensions{
		Width:  config.Width,
		Height: config.Height,
	}, nil
}

func PhotoDimensionsFromRect(rect image.Rectangle) PhotoDimensions {
	return PhotoDimensions{
		Width:  rect.Bounds().Max.X,
		Height: rect.Bounds().Max.Y,
	}
}

func (dimensions *PhotoDimensions) ThumbnailScale() PhotoDimensions {
	aspect := float64(dimensions.Width) / float64(dimensions.Height)

	var width, height int

	if aspect > 1 {
		width = defaultWidth
		height = int(defaultWidth / aspect)
	} else {
		width = int(defaultWidth * aspect)
		height = defaultWidth
	}

	if width > dimensions.Width {
		width = dimensions.Width
		height = dimensions.Height
	}

	return PhotoDimensions{
		Width:  width,
		Height: height,
	}
}
