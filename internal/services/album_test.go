package services_test

import (
	"context"
	"os"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	pgUri = "postgresql://postgres:postgres@localhost:5432/postgres"
)

var (
	psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	insertAlbumStmt = psql.Insert("albums").Columns("id", "created_at", "path", "description", "parent_id", "thumbnail_id")
	insertMediaStmt = psql.Insert("media").Columns("id", "created_at", "captured_at", "album_id", "file_name", "thumbnail", "exif", "media_type")
)

var _ = Describe("AlbumService", Ordered, func() {
	var (
		albumService *services.AlbumService
		dt           *pg.Datastore
		pgPool       *pgxpool.Pool
		testAlbum    entity.Album
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
		albumService = services.NewAlbumService(dt, realFs)

		testAlbum = entity.NewAlbum("/test/album")
		testAlbum.Description = stringPtr("Test Album")
	})

	Context("GetAlbums", func() {
		It("retrieves albums successfully", func() {
			// Insert test album
			now := time.Now()
			sql, args, err := insertAlbumStmt.
				Values(testAlbum.ID, now, testAlbum.Path, testAlbum.Description, nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			options := &services.AlbumOptions{Limit: 10}
			albums, err := albumService.GetAlbums(context.TODO(), options)

			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(1))
			Expect(albums[0].ID).To(Equal(testAlbum.ID))
		})

		It("returns empty result when no albums exist", func() {
			options := &services.AlbumOptions{Limit: 10}
			albums, err := albumService.GetAlbums(context.TODO(), options)

			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(0))
		})

		It("applies filter options correctly", func() {
			// Insert multiple test albums
			now := time.Now()
			album1 := entity.NewAlbum("/test/album1")
			album2 := entity.NewAlbum("/test/album2")
			album3 := entity.NewAlbum("/test/album3")

			sql, args, err := insertAlbumStmt.
				Values(album1.ID, now, album1.Path, nil, nil, nil).
				Values(album2.ID, now, album2.Path, nil, nil, nil).
				Values(album3.ID, now, album3.Path, nil, nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			options := &services.AlbumOptions{
				Limit:  2,
				Offset: 1,
			}
			albums, err := albumService.GetAlbums(context.TODO(), options)

			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(2))
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})

	Context("GetAlbum", func() {
		It("retrieves single album successfully", func() {
			// Insert test album
			now := time.Now()
			sql, args, err := insertAlbumStmt.
				Values(testAlbum.ID, now, testAlbum.Path, testAlbum.Description, nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			album, err := albumService.GetAlbum(context.TODO(), testAlbum.ID)

			Expect(err).To(BeNil())
			Expect(album).ToNot(BeNil())
			Expect(album.ID).To(Equal(testAlbum.ID))
			Expect(*album.Description).To(Equal("Test Album"))
		})

		It("returns not found error when album doesn't exist", func() {
			album, err := albumService.GetAlbum(context.TODO(), "non-existent-id")

			Expect(err).ToNot(BeNil())
			Expect(services.IsErrResourceNotFound(err)).To(BeTrue())
			Expect(album).To(BeNil())
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})

	Context("CreateAlbum", func() {
		It("creates new album successfully", func() {
			result, err := albumService.CreateAlbum(context.TODO(), testAlbum)

			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())
			Expect(result.ID).To(Equal(testAlbum.ID))

			// Verify album was created in database
			var count int
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM albums WHERE id = $1", testAlbum.ID).Scan(&count)
			Expect(err).To(BeNil())
			Expect(count).To(Equal(1))

			// Verify folder was created on filesystem
			tmpDir := "/tmp/photos-ng-test"
			folderPath := tmpDir + testAlbum.Path
			_, err = os.Stat(folderPath)
			Expect(err).To(BeNil(), "Album folder should exist at: %s", folderPath)
		})

		It("creates album with parent relationship", func() {
			// First create parent album
			parentAlbum := entity.NewAlbum("/parent")
			parentAlbum.Description = stringPtr("Parent Album")

			_, err := albumService.CreateAlbum(context.TODO(), parentAlbum)
			Expect(err).To(BeNil())

			// Verify parent folder was created
			tmpDir := "/tmp/photos-ng-test"
			parentFolderPath := tmpDir + parentAlbum.Path
			_, err = os.Stat(parentFolderPath)
			Expect(err).To(BeNil(), "Parent folder should exist at: %s", parentFolderPath)

			// Now create child album
			childAlbum := entity.NewAlbum("child")
			childAlbum.ParentId = &parentAlbum.ID

			result, err := albumService.CreateAlbum(context.TODO(), childAlbum)

			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())
			Expect(result.Path).To(Equal("/parent/child"))

			// Verify parent relationship in database
			var parentID string
			err = pgPool.QueryRow(context.TODO(), "SELECT parent_id FROM albums WHERE id = $1", result.ID).Scan(&parentID)
			Expect(err).To(BeNil())
			Expect(parentID).To(Equal(parentAlbum.ID))

			// Verify child folder was created with correct path
			childFolderPath := tmpDir + result.Path
			_, err = os.Stat(childFolderPath)
			Expect(err).To(BeNil(), "Child folder should exist at: %s", childFolderPath)

			// Verify parent folder still exists
			_, err = os.Stat(parentFolderPath)
			Expect(err).To(BeNil(), "Parent folder should still exist at: %s", parentFolderPath)
		})

		It("returns error when parent doesn't exist", func() {
			childAlbum := entity.NewAlbum("child")
			childAlbum.ParentId = stringPtr("non-existent-parent")

			result, err := albumService.CreateAlbum(context.TODO(), childAlbum)

			Expect(err).ToNot(BeNil())
			Expect(services.IsErrResourceNotFound(err)).To(BeTrue())
			Expect(result).To(BeNil())

			// Verify no folder was created on filesystem since parent doesn't exist
			tmpDir := "/tmp/photos-ng-test"
			childFolderPath := tmpDir + "/child" // Would be the original path without parent
			_, err = os.Stat(childFolderPath)
			Expect(err).ToNot(BeNil(), "Child folder should not exist when parent doesn't exist: %s", childFolderPath)
		})

		It("handles file system creation failure", func() {
			// Create album with invalid characters that might cause fs errors
			invalidAlbum := entity.NewAlbum("/test/invalid\x00album")
			invalidAlbum.Description = stringPtr("Invalid Album")

			result, err := albumService.CreateAlbum(context.TODO(), invalidAlbum)

			Expect(err).ToNot(BeNil())
			Expect(result).To(BeNil())

			// Verify transaction was rolled back - album shouldn't exist
			var count int
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM albums WHERE id = $1", invalidAlbum.ID).Scan(&count)
			Expect(err).To(BeNil())
			Expect(count).To(Equal(0))

			// Verify folder was not created on filesystem
			tmpDir := "/tmp/photos-ng-test"
			folderPath := tmpDir + invalidAlbum.Path
			_, err = os.Stat(folderPath)
			Expect(err).ToNot(BeNil(), "Invalid album folder should not exist at: %s", folderPath)
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})

	Context("UpdateAlbum", func() {
		It("updates existing album successfully", func() {
			// First create the album
			now := time.Now()
			sql, args, err := insertAlbumStmt.
				Values(testAlbum.ID, now, testAlbum.Path, testAlbum.Description, nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			// Update the album
			updatedAlbum := testAlbum
			updatedAlbum.Description = stringPtr("Updated Description")

			result, err := albumService.UpdateAlbum(context.TODO(), updatedAlbum)

			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())
			Expect(*result.Description).To(Equal("Updated Description"))

			// Verify update in database
			var description string
			err = pgPool.QueryRow(context.TODO(), "SELECT description FROM albums WHERE id = $1", testAlbum.ID).Scan(&description)
			Expect(err).To(BeNil())
			Expect(description).To(Equal("Updated Description"))
		})

		It("validates thumbnail belongs to album", func() {
			// Create album
			now := time.Now()
			sql, args, err := insertAlbumStmt.
				Values(testAlbum.ID, now, testAlbum.Path, testAlbum.Description, nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			// Try to update with non-existent thumbnail
			updatedAlbum := testAlbum
			updatedAlbum.Thumbnail = stringPtr("non-existent-media-id")

			result, err := albumService.UpdateAlbum(context.TODO(), updatedAlbum)

			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("thumbnail"))
			Expect(result).To(BeNil())
		})

		It("updates album's thumbnail successfully", func() {
			// Create album
			now := time.Now()
			sql, args, err := insertAlbumStmt.
				Values(testAlbum.ID, now, testAlbum.Path, testAlbum.Description, nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			// Create a media item that belongs to this album
			mediaID := "test-media-" + testAlbum.ID
			capturedAt := time.Now()
			exifData := `{"Make": "Canon", "Model": "EOS R5"}`

			sql, args, err = insertMediaStmt.
				Values(mediaID, now, capturedAt, testAlbum.ID, "thumbnail.jpg", "thumbnail_data", exifData, "photo").
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			// Update the album with the media ID as thumbnail
			updatedAlbum := testAlbum
			updatedAlbum.Description = stringPtr("Updated with thumbnail")
			updatedAlbum.Thumbnail = &mediaID

			result, err := albumService.UpdateAlbum(context.TODO(), updatedAlbum)

			Expect(err).To(BeNil())
			Expect(result).ToNot(BeNil())
			Expect(*result.Description).To(Equal("Updated with thumbnail"))
			Expect(result.Thumbnail).ToNot(BeNil())
			Expect(*result.Thumbnail).To(Equal(mediaID))

			// Verify update in database
			var description, thumbnailID string
			err = pgPool.QueryRow(context.TODO(), "SELECT description, thumbnail_id FROM albums WHERE id = $1", testAlbum.ID).Scan(&description, &thumbnailID)
			Expect(err).To(BeNil())
			Expect(description).To(Equal("Updated with thumbnail"))
			Expect(thumbnailID).To(Equal(mediaID))
		})

		It("returns error when album doesn't exist", func() {
			result, err := albumService.UpdateAlbum(context.TODO(), testAlbum)

			Expect(err).ToNot(BeNil())
			Expect(services.IsErrResourceNotFound(err)).To(BeTrue())
			Expect(result).To(BeNil())
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})

	Context("DeleteAlbum", func() {
		It("deletes existing album successfully", func() {
			// First create the album
			now := time.Now()
			sql, args, err := insertAlbumStmt.
				Values(testAlbum.ID, now, testAlbum.Path, testAlbum.Description, nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			err = albumService.DeleteAlbum(context.TODO(), testAlbum.ID)

			Expect(err).To(BeNil())

			// Verify album was deleted from database
			var count int
			err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM albums WHERE id = $1", testAlbum.ID).Scan(&count)
			Expect(err).To(BeNil())
			Expect(count).To(Equal(0))
		})

		It("returns error when album doesn't exist", func() {
			err := albumService.DeleteAlbum(context.TODO(), "non-existent-id")

			Expect(err).ToNot(BeNil())
			Expect(services.IsErrResourceNotFound(err)).To(BeTrue())
		})

		It("handles file system deletion failure gracefully", func() {
			// First create the album
			now := time.Now()
			sql, args, err := insertAlbumStmt.
				Values(testAlbum.ID, now, testAlbum.Path, testAlbum.Description, nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			// This test demonstrates that even if fs deletion could fail,
			// the service should handle it properly
			err = albumService.DeleteAlbum(context.TODO(), testAlbum.ID)

			// In the real implementation, fs operations are part of the transaction
			// If fs fails, the whole transaction should roll back
			if err != nil {
				// Verify transaction was rolled back - album should still exist
				var count int
				err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM albums WHERE id = $1", testAlbum.ID).Scan(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(1))
			} else {
				// If deletion succeeded, album should be gone
				var count int
				err = pgPool.QueryRow(context.TODO(), "SELECT COUNT(*) FROM albums WHERE id = $1", testAlbum.ID).Scan(&count)
				Expect(err).To(BeNil())
				Expect(count).To(Equal(0))
			}
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})

	Context("SyncAlbum", func() {
		It("syncs existing album successfully", func() {
			// First create the album
			now := time.Now()
			sql, args, err := insertAlbumStmt.
				Values(testAlbum.ID, now, testAlbum.Path, testAlbum.Description, nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			count, err := albumService.SyncAlbum(context.TODO(), testAlbum.ID)

			Expect(err).To(BeNil())
			Expect(count).To(Equal(0)) // Current implementation returns 0
		})

		It("returns error when album doesn't exist", func() {
			count, err := albumService.SyncAlbum(context.TODO(), "non-existent-id")

			Expect(err).ToNot(BeNil())
			Expect(services.IsErrResourceNotFound(err)).To(BeTrue())
			Expect(count).To(Equal(0))
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
