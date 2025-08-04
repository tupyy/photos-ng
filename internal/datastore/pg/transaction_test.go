package pg_test

import (
	"context"
	"errors"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"github.com/jackc/pgx/v5/pgxpool"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transaction", Ordered, func() {
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

	Context("WriteTx", func() {
		It("commits successful transaction", func() {
			album := entity.NewAlbum("/path/to/success")
			album.Description = stringPtr("Successful Album")

			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteAlbum(ctx, album)
			})
			Expect(err).To(BeNil())

			// Verify the album was committed
			var count int
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM albums WHERE id = $1", album.ID).Scan(&count)
			Expect(err).To(BeNil())
			Expect(count).To(Equal(1))
		})

		It("rolls back failed transaction", func() {
			album1 := entity.NewAlbum("/path/to/rollback1")
			album1.Description = stringPtr("First Album")

			album2 := entity.NewAlbum("/path/to/rollback2")
			album2.Description = stringPtr("Second Album")

			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				// Write first album successfully
				err := w.WriteAlbum(ctx, album1)
				if err != nil {
					return err
				}

				// Write second album successfully
				err = w.WriteAlbum(ctx, album2)
				if err != nil {
					return err
				}

				// Force an error to trigger rollback
				return errors.New("forced error for rollback test")
			})
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("forced error for rollback test"))

			// Verify neither album was committed
			var count int
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM albums WHERE id IN ($1, $2)", album1.ID, album2.ID).Scan(&count)
			Expect(err).To(BeNil())
			Expect(count).To(Equal(0))
		})

		It("handles multiple operations in single transaction", func() {
			// Create album
			album := entity.NewAlbum("/path/to/multi-op")
			album.Description = stringPtr("Multi-operation Album")

			// Create media for the album
			media1 := entity.NewMedia("photo1.jpg", album)
			media1.CapturedAt = time.Now()
			media1.MediaType = entity.Photo
			media1.Exif = map[string]string{"Make": "Canon"}

			media2 := entity.NewMedia("photo2.jpg", album)
			media2.CapturedAt = time.Now()
			media2.MediaType = entity.Photo
			media2.Exif = map[string]string{"Make": "Nikon"}

			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				// Write album
				if err := w.WriteAlbum(ctx, album); err != nil {
					return err
				}

				// Write both media items
				if err := w.WriteMedia(ctx, media1); err != nil {
					return err
				}

				if err := w.WriteMedia(ctx, media2); err != nil {
					return err
				}

				return nil
			})
			Expect(err).To(BeNil())

			// Verify all operations were committed
			var albumCount, mediaCount int
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM albums WHERE id = $1", album.ID).Scan(&albumCount)
			Expect(err).To(BeNil())
			Expect(albumCount).To(Equal(1))

			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM media WHERE album_id = $1", album.ID).Scan(&mediaCount)
			Expect(err).To(BeNil())
			Expect(mediaCount).To(Equal(2))
		})

		It("handles partial failure in multi-operation transaction", func() {
			// Create album
			album := entity.NewAlbum("/path/to/partial-fail")
			album.Description = stringPtr("Partial Failure Album")

			// Create valid media
			validMedia := entity.NewMedia("valid.jpg", album)
			validMedia.CapturedAt = time.Now()
			validMedia.MediaType = entity.Photo
			validMedia.Exif = map[string]string{"Make": "Canon"}

			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				// Write album successfully
				if err := w.WriteAlbum(ctx, album); err != nil {
					return err
				}

				// Write valid media successfully
				if err := w.WriteMedia(ctx, validMedia); err != nil {
					return err
				}

				// Try to write media with invalid album reference (should cause rollback)
				invalidMedia := entity.NewMedia("invalid.jpg", entity.Album{ID: "non_existent_album"})
				invalidMedia.CapturedAt = time.Now()
				invalidMedia.MediaType = entity.Photo
				invalidMedia.Exif = map[string]string{"Make": "Invalid"}

				// This should fail due to foreign key constraint
				return w.WriteMedia(ctx, invalidMedia)
			})
			Expect(err).ToNot(BeNil())

			// Verify entire transaction was rolled back
			var albumCount, mediaCount int
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM albums WHERE id = $1", album.ID).Scan(&albumCount)
			Expect(err).To(BeNil())
			Expect(albumCount).To(Equal(0))

			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM media WHERE id = $1", validMedia.ID).Scan(&mediaCount)
			Expect(err).To(BeNil())
			Expect(mediaCount).To(Equal(0))
		})

		It("handles update operations in transaction", func() {
			// First, create an album outside the test transaction
			originalAlbum := entity.NewAlbum("/path/to/update-test")
			originalAlbum.Description = stringPtr("Original Description")

			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteAlbum(ctx, originalAlbum)
			})
			Expect(err).To(BeNil())

			// Now update it in a new transaction
			updatedAlbum := originalAlbum
			updatedAlbum.Description = stringPtr("Updated Description")

			err = dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				return w.WriteAlbum(ctx, updatedAlbum)
			})
			Expect(err).To(BeNil())

			// Verify the update was committed
			var description string
			err = pgPool.QueryRow(context.TODO(),
				"SELECT description FROM albums WHERE id = $1", originalAlbum.ID).
				Scan(&description)
			Expect(err).To(BeNil())
			Expect(description).To(Equal("Updated Description"))
		})

		It("handles delete operations in transaction", func() {
			// Create test data
			album := entity.NewAlbum("/path/to/delete-test")
			album.Description = stringPtr("To Be Deleted")

			media := entity.NewMedia("delete-me.jpg", album)
			media.CapturedAt = time.Now()
			media.MediaType = entity.Photo
			media.Exif = map[string]string{"Make": "Delete"}

			// Insert test data
			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				if err := w.WriteAlbum(ctx, album); err != nil {
					return err
				}
				return w.WriteMedia(ctx, media)
			})
			Expect(err).To(BeNil())

			// Delete in a transaction
			err = dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				// Delete media first (foreign key constraint)
				if err := w.DeleteMedia(ctx, media.ID); err != nil {
					return err
				}
				// Then delete album
				return w.DeleteAlbum(ctx, album.ID)
			})
			Expect(err).To(BeNil())

			// Verify deletion was committed
			var albumCount, mediaCount int
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM albums WHERE id = $1", album.ID).Scan(&albumCount)
			Expect(err).To(BeNil())
			Expect(albumCount).To(Equal(0))

			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM media WHERE id = $1", media.ID).Scan(&mediaCount)
			Expect(err).To(BeNil())
			Expect(mediaCount).To(Equal(0))
		})

		It("handles context cancellation", func() {
			album := entity.NewAlbum("/path/to/cancelled")
			album.Description = stringPtr("Cancelled Album")

			// Create a context that gets cancelled
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately

			err := dt.WriteTx(ctx, func(ctx context.Context, w *pg.Writer) error {
				return w.WriteAlbum(ctx, album)
			})
			Expect(err).ToNot(BeNil())

			// Verify transaction was not committed
			var count int
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM albums WHERE id = $1", album.ID).Scan(&count)
			Expect(err).To(BeNil())
			Expect(count).To(Equal(0))
		})

		It("handles nested transaction-like operations", func() {
			// Test multiple sequential operations that depend on each other
			parentAlbum := entity.NewAlbum("/path/to/parent")
			parentAlbum.Description = stringPtr("Parent Album")

			childAlbum := entity.NewAlbum("/path/to/parent/child")
			childAlbum.Description = stringPtr("Child Album")
			childAlbum.ParentId = &parentAlbum.ID

			media := entity.NewMedia("family-photo.jpg", childAlbum)
			media.CapturedAt = time.Now()
			media.MediaType = entity.Photo
			media.Exif = map[string]string{"Make": "Family"}

			err := dt.WriteTx(context.TODO(), func(ctx context.Context, w *pg.Writer) error {
				// Create parent first
				if err := w.WriteAlbum(ctx, parentAlbum); err != nil {
					return err
				}

				// Create child (depends on parent)
				if err := w.WriteAlbum(ctx, childAlbum); err != nil {
					return err
				}

				// Create media (depends on child album)
				if err := w.WriteMedia(ctx, media); err != nil {
					return err
				}

				// Update parent to use this media as thumbnail
				parentAlbum.Thumbnail = &media.ID
				return w.WriteAlbum(ctx, parentAlbum)
			})
			Expect(err).To(BeNil())

			// Verify all related data was committed properly
			var parentCount, childCount, mediaCount int
			var thumbnailID string

			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM albums WHERE id = $1", parentAlbum.ID).Scan(&parentCount)
			Expect(err).To(BeNil())
			Expect(parentCount).To(Equal(1))

			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM albums WHERE id = $1", childAlbum.ID).Scan(&childCount)
			Expect(err).To(BeNil())
			Expect(childCount).To(Equal(1))

			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM media WHERE id = $1", media.ID).Scan(&mediaCount)
			Expect(err).To(BeNil())
			Expect(mediaCount).To(Equal(1))

			err = pgPool.QueryRow(context.TODO(), "SELECT thumbnail_id FROM albums WHERE id = $1", parentAlbum.ID).Scan(&thumbnailID)
			Expect(err).To(BeNil())
			Expect(thumbnailID).To(Equal(media.ID))
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})
})
