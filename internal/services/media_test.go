package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
	"github.com/jackc/pgx/v5/pgxpool"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MediaService", Ordered, func() {
	var (
		mediaService *services.MediaService
		dt           *pg.Datastore
		pgPool       *pgxpool.Pool
		testAlbum    entity.Album
		testMedia    entity.Media
	)

	BeforeAll(func() {
		pgDt, err := pg.NewPostgresDatastore(context.TODO(), pgUri)
		Expect(err).To(BeNil())
		Expect(pgDt).ToNot(BeNil())

		pgxConfig, err := pgxpool.ParseConfig(pgUri)
		Expect(err).To(BeNil())

		pool, err := pgxpool.NewWithConfig(context.TODO(), pgxConfig)
		Expect(err).To(BeNil())
		Expect(pool).ToNot(BeNil())

		dt = pgDt
		pgPool = pool

		// Clean up any existing data
		_, err = pgPool.Exec(context.TODO(), "DELETE FROM media;")
		Expect(err).To(BeNil())
		_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
		Expect(err).To(BeNil())
	})

	AfterAll(func() {
		dt.Close()
		// Clean up test directory
		os.RemoveAll("/tmp/photos-ng-test")
	})

	BeforeEach(func() {
		// Create a temporary directory for fs operations in tests
		tmpDir := "/tmp/photos-ng-test"
		// Clean up and recreate the temp directory for each test
		os.RemoveAll(tmpDir)
		err := os.MkdirAll(tmpDir, 0755)
		Expect(err).To(BeNil())

		realFs := fs.NewFsDatastore(tmpDir)
		mediaService = services.NewMediaService(dt, realFs)

		// Create test album first (required for media)
		testAlbum = entity.NewAlbum("/test/album")
		testAlbum.Description = stringPtr("Test Album for Media")

		// Create the album directory structure in the filesystem
		err = realFs.CreateFolder(context.TODO(), testAlbum.Path)
		Expect(err).To(BeNil())

		now := time.Now()
		_, err = pgPool.Exec(context.TODO(),
			"INSERT INTO albums (id, created_at, path, description, parent_id, thumbnail_id) VALUES ($1, $2, $3, $4, $5, $6)",
			testAlbum.ID, now, testAlbum.Path, testAlbum.Description, nil, nil)
		Expect(err).To(BeNil())

		// Create test media
		testMedia = entity.NewMedia("test-photo.jpg", testAlbum)
		testMedia.CapturedAt = time.Now()
		testMedia.MediaType = entity.Photo
		testMedia.Thumbnail = []byte("test-thumbnail-data")
		testMedia.Exif = map[string]string{
			"Make":       "Canon",
			"Model":      "EOS R5",
			"CreateDate": "2024:01:15 10:30:00",
		}
		testMedia.Content = func() (io.Reader, error) {
			return strings.NewReader("test-image-content"), nil
		}
	})

	Context("GetMedia", func() {
		It("retrieves media successfully", func() {
			// Insert test media
			exifJSON, err := json.Marshal(testMedia.Exif)
			Expect(err).To(BeNil())

			now := time.Now()
			sql, args, err := insertMediaStmt.
				Values(testMedia.ID, now, testMedia.CapturedAt, testMedia.Album.ID, testMedia.Filename, testMedia.Thumbnail, exifJSON, string(testMedia.MediaType)).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			options := &services.MediaOptions{MediaLimit: 10}
			media, err := mediaService.GetMedia(context.TODO(), options)

			Expect(err).To(BeNil())
			Expect(media).To(HaveLen(1))
			Expect(media[0].ID).To(Equal(testMedia.ID))
			Expect(media[0].Filename).To(Equal(testMedia.Filename))
			Expect(media[0].MediaType).To(Equal(testMedia.MediaType))
		})

		It("returns empty result when no media exist", func() {
			options := &services.MediaOptions{MediaLimit: 10}
			media, err := mediaService.GetMedia(context.TODO(), options)

			Expect(err).To(BeNil())
			Expect(media).To(HaveLen(0))
		})

		It("applies filter options correctly", func() {
			// Insert multiple test media
			exifJSON, err := json.Marshal(testMedia.Exif)
			Expect(err).To(BeNil())

			now := time.Now()
			media1 := entity.NewMedia("photo1.jpg", testAlbum)
			media2 := entity.NewMedia("photo2.jpg", testAlbum)
			media3 := entity.NewMedia("video1.mp4", testAlbum)
			media3.MediaType = entity.Video

			_, err = pgPool.Exec(context.TODO(),
				"INSERT INTO media (id, created_at, captured_at, album_id, file_name, thumbnail, exif, media_type) VALUES ($1, $2, $3, $4, $5, $6, $7, $8), ($9, $10, $11, $12, $13, $14, $15, $16), ($17, $18, $19, $20, $21, $22, $23, $24)",
				media1.ID, now, now, testAlbum.ID, media1.Filename, []byte("thumb1"), exifJSON, string(entity.Photo),
				media2.ID, now, now, testAlbum.ID, media2.Filename, []byte("thumb2"), exifJSON, string(entity.Photo),
				media3.ID, now, now, testAlbum.ID, media3.Filename, []byte("thumb3"), exifJSON, string(entity.Video))
			Expect(err).To(BeNil())

			// Test limit only (no cursor for first page)
			options := &services.MediaOptions{
				MediaLimit: 2,
			}
			media, err := mediaService.GetMedia(context.TODO(), options)

			Expect(err).To(BeNil())
			Expect(media).To(HaveLen(2))

			// Test media type filter
			photoType := string(entity.Photo)
			options = &services.MediaOptions{
				MediaType: &photoType,
			}
			media, err = mediaService.GetMedia(context.TODO(), options)

			Expect(err).To(BeNil())
			Expect(media).To(HaveLen(2)) // Should only return photos
			for _, m := range media {
				Expect(m.MediaType).To(Equal(entity.Photo))
			}
		})

		It("applies date filtering", func() {
			// Insert media with different dates
			exifJSON, err := json.Marshal(testMedia.Exif)
			Expect(err).To(BeNil())

			now := time.Now()
			yesterday := now.Add(-24 * time.Hour)
			tomorrow := now.Add(24 * time.Hour)

			media1 := entity.NewMedia("old-photo.jpg", testAlbum)
			media2 := entity.NewMedia("new-photo.jpg", testAlbum)

			_, err = pgPool.Exec(context.TODO(),
				"INSERT INTO media (id, created_at, captured_at, album_id, file_name, thumbnail, exif, media_type) VALUES ($1, $2, $3, $4, $5, $6, $7, $8), ($9, $10, $11, $12, $13, $14, $15, $16)",
				media1.ID, now, yesterday, testAlbum.ID, media1.Filename, []byte("thumb1"), exifJSON, string(entity.Photo),
				media2.ID, now, tomorrow, testAlbum.ID, media2.Filename, []byte("thumb2"), exifJSON, string(entity.Photo))
			Expect(err).To(BeNil())

			// Filter by start date (should only get new photo)
			options := &services.MediaOptions{
				StartDate: &now,
			}
			media, err := mediaService.GetMedia(context.TODO(), options)

			Expect(err).To(BeNil())
			Expect(media).To(HaveLen(1))
			Expect(media[0].Filename).To(Equal("new-photo.jpg"))
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})

	Context("GetMediaByID", func() {
		It("retrieves single media successfully", func() {
			// Insert test media
			exifJSON, err := json.Marshal(testMedia.Exif)
			Expect(err).To(BeNil())

			now := time.Now()
			_, err = pgPool.Exec(context.TODO(),
				"INSERT INTO media (id, created_at, captured_at, album_id, file_name, thumbnail, exif, media_type) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
				testMedia.ID, now, testMedia.CapturedAt, testMedia.Album.ID, testMedia.Filename, testMedia.Thumbnail, exifJSON, string(testMedia.MediaType))
			Expect(err).To(BeNil())

			media, err := mediaService.GetMediaByID(context.TODO(), testMedia.ID)

			Expect(err).To(BeNil())
			Expect(media).ToNot(BeNil())
			Expect(media.ID).To(Equal(testMedia.ID))
			Expect(media.Filename).To(Equal(testMedia.Filename))
			Expect(media.Exif).To(HaveKeyWithValue("Make", "Canon"))
		})

		It("returns not found error when media doesn't exist", func() {
			media, err := mediaService.GetMediaByID(context.TODO(), "non-existent-id")

			Expect(err).ToNot(BeNil())
			Expect(services.IsErrResourceNotFound(err)).To(BeTrue())
			Expect(media).To(BeNil())
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})

	Context("WriteMedia", func() {
		It("creates new media successfully", func() {
			result, err := mediaService.WriteMedia(context.TODO(), testMedia)

			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())
			Expect(result.ID).To(Equal(testMedia.ID))

			// Verify media was created in database
			var count int
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM media WHERE id = $1", testMedia.ID).Scan(&count)
			Expect(err).To(BeNil())
			Expect(count).To(Equal(1))

			// Verify media data
			var filename, mediaType string
			var exifData []byte
			err = pgPool.QueryRow(context.TODO(),
				"SELECT file_name, media_type, exif FROM media WHERE id = $1", testMedia.ID).
				Scan(&filename, &mediaType, &exifData)
			Expect(err).To(BeNil())
			Expect(filename).To(Equal(testMedia.Filename))
			Expect(mediaType).To(Equal(string(testMedia.MediaType)))

			var exif map[string]string
			err = json.Unmarshal(exifData, &exif)
			Expect(err).To(BeNil())
			Expect(exif).To(HaveKeyWithValue("Make", "Canon"))
		})

		It("updates existing media successfully", func() {
			// First create the media
			_, err := mediaService.WriteMedia(context.TODO(), testMedia)
			Expect(err).To(BeNil())

			// Update media properties
			updatedMedia := testMedia
			updatedMedia.Exif = map[string]string{
				"Make":  "Nikon",
				"Model": "D850",
			}
			updatedMedia.Thumbnail = []byte("updated-thumbnail")

			result, err := mediaService.WriteMedia(context.TODO(), updatedMedia)

			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())

			// Verify update in database
			var exifData []byte
			var thumbnail []byte
			err = pgPool.QueryRow(context.TODO(),
				"SELECT exif, thumbnail FROM media WHERE id = $1", testMedia.ID).
				Scan(&exifData, &thumbnail)
			Expect(err).To(BeNil())

			var exif map[string]string
			err = json.Unmarshal(exifData, &exif)
			Expect(err).To(BeNil())
			Expect(exif).To(HaveKeyWithValue("Make", "Nikon"))
			Expect(thumbnail).To(Equal([]byte("updated-thumbnail")))
		})

		It("handles media with invalid album", func() {
			invalidMedia := testMedia
			invalidMedia.Album.ID = "non-existent-album"

			result, err := mediaService.WriteMedia(context.TODO(), invalidMedia)

			Expect(err).ToNot(BeNil()) // Should fail due to foreign key constraint
			Expect(result).To(BeNil())
		})

		It("handles file content writing", func() {
			mediaWithContent := testMedia
			mediaWithContent.Content = func() (io.Reader, error) {
				return strings.NewReader("test-file-content"), nil
			}

			result, err := mediaService.WriteMedia(context.TODO(), mediaWithContent)

			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())

			// Verify media was created in database
			var count int
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM media WHERE id = $1", mediaWithContent.ID).Scan(&count)
			Expect(err).To(BeNil())
			Expect(count).To(Equal(1))
		})

		It("handles content reading failure", func() {
			mediaWithBadContent := testMedia
			mediaWithBadContent.Content = func() (io.Reader, error) {
				return nil, errors.New("content read error")
			}

			result, err := mediaService.WriteMedia(context.TODO(), mediaWithBadContent)

			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("content read error"))
			Expect(result).To(BeNil())

			// Verify transaction was rolled back - media shouldn't exist
			var count int
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM media WHERE id = $1", mediaWithBadContent.ID).Scan(&count)
			Expect(err).To(BeNil())
			Expect(count).To(Equal(0))
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})

	Context("UpdateMedia", func() {
		It("updates existing media successfully", func() {
			// First create the media
			exifJSON, err := json.Marshal(testMedia.Exif)
			Expect(err).To(BeNil())

			now := time.Now()
			_, err = pgPool.Exec(context.TODO(),
				"INSERT INTO media (id, created_at, captured_at, album_id, file_name, thumbnail, exif, media_type) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
				testMedia.ID, now, testMedia.CapturedAt, testMedia.Album.ID, testMedia.Filename, testMedia.Thumbnail, exifJSON, string(testMedia.MediaType))
			Expect(err).To(BeNil())

			// Update media metadata
			updatedMedia := testMedia
			updatedMedia.Exif = map[string]string{
				"Make":        "Sony",
				"Model":       "A7R IV",
				"ISO":         "800",
				"FocalLength": "85mm",
			}
			updatedTime := time.Now().Add(time.Hour)
			updatedMedia.CapturedAt = updatedTime

			result, err := mediaService.UpdateMedia(context.TODO(), updatedMedia)

			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())

			// Verify update in database
			var exifData []byte
			err = pgPool.QueryRow(context.TODO(), "SELECT exif FROM media WHERE id = $1", testMedia.ID).Scan(&exifData)
			Expect(err).To(BeNil())

			var exif map[string]string
			err = json.Unmarshal(exifData, &exif)
			Expect(err).To(BeNil())
			Expect(exif).To(HaveKeyWithValue("Make", "Sony"))
			Expect(exif).To(HaveKeyWithValue("Model", "A7R IV"))
			Expect(exif).To(HaveKeyWithValue("ISO", "800"))
		})

		It("clears content function during update", func() {
			// First create the media
			exifJSON, err := json.Marshal(testMedia.Exif)
			Expect(err).To(BeNil())

			now := time.Now()
			_, err = pgPool.Exec(context.TODO(),
				"INSERT INTO media (id, created_at, captured_at, album_id, file_name, thumbnail, exif, media_type) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
				testMedia.ID, now, testMedia.CapturedAt, testMedia.Album.ID, testMedia.Filename, testMedia.Thumbnail, exifJSON, string(testMedia.MediaType))
			Expect(err).To(BeNil())

			// Update with content function (should be cleared)
			updatedMedia := testMedia
			updatedMedia.Content = func() (io.Reader, error) {
				return strings.NewReader("should-not-be-written"), nil
			}
			updatedMedia.Exif = map[string]string{"Updated": "true"}

			result, err := mediaService.UpdateMedia(context.TODO(), updatedMedia)

			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())
			Expect(result.Content).To(BeNil()) // Content should be cleared
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})

	Context("DeleteMedia", func() {
		It("deletes existing media successfully", func() {
			// First create the media
			exifJSON, err := json.Marshal(testMedia.Exif)
			Expect(err).To(BeNil())

			now := time.Now()
			_, err = pgPool.Exec(context.TODO(),
				"INSERT INTO media (id, created_at, captured_at, album_id, file_name, thumbnail, exif, media_type) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
				testMedia.ID, now, testMedia.CapturedAt, testMedia.Album.ID, testMedia.Filename, testMedia.Thumbnail, exifJSON, string(testMedia.MediaType))
			Expect(err).To(BeNil())

			err = mediaService.DeleteMedia(context.TODO(), testMedia.ID)

			Expect(err).To(BeNil())

			// Verify media was deleted from database
			var count int
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM media WHERE id = $1", testMedia.ID).Scan(&count)
			Expect(err).To(BeNil())
			Expect(count).To(Equal(0))
		})

		It("returns error when media doesn't exist", func() {
			err := mediaService.DeleteMedia(context.TODO(), "non-existent-id")

			Expect(err).ToNot(BeNil())
			Expect(services.IsErrResourceNotFound(err)).To(BeTrue())
		})

		It("handles deletion of media used as album thumbnail", func() {
			// First create media and set it as album thumbnail
			exifJSON, err := json.Marshal(testMedia.Exif)
			Expect(err).To(BeNil())

			now := time.Now()
			_, err = pgPool.Exec(context.TODO(),
				"INSERT INTO media (id, created_at, captured_at, album_id, file_name, thumbnail, exif, media_type) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
				testMedia.ID, now, testMedia.CapturedAt, testMedia.Album.ID, testMedia.Filename, testMedia.Thumbnail, exifJSON, string(testMedia.MediaType))
			Expect(err).To(BeNil())

			// Update album to use this media as thumbnail
			_, err = pgPool.Exec(context.TODO(), "UPDATE albums SET thumbnail_id = $1 WHERE id = $2", testMedia.ID, testAlbum.ID)
			Expect(err).To(BeNil())

			// Delete the media (should succeed due to ON DELETE SET NULL)
			err = mediaService.DeleteMedia(context.TODO(), testMedia.ID)

			Expect(err).To(BeNil())

			// Verify media was deleted
			var count int
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM media WHERE id = $1", testMedia.ID).Scan(&count)
			Expect(err).To(BeNil())
			Expect(count).To(Equal(0))

			// Verify album thumbnail was set to NULL
			var thumbnailID *string
			err = pgPool.QueryRow(context.TODO(), "SELECT thumbnail_id FROM albums WHERE id = $1", testAlbum.ID).Scan(&thumbnailID)
			Expect(err).To(BeNil())
			Expect(thumbnailID).To(BeNil())
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})

	Context("Integration Tests", func() {
		It("handles complete media lifecycle", func() {
			// 1. Create media
			result, err := mediaService.WriteMedia(context.TODO(), testMedia)
			Expect(err).To(BeNil())
			Expect(result.ID).To(Equal(testMedia.ID))

			// 2. Retrieve by ID
			retrieved, err := mediaService.GetMediaByID(context.TODO(), testMedia.ID)
			Expect(err).To(BeNil())
			Expect(retrieved.ID).To(Equal(testMedia.ID))
			Expect(retrieved.Filename).To(Equal(testMedia.Filename))

			// 3. Update metadata
			updatedMedia := *retrieved
			updatedMedia.Exif = map[string]string{
				"Make":    "Updated Make",
				"Model":   "Updated Model",
				"Updated": "true",
			}

			updated, err := mediaService.UpdateMedia(context.TODO(), updatedMedia)
			Expect(err).To(BeNil())
			Expect(updated.Exif).To(HaveKeyWithValue("Updated", "true"))

			// 4. Verify update persisted
			retrieved2, err := mediaService.GetMediaByID(context.TODO(), testMedia.ID)
			Expect(err).To(BeNil())
			Expect(retrieved2.Exif).To(HaveKeyWithValue("Updated", "true"))

			// 5. Delete media
			err = mediaService.DeleteMedia(context.TODO(), testMedia.ID)
			Expect(err).To(BeNil())

			// 6. Verify deletion
			_, err = mediaService.GetMediaByID(context.TODO(), testMedia.ID)
			Expect(err).ToNot(BeNil())
			Expect(services.IsErrResourceNotFound(err)).To(BeTrue())
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})
})
