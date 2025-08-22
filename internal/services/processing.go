package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/encoder"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/logger"
	"github.com/barasher/go-exiftool"
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
	debug *logger.DebugLogger
}

func NewProcessingMediaService() (*ProcessingMediaService, error) {
	return &ProcessingMediaService{
		debug: logger.NewDebugLogger("processing_service"),
	}, nil
}

func (p *ProcessingMediaService) Process(ctx context.Context, content io.Reader) (io.Reader, map[string]string, error) {
	debug := p.debug.WithContext(ctx)
	tracer := debug.StartOperation("process_media").Build()

	// Initialize exiftool
	tracer.Step("exiftool_init").Log()

	start := time.Now()
	et, err := exiftool.NewExiftool()
	exiftoolInitDuration := time.Since(start)
	if err != nil {
		return nil, map[string]string{}, fmt.Errorf("failed to open exiftool: %s", err)
	}
	defer func() {
		et.Close()
	}()

	tracer.Performance("exiftool_init", exiftoolInitDuration)

	// Read content into memory
	tracer.Step("content_read").Log()

	start = time.Now()
	data, err := io.ReadAll(content)
	contentReadDuration := time.Since(start)
	if err != nil {
		return nil, map[string]string{}, err
	}

	debug.FileOperation("read_content", "memory", int64(len(data)), contentReadDuration)

	// Generate thumbnail
	tracer.Step("thumbnail_generation").
		WithInt("content_size", len(data)).
		Log()

	buff := bytes.NewBuffer([]byte{})
	start = time.Now()
	if err := p.generateThumbnail(bytes.NewReader(data), buff); err != nil {
		return nil, map[string]string{}, fmt.Errorf("failed to generate thumbnail: %w", err)
	}
	thumbnailDuration := time.Since(start)

	debug.Processing("thumbnail_generated", "memory").
		WithInt("original_size", len(data)).
		WithInt("thumbnail_size", buff.Len()).
		WithParam("duration", thumbnailDuration).
		Log()

	// Create temporary file for EXIF extraction
	tracer.Step("temp_file_creation").Log()

	start = time.Now()
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
	tempFileCreationDuration := time.Since(start)

	debug.FileOperation("create_temp_file", tmp.Name(), int64(len(data)), tempFileCreationDuration)

	// Extract EXIF metadata
	tracer.Step("exif_extraction").
		WithString("temp_file", tmp.Name()).
		Log()

	start = time.Now()
	fileInfos := et.ExtractMetadata(tmp.Name())
	exifExtractionDuration := time.Since(start)

	if len(fileInfos) == 0 {
		debug.Processing("no_exif_metadata_found", tmp.Name()).
			WithParam("extraction_duration", exifExtractionDuration).
			Log()
		tracer.Success().
			WithInt("thumbnail_size", buff.Len()).
			WithInt("exif_fields", 0).
			Log()
		return buff, map[string]string{}, nil
	}

	// Process EXIF fields
	tracer.Step("exif_processing").
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
			debug.Processing("unsupported_exif_value_type", tmp.Name()).
				WithString("key", k).
				WithString("value_type", fmt.Sprintf("%T", v)).
				WithParam("value", v).
				Log()
		}
	}

	debug.Processing("exif_processing_completed", tmp.Name()).
		WithInt("total_raw_fields", len(fileInfos[0].Fields)).
		WithInt("processed_fields", len(exif)).
		WithInt("ignored_fields", ignoredCount).
		WithInt("unsupported_fields", unsupportedCount).
		WithParam("extraction_duration", exifExtractionDuration).
		Log()

	tracer.Success().
		WithInt("thumbnail_size", buff.Len()).
		WithInt("exif_fields", len(exif)).
		Log()

	return buff, exif, nil
}

func (p *ProcessingMediaService) generateThumbnail(r io.Reader, w io.Writer) error {
	// This is a helper method, so we use a simple debug context
	debug := p.debug
	
	start := time.Now()
	
	// Encode thumbnail
	encodedReader, err := encoder.EncodeThumbnail(r)
	encodeStartDuration := time.Since(start)
	if err != nil {
		return err
	}

	// Read encoded data
	readStart := time.Now()
	data, err := io.ReadAll(encodedReader)
	readDuration := time.Since(readStart)
	if err != nil {
		return err
	}

	// Write to output
	writeStart := time.Now()
	_, err = w.Write(data)
	writeDuration := time.Since(writeStart)
	if err != nil {
		return err
	}

	totalDuration := time.Since(start)
	
	debug.Processing("thumbnail_generation_details", "internal").
		WithInt("thumbnail_size", len(data)).
		WithParam("encode_duration", encodeStartDuration).
		WithParam("read_duration", readDuration).
		WithParam("write_duration", writeDuration).
		WithParam("total_duration", totalDuration).
		Log()

	return nil
}
