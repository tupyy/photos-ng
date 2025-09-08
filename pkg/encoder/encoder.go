package encoder

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"

	"github.com/disintegration/imaging"
)

func EncodeThumbnail(inputPhoto io.Reader) (io.Reader, error) {
	inputImage, err := imaging.Decode(inputPhoto, imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}

	buff := new(bytes.Buffer)
	if err = encodeImageJPEG(inputImage, buff, 20); err != nil {
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
