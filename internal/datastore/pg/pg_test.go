package pg_test

import (
	"context"
	"encoding/json"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
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

var _ = Describe("Query", Ordered, func() {
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

	Context("QueryAlbums", func() {
		It("query successfully albums -- empty response", func() {
			albums, err := dt.QueryAlbums(context.TODO())
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(0))
		})

		It("query successfully albums -- no children or media", func() {
			now := time.Now()
			sql, args, err := insertAlbumStmt.
				Values("album1", now, "/path/to/album1", "Album 1 Description", nil, nil).
				Values("album2", now, "/path/to/album2", "Album 2 Description", nil, nil).
				ToSql()
			Expect(err).To(BeNil())

			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			albums, err := dt.QueryAlbums(context.TODO())
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(2))

			// Verify album properties
			albumPaths := []string{albums[0].Path, albums[1].Path}
			Expect(albumPaths).To(ContainElements("/path/to/album1", "/path/to/album2"))

			albumDescriptions := []string{*albums[0].Description, *albums[1].Description}
			Expect(albumDescriptions).To(ContainElements("Album 1 Description", "Album 2 Description"))
		})

		It("query successfully albums -- with children", func() {
			// First clean up
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())

			now := time.Now()
			// Insert parent album
			sql, args, err := insertAlbumStmt.
				Values("parent1", now, "/path/to/parent", "Parent Album", nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			// Insert child album
			sql, args, err = insertAlbumStmt.
				Values("child1", now, "/path/to/parent/child", "Child Album", "parent1", nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			albums, err := dt.QueryAlbums(context.TODO())
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(2))

			// Find parent album
			var parentAlbum entity.Album
			for _, album := range albums {
				if album.ID == "parent1" {
					parentAlbum = album
					break
				}
			}

			Expect(parentAlbum.ID).To(Equal("parent1"))
			Expect(parentAlbum.Children).To(HaveLen(1))
			Expect(parentAlbum.Children[0].ID).To(Equal("child1"))
		})

		It("query successfully albums -- with media", func() {
			// First clean up
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())

			now := time.Now()
			// Insert album
			sql, args, err := insertAlbumStmt.
				Values("album_with_media", now, "/path/to/album", "Album with Media", nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			// Insert media
			exifData := map[string]string{
				"Make":  "Canon",
				"Model": "EOS R5",
			}
			exifJSON, err := json.Marshal(exifData)
			Expect(err).To(BeNil())

			sql, args, err = insertMediaStmt.
				Values("media1", now, now, "album_with_media", "photo1.jpg", []byte("thumbnail"), exifJSON, "photo").
				Values("media2", now, now, "album_with_media", "photo2.jpg", []byte("thumbnail"), exifJSON, "photo").
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			albums, err := dt.QueryAlbums(context.TODO())
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(1))
			Expect(albums[0].Media).To(HaveLen(2))

			mediaFilenames := []string{albums[0].Media[0].Filename, albums[0].Media[1].Filename}
			Expect(mediaFilenames).To(ContainElements("photo1.jpg", "photo2.jpg"))
		})

		It("query albums with filter by ID", func() {
			// First clean up
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())

			now := time.Now()
			sql, args, err := insertAlbumStmt.
				Values("album1", now, "/path/to/album1", "Album 1", nil, nil).
				Values("album2", now, "/path/to/album2", "Album 2", nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			albums, err := dt.QueryAlbums(context.TODO(), pg.FilterByAlbumId("album1"))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(1))
			Expect(albums[0].ID).To(Equal("album1"))
		})

		It("query albums with parent filter", func() {
			// First clean up
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())

			now := time.Now()
			// Insert parent and child albums
			sql, args, err := insertAlbumStmt.
				Values("parent1", now, "/path/to/parent", "Parent", nil, nil).
				Values("child1", now, "/path/to/child1", "Child 1", "parent1", nil).
				Values("child2", now, "/path/to/child2", "Child 2", "parent1", nil).
				Values("root", now, "/path/to/root", "Root Album", nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			// Query for children of parent1
			albums, err := dt.QueryAlbums(context.TODO(), pg.FilterAlbumByParentId("parent1"))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(2))

			// Query for root albums (no parent)
			albums, err = dt.QueryAlbums(context.TODO(), pg.FilterAlbumByParentId(""))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(2)) // parent1 and root
		})

		It("query albums with limit", func() {
			// First clean up
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())

			now := time.Now()
			sql, args, err := insertAlbumStmt.
				Values("album1", now, "/path/1", "Album 1", nil, nil).
				Values("album2", now, "/path/2", "Album 2", nil, nil).
				Values("album3", now, "/path/3", "Album 3", nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			albums, err := dt.QueryAlbums(context.TODO(), pg.Limit(2))
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

	Context("QueryMedia", func() {
		It("query successfully media -- empty response", func() {
			media, err := dt.QueryMedia(context.TODO())
			Expect(err).To(BeNil())
			Expect(media).To(HaveLen(0))
		})

		It("query successfully media -- with albums", func() {
			now := time.Now()

			// Insert album first
			sql, args, err := insertAlbumStmt.
				Values("test_album", now, "/path/to/album", "Test Album", nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			// Insert media
			exifData := map[string]string{
				"Make":       "Canon",
				"Model":      "EOS R5",
				"CreateDate": "2024:01:15 10:30:00",
			}
			exifJSON, err := json.Marshal(exifData)
			Expect(err).To(BeNil())

			sql, args, err = insertMediaStmt.
				Values("media1", now, now, "test_album", "photo1.jpg", []byte("thumbnail1"), exifJSON, "photo").
				Values("media2", now, now, "test_album", "video1.mp4", []byte("thumbnail2"), exifJSON, "Video").
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			media, err := dt.QueryMedia(context.TODO())
			Expect(err).To(BeNil())
			Expect(media).To(HaveLen(2))

			// Verify media properties
			filenames := []string{media[0].Filename, media[1].Filename}
			Expect(filenames).To(ContainElements("photo1.jpg", "video1.mp4"))

			mediaTypes := []entity.MediaType{media[0].MediaType, media[1].MediaType}
			Expect(mediaTypes).To(ContainElements(entity.Photo, entity.Video))

			// Verify EXIF data is properly deserialized
			for _, m := range media {
				Expect(m.Exif).To(HaveKeyWithValue("Make", "Canon"))
				Expect(m.Exif).To(HaveKeyWithValue("Model", "EOS R5"))
			}
		})

		It("query media with filter by ID", func() {
			now := time.Now()

			// Insert album first
			sql, args, err := insertAlbumStmt.
				Values("test_album", now, "/path/to/album", "Test Album", nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			// Insert media
			exifData := map[string]string{"Make": "Canon"}
			exifJSON, err := json.Marshal(exifData)
			Expect(err).To(BeNil())

			sql, args, err = insertMediaStmt.
				Values("media1", now, now, "test_album", "photo1.jpg", []byte("thumbnail"), exifJSON, "photo").
				Values("media2", now, now, "test_album", "photo2.jpg", []byte("thumbnail"), exifJSON, "photo").
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			media, err := dt.QueryMedia(context.TODO(), pg.FilterByMediaId("media1"))
			Expect(err).To(BeNil())
			Expect(media).To(HaveLen(1))
			Expect(media[0].ID).To(Equal("media1"))
		})

		It("query media with limit", func() {
			now := time.Now()

			// Insert album first
			sql, args, err := insertAlbumStmt.
				Values("test_album", now, "/path/to/album", "Test Album", nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			// Insert multiple media items
			exifData := map[string]string{"Make": "Canon"}
			exifJSON, err := json.Marshal(exifData)
			Expect(err).To(BeNil())

			sql, args, err = insertMediaStmt.
				Values("media1", now, now, "test_album", "photo1.jpg", []byte("thumbnail"), exifJSON, "photo").
				Values("media2", now, now, "test_album", "photo2.jpg", []byte("thumbnail"), exifJSON, "photo").
				Values("media3", now, now, "test_album", "photo3.jpg", []byte("thumbnail"), exifJSON, "photo").
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			media, err := dt.QueryMedia(context.TODO(), pg.Limit(2))
			Expect(err).To(BeNil())
			Expect(media).To(HaveLen(2))
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})

	Context("Stats", func() {
		It("returns empty stats", func() {
			stats, err := dt.Stats(context.TODO())
			Expect(err).To(BeNil())
			Expect(stats.CountMedia).To(Equal(0))
			Expect(stats.CountAlbum).To(Equal(0))
			Expect(stats.TimelineYears).To(HaveLen(0))
		})
	})
})
