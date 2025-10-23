package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/encoder"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/logger"
	"github.com/barasher/go-exiftool"
)

var ignoreExifKeys = map[string]any{
	"filename":        true,
	"directory":       true,
	"sourceFile":      true,
	"filepermissions": true,
}

type ProcessingMediaService struct {
	logger *logger.StructuredLogger
}

func NewProcessingMediaService() (*ProcessingMediaService, error) {
	return &ProcessingMediaService{
		logger: logger.New("processing_service"),
	}, nil
}

func (p *ProcessingMediaService) Process(ctx context.Context, content io.Reader) (io.Reader, map[string]string, error) {
	logger := p.logger.WithContext(ctx).Debug("process_media").Build()

	// Initialize exiftool
	logger.Step("exiftool_init").Log()

	et, err := exiftool.NewExiftool()
	if err != nil {
		return nil, map[string]string{}, fmt.Errorf("failed to open exiftool: %s", err)
	}
	defer func() {
		et.Close()
	}()

	// Read content into memory
	logger.Step("content_read").Log()

	data, err := io.ReadAll(content)
	if err != nil {
		return nil, map[string]string{}, err
	}

	// Generate thumbnail
	logger.Step("thumbnail_generation").
		WithInt("content_size", len(data)).
		Log()

	buff := bytes.NewBuffer([]byte{})
	if err := p.generateThumbnail(bytes.NewReader(data), buff); err != nil {
		return nil, map[string]string{}, fmt.Errorf("failed to generate thumbnail: %w", err)
	}

	logger.Step("thumbnail_generated").
		WithInt("original_size", len(data)).
		WithInt("thumbnail_size", buff.Len()).
		Log()

	// Create temporary file for EXIF extraction
	logger.Step("temp_file_creation").Log()

	tmp, err := os.CreateTemp("", "photo-")
	if err != nil {
		return nil, map[string]string{}, fmt.Errorf("failed to create temporary folder: %w", err)
	}
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
	}()

	// Copy content to temporary file
	if _, err = io.Copy(tmp, bytes.NewReader(data)); err != nil {
		return nil, map[string]string{}, fmt.Errorf("failed to copy photo content to temporary file: %w", err)
	}
	tmp.Close()

	// Extract EXIF metadata
	logger.Step("exif_extraction").
		WithString("temp_file", tmp.Name()).
		Log()

	fileInfos := et.ExtractMetadata(tmp.Name())

	if len(fileInfos) == 0 {
		logger.Step("no_exif_metadata_found").
			WithString("temp_file", tmp.Name()).
			Log()
		logger.Success().
			WithInt("thumbnail_size", buff.Len()).
			WithInt("exif_fields", 0).
			Log()
		return buff, map[string]string{}, nil
	}

	// Process EXIF fields
	logger.Step("exif_processing").
		WithInt("raw_fields", len(fileInfos[0].Fields)).
		Log()

	exif := make(map[string]string)
	ignoredCount := 0
	unsupportedCount := 0

	for k, v := range fileInfos[0].Fields {
		if _, toBeIgnored := ignoreExifKeys[strings.ToLower(k)]; toBeIgnored {
			ignoredCount++
			continue
		}
		switch val := v.(type) {
		case string:
			exif[k] = val
			continue
		case int:
			exif[k] = fmt.Sprintf("%d", val)
		case float32, float64:
			exif[k] = fmt.Sprintf("%f", val)
		default:
			unsupportedCount++
			logger.Step("unsupported_exif_value_type").
				WithString("temp_file", tmp.Name()).
				WithString("key", k).
				WithString("value_type", fmt.Sprintf("%T", v)).
				WithParam("value", v).
				Log()
		}
	}

	logger.Step("exif_processing_completed").
		WithString("temp_file", tmp.Name()).
		WithInt("total_raw_fields", len(fileInfos[0].Fields)).
		WithInt("processed_fields", len(exif)).
		WithInt("ignored_fields", ignoredCount).
		WithInt("unsupported_fields", unsupportedCount).
		Log()

	logger.Success().
		WithInt("thumbnail_size", buff.Len()).
		WithInt("exif_fields", len(exif)).
		Log()

	return buff, exif, nil
}

func (p *ProcessingMediaService) generateThumbnail(r io.Reader, w io.Writer) error {
	// Encode thumbnail
	encodedReader, err := encoder.EncodeThumbnail(r)
	if err != nil {
		return err
	}

	// Read encoded data
	data, err := io.ReadAll(encodedReader)
	if err != nil {
		return err
	}

	// Write to output
	_, err = w.Write(data)
	if err != nil {
		return err
	}

	return nil
}
