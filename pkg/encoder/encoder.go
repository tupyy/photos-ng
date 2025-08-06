package encoder

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"

	"github.com/disintegration/imaging"
)

var (
	defaultThumbnailMethod = imaging.NearestNeighbor
)

func EncodeThumbnail(inputPhoto io.Reader) (io.Reader, error) {
	inputImage, err := imaging.Decode(inputPhoto, imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}

	dimensions := PhotoDimensionsFromRect(inputImage.Bounds())
	dimensions = dimensions.ThumbnailScale()

	buff := new(bytes.Buffer)
	thumbImage := imaging.Resize(inputImage, dimensions.Width, dimensions.Height, defaultThumbnailMethod)
	if err = encodeImageJPEG(thumbImage, buff, 60); err != nil {
		return nil, err
	}

	return buff, nil
}

func encodeImageJPEG(image image.Image, w io.Writer, jpegQuality int) error {
	err := jpeg.Encode(w, image, &jpeg.Options{Quality: jpegQuality})
	if err != nil {
		return err
	}

	return nil
}
