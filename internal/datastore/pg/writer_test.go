package pg_test

import (
	"context"
	"encoding/json"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"github.com/jackc/pgx/v5/pgxpool"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Writer", Ordered, func() {
	var (
		dt     *pg.Datastore
		pgPool *pgxpool.Pool
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
	})

	Context("WriteAlbum", func() {
		It("write album successfully -- new album", func() {
			album := entity.NewAlbum("/path/to/new/album")
			album.Description = stringPtr("New Album Description")

			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteAlbum(ctx, album)
			})
			Expect(err).To(BeNil())

			// Verify album was inserted
			var count int
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM albums WHERE id = $1", album.ID).Scan(&count)
			Expect(err).To(BeNil())
			Expect(count).To(Equal(1))

			// Verify album data
			var path, description string
			err = pgPool.QueryRow(context.TODO(),
				"SELECT path, description FROM albums WHERE id = $1", album.ID).Scan(&path, &description)
			Expect(err).To(BeNil())
			Expect(path).To(Equal("/path/to/new/album"))
			Expect(description).To(Equal("New Album Description"))
		})

		It("update album successfully -- existing album", func() {
			// First insert an album
			album := entity.NewAlbum("/path/to/original")
			album.Description = stringPtr("Original Description")

			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteAlbum(ctx, album)
			})
			Expect(err).To(BeNil())

			// Create a media item that can be used as thumbnail
			media := entity.NewMedia("thumbnail.jpg", album)
			media.CapturedAt = time.Now()
			media.MediaType = entity.Photo
			media.Exif = map[string]string{"Make": "Canon"}

			err = dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteMedia(ctx, media)
			})
			Expect(err).To(BeNil())

			// Update the album with the media ID as thumbnail
			album.Description = stringPtr("Updated Description")
			album.Thumbnail = &media.ID

			err = dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteAlbum(ctx, album)
			})
			Expect(err).To(BeNil())

			// Verify update
			var description, thumbnailID string
			err = pgPool.QueryRow(context.TODO(),
				"SELECT description, thumbnail_id FROM albums WHERE id = $1", album.ID).
				Scan(&description, &thumbnailID)
			Expect(err).To(BeNil())
			Expect(description).To(Equal("Updated Description"))
			Expect(thumbnailID).To(Equal(media.ID))
		})

		It("write album with parent relationship", func() {
			// Create parent album
			parentAlbum := entity.NewAlbum("/path/to/parent")
			parentAlbum.Description = stringPtr("Parent Album")

			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteAlbum(ctx, parentAlbum)
			})
			Expect(err).To(BeNil())

			// Create child album
			childAlbum := entity.NewAlbum("/path/to/parent/child")
			childAlbum.Description = stringPtr("Child Album")
			childAlbum.ParentId = &parentAlbum.ID

			err = dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteAlbum(ctx, childAlbum)
			})
			Expect(err).To(BeNil())

			// Verify relationship
			var parentID string
			err = pgPool.QueryRow(context.TODO(),
				"SELECT parent_id FROM albums WHERE id = $1", childAlbum.ID).Scan(&parentID)
			Expect(err).To(BeNil())
			Expect(parentID).To(Equal(parentAlbum.ID))
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})

	Context("WriteMedia", func() {
		var testAlbum entity.Album

		BeforeEach(func() {
			// Create a test album for media
			testAlbum = entity.NewAlbum("/path/to/test/album")
			testAlbum.Description = stringPtr("Test Album for Media")

			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteAlbum(ctx, testAlbum)
			})
			Expect(err).To(BeNil())
		})

		It("write media successfully -- new media", func() {
			media := entity.NewMedia("photo1.jpg", testAlbum)
			media.CapturedAt = time.Now()
			media.MediaType = entity.Photo
			media.Thumbnail = []byte("thumbnail_data")
			media.Exif = map[string]string{
				"Make":       "Canon",
				"Model":      "EOS R5",
				"CreateDate": "2024:01:15 10:30:00",
			}

			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteMedia(ctx, media)
			})
			Expect(err).To(BeNil())

			// Verify media was inserted
			var count int
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM media WHERE id = $1", media.ID).Scan(&count)
			Expect(err).To(BeNil())
			Expect(count).To(Equal(1))

			// Verify media data
			var filename, mediaType, albumID string
			var exifData []byte
			err = pgPool.QueryRow(context.TODO(),
				"SELECT file_name, media_type, album_id, exif FROM media WHERE id = $1", media.ID).
				Scan(&filename, &mediaType, &albumID, &exifData)
			Expect(err).To(BeNil())
			Expect(filename).To(Equal("photo1.jpg"))
			Expect(mediaType).To(Equal("photo"))
			Expect(albumID).To(Equal(testAlbum.ID))

			// Verify EXIF data
			var exif map[string]string
			err = json.Unmarshal(exifData, &exif)
			Expect(err).To(BeNil())
			Expect(exif).To(HaveKeyWithValue("Make", "Canon"))
			Expect(exif).To(HaveKeyWithValue("Model", "EOS R5"))
		})

		It("update media successfully -- existing media", func() {
			media := entity.NewMedia("photo1.jpg", testAlbum)
			media.CapturedAt = time.Now()
			media.MediaType = entity.Photo
			media.Thumbnail = []byte("original_thumbnail")
			media.Exif = map[string]string{"Make": "Canon"}

			// Insert original media
			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteMedia(ctx, media)
			})
			Expect(err).To(BeNil())

			// Update media
			updatedTime := time.Now().Add(time.Hour)
			media.CapturedAt = updatedTime
			media.Thumbnail = []byte("updated_thumbnail")
			media.Exif = map[string]string{
				"Make":  "Canon",
				"Model": "EOS R6",
			}

			err = dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteMedia(ctx, media)
			})
			Expect(err).To(BeNil())

			// Verify update
			var thumbnail []byte
			var exifData []byte
			err = pgPool.QueryRow(context.TODO(),
				"SELECT thumbnail, exif FROM media WHERE id = $1", media.ID).
				Scan(&thumbnail, &exifData)
			Expect(err).To(BeNil())
			Expect(thumbnail).To(Equal([]byte("updated_thumbnail")))

			var exif map[string]string
			err = json.Unmarshal(exifData, &exif)
			Expect(err).To(BeNil())
			Expect(exif).To(HaveKeyWithValue("Model", "EOS R6"))
		})

		It("write video media successfully", func() {
			media := entity.NewMedia("video1.mp4", testAlbum)
			media.CapturedAt = time.Now()
			media.MediaType = entity.Video
			media.Thumbnail = []byte("video_thumbnail")
			media.Exif = map[string]string{
				"Duration": "00:02:30",
				"Format":   "H.264",
			}

			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteMedia(ctx, media)
			})
			Expect(err).To(BeNil())

			// Verify video media
			var mediaType string
			var exifData []byte
			err = pgPool.QueryRow(context.TODO(),
				"SELECT media_type, exif FROM media WHERE id = $1", media.ID).
				Scan(&mediaType, &exifData)
			Expect(err).To(BeNil())
			Expect(mediaType).To(Equal("Video"))

			var exif map[string]string
			err = json.Unmarshal(exifData, &exif)
			Expect(err).To(BeNil())
			Expect(exif).To(HaveKeyWithValue("Duration", "00:02:30"))
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})

	Context("DeleteAlbum", func() {
		It("delete album successfully", func() {
			// Insert album first
			album := entity.NewAlbum("/path/to/deletable")
			album.Description = stringPtr("Deletable Album")

			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteAlbum(ctx, album)
			})
			Expect(err).To(BeNil())

			// Verify album exists
			var count int
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM albums WHERE id = $1", album.ID).Scan(&count)
			Expect(err).To(BeNil())
			Expect(count).To(Equal(1))

			// Delete album
			err = dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.DeleteAlbum(ctx, album.ID)
			})
			Expect(err).To(BeNil())

			// Verify album is deleted
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM albums WHERE id = $1", album.ID).Scan(&count)
			Expect(err).To(BeNil())
			Expect(count).To(Equal(0))
		})

		It("delete album -- album not found", func() {
			// Delete non-existent album (should not error)
			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.DeleteAlbum(ctx, "non_existent_id")
			})
			Expect(err).To(BeNil())
		})

		It("delete album cascades to children and media", func() {
			// Create parent album
			parentAlbum := entity.NewAlbum("/path/to/parent")
			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteAlbum(ctx, parentAlbum)
			})
			Expect(err).To(BeNil())

			// Create child album
			childAlbum := entity.NewAlbum("/path/to/parent/child")
			childAlbum.ParentId = &parentAlbum.ID
			err = dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteAlbum(ctx, childAlbum)
			})
			Expect(err).To(BeNil())

			// Create media in parent album
			media := entity.NewMedia("photo.jpg", parentAlbum)
			media.Exif = map[string]string{"Make": "Canon"}
			err = dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteMedia(ctx, media)
			})
			Expect(err).To(BeNil())

			// Delete parent album
			err = dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.DeleteAlbum(ctx, parentAlbum.ID)
			})
			Expect(err).To(BeNil())

			// Verify cascading deletion
			var albumCount, mediaCount int
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM albums").Scan(&albumCount)
			Expect(err).To(BeNil())
			Expect(albumCount).To(Equal(0))

			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM media").Scan(&mediaCount)
			Expect(err).To(BeNil())
			Expect(mediaCount).To(Equal(0))
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})

	Context("DeleteMedia", func() {
		var testAlbum entity.Album

		BeforeEach(func() {
			// Create test album
			testAlbum = entity.NewAlbum("/path/to/test/album")
			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteAlbum(ctx, testAlbum)
			})
			Expect(err).To(BeNil())
		})

		It("delete media successfully", func() {
			// Insert media first
			media := entity.NewMedia("deletable.jpg", testAlbum)
			media.Exif = map[string]string{"Make": "Canon"}

			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteMedia(ctx, media)
			})
			Expect(err).To(BeNil())

			// Verify media exists
			var count int
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM media WHERE id = $1", media.ID).Scan(&count)
			Expect(err).To(BeNil())
			Expect(count).To(Equal(1))

			// Delete media
			err = dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.DeleteMedia(ctx, media.ID)
			})
			Expect(err).To(BeNil())

			// Verify media is deleted
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM media WHERE id = $1", media.ID).Scan(&count)
			Expect(err).To(BeNil())
			Expect(count).To(Equal(0))
		})

		It("delete media -- media not found", func() {
			// Delete non-existent media (should not error)
			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.DeleteMedia(ctx, "non_existent_id")
			})
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})
})

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
