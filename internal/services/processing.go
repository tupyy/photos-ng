package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/encoder"
	"github.com/barasher/go-exiftool"
	"go.uber.org/zap"
)

var (
	ignoreExifKeys = map[string]any{
		"filename":        true,
		"directory":       true,
		"sourceFile":      true,
		"filepermissions": true,
	}
)

type ProcessingMediaService struct {
}

func NewProcessingMediaService() (*ProcessingMediaService, error) {
	return &ProcessingMediaService{}, nil
}

func (p *ProcessingMediaService) Process(ctx context.Context, content io.Reader) (io.Reader, map[string]string, error) {
	et, err := exiftool.NewExiftool()
	if err != nil {
		return nil, map[string]string{}, fmt.Errorf("failed to open exiftool: %s", err)
	}
	defer func() {
		et.Close()
	}()

	data, err := io.ReadAll(content)
	if err != nil {
		return nil, map[string]string{}, err
	}
	// generate thumbnail
	buff := bytes.NewBuffer([]byte{})
	if err := p.generateThumbnail(bytes.NewReader(data), buff); err != nil {
		return nil, map[string]string{}, fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	// write the photo a tmp file
	tmp, err := os.CreateTemp("", "photo-")
	if err != nil {
		return nil, map[string]string{}, fmt.Errorf("failed to create temporary folder: %w", err)
	}
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
	}()

	if _, err = io.Copy(tmp, bytes.NewReader(data)); err != nil {
		return nil, map[string]string{}, fmt.Errorf("failed to copy photo content to temporary file: %w", err)
	}
	tmp.Close()

	// extract exif
	fileInfos := et.ExtractMetadata(tmp.Name())

	exif := make(map[string]string)
	for k, v := range fileInfos[0].Fields {
		if value, ok := v.(string); ok {
			if _, toBeIgnored := ignoreExifKeys[strings.ToLower(k)]; toBeIgnored {
				continue
			}
			exif[k] = value
			continue
		}
		zap.S().Warnw("failed to read exif metadata value", "error", "value is not a string")
	}

	return buff, exif, nil
}

func (p *ProcessingMediaService) generateThumbnail(r io.Reader, w io.Writer) error {
	r, err := encoder.EncodeThumbnail(r)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	if err != nil {
		return err
	}

	return nil
}
