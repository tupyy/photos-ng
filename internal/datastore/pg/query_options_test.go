package pg_test

import (
	"context"
	"encoding/json"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"github.com/jackc/pgx/v5/pgxpool"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("QueryOptions", Ordered, func() {
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

	Context("FilterById", func() {
		BeforeEach(func() {
			// Set up test data
			now := time.Now()
			sql, args, err := insertAlbumStmt.
				Values("album1", now, "/path1", "Album 1", nil, nil).
				Values("album2", now, "/path2", "Album 2", nil, nil).
				Values("album3", now, "/path3", "Album 3", nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())
		})

		It("filters albums by specific ID", func() {
			albums, err := dt.QueryAlbums(context.TODO(), pg.FilterByAlbumId("album2"))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(1))
			Expect(albums[0].ID).To(Equal("album2"))
			Expect(albums[0].Path).To(Equal("/path2"))
		})

		It("returns empty result for non-existent ID", func() {
			albums, err := dt.QueryAlbums(context.TODO(), pg.FilterByAlbumId("non_existent"))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(0))
		})

		It("returns all albums when ID is empty", func() {
			albums, err := dt.QueryAlbums(context.TODO(), pg.FilterByAlbumId(""))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(3))
		})

		It("filters media by specific ID", func() {
			// Insert test media
			exifData := map[string]string{"Make": "Canon"}
			exifJSON, err := json.Marshal(exifData)
			Expect(err).To(BeNil())

			now := time.Now()
			sql, args, err := insertMediaStmt.
				Values("media1", now, now, "album1", "photo1.jpg", []byte("thumb"), exifJSON, "photo").
				Values("media2", now, now, "album1", "photo2.jpg", []byte("thumb"), exifJSON, "photo").
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())

			media, err := dt.QueryMedia(context.TODO(), pg.FilterByMediaId("media2"))
			Expect(err).To(BeNil())
			Expect(media).To(HaveLen(1))
			Expect(media[0].ID).To(Equal("media2"))
			Expect(media[0].Filename).To(Equal("photo2.jpg"))
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})

	Context("FilterAlbumByParentId", func() {
		BeforeEach(func() {
			// Set up test data with parent-child relationships
			now := time.Now()
			sql, args, err := insertAlbumStmt.
				Values("root1", now, "/root1", "Root Album 1", nil, nil).
				Values("root2", now, "/root2", "Root Album 2", nil, nil).
				Values("child1", now, "/root1/child1", "Child 1", "root1", nil).
				Values("child2", now, "/root1/child2", "Child 2", "root1", nil).
				Values("child3", now, "/root2/child3", "Child 3", "root2", nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())
		})

		It("filters albums by parent ID", func() {
			albums, err := dt.QueryAlbums(context.TODO(), pg.FilterAlbumByParentId("root1"))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(2))

			childIDs := []string{albums[0].ID, albums[1].ID}
			Expect(childIDs).To(ContainElements("child1", "child2"))
		})

		It("filters albums by different parent ID", func() {
			albums, err := dt.QueryAlbums(context.TODO(), pg.FilterAlbumByParentId("root2"))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(1))
			Expect(albums[0].ID).To(Equal("child3"))
		})

		It("returns root albums when parent ID is empty", func() {
			albums, err := dt.QueryAlbums(context.TODO(), pg.FilterAlbumByParentId(""))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(2))

			albumIDs := []string{albums[0].ID, albums[1].ID}
			Expect(albumIDs).To(ContainElements("root1", "root2"))
		})

		It("returns empty result for non-existent parent ID", func() {
			albums, err := dt.QueryAlbums(context.TODO(), pg.FilterAlbumByParentId("non_existent"))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(0))
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})

	Context("Limit", func() {
		BeforeEach(func() {
			// Set up test data
			now := time.Now()
			sql, args, err := insertAlbumStmt.
				Values("album1", now, "/path1", "Album 1", nil, nil).
				Values("album2", now, "/path2", "Album 2", nil, nil).
				Values("album3", now, "/path3", "Album 3", nil, nil).
				Values("album4", now, "/path4", "Album 4", nil, nil).
				Values("album5", now, "/path5", "Album 5", nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())
		})

		It("limits albums to specified count", func() {
			albums, err := dt.QueryAlbums(context.TODO(), pg.Limit(3))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(3))
		})

		It("returns all albums when limit is larger than count", func() {
			albums, err := dt.QueryAlbums(context.TODO(), pg.Limit(10))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(5))
		})

		It("returns all albums when limit is zero", func() {
			albums, err := dt.QueryAlbums(context.TODO(), pg.Limit(0))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(5))
		})

		It("returns all albums when limit is negative", func() {
			albums, err := dt.QueryAlbums(context.TODO(), pg.Limit(-5))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(5))
		})

		It("limits media to specified count", func() {
			// Insert test media
			exifData := map[string]string{"Make": "Canon"}
			exifJSON, err := json.Marshal(exifData)
			Expect(err).To(BeNil())

			now := time.Now()
			sql, args, err := insertMediaStmt.
				Values("media1", now, now, "album1", "photo1.jpg", []byte("thumb"), exifJSON, "photo").
				Values("media2", now, now, "album1", "photo2.jpg", []byte("thumb"), exifJSON, "photo").
				Values("media3", now, now, "album1", "photo3.jpg", []byte("thumb"), exifJSON, "photo").
				Values("media4", now, now, "album1", "photo4.jpg", []byte("thumb"), exifJSON, "photo").
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

	Context("Offset", func() {
		BeforeEach(func() {
			// Set up test data with ordered albums
			now := time.Now()
			sql, args, err := insertAlbumStmt.
				Values("album1", now, "/path1", "Album 1", nil, nil).
				Values("album2", now, "/path2", "Album 2", nil, nil).
				Values("album3", now, "/path3", "Album 3", nil, nil).
				Values("album4", now, "/path4", "Album 4", nil, nil).
				Values("album5", now, "/path5", "Album 5", nil, nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())
		})

		It("skips specified number of albums", func() {
			albums, err := dt.QueryAlbums(context.TODO(), pg.Offset(2))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(3))
		})

		It("returns empty when offset is larger than count", func() {
			albums, err := dt.QueryAlbums(context.TODO(), pg.Offset(10))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(0))
		})

		It("returns all albums when offset is zero", func() {
			albums, err := dt.QueryAlbums(context.TODO(), pg.Offset(0))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(5))
		})

		It("returns all albums when offset is negative", func() {
			albums, err := dt.QueryAlbums(context.TODO(), pg.Offset(-2))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(5))
		})

		AfterEach(func() {
			_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
			Expect(err).To(BeNil())
		})
	})

	Context("Combined Filters", func() {
		BeforeEach(func() {
			// Set up complex test data
			now := time.Now()
			sql, args, err := insertAlbumStmt.
				Values("root1", now, "/root1", "Root Album 1", nil, nil).
				Values("root2", now, "/root2", "Root Album 2", nil, nil).
				Values("child1", now, "/root1/child1", "Child 1", "root1", nil).
				Values("child2", now, "/root1/child2", "Child 2", "root1", nil).
				Values("child3", now, "/root1/child3", "Child 3", "root1", nil).
				Values("child4", now, "/root2/child4", "Child 4", "root2", nil).
				ToSql()
			Expect(err).To(BeNil())
			_, err = pgPool.Exec(context.TODO(), sql, args...)
			Expect(err).To(BeNil())
		})

		It("combines parent filter with limit", func() {
			albums, err := dt.QueryAlbums(context.TODO(),
				pg.FilterAlbumByParentId("root1"),
				pg.Limit(2))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(2))

			// All returned albums should have root1 as parent
			for _, album := range albums {
				Expect(album.ParentId).ToNot(BeNil())
				Expect(*album.ParentId).To(Equal("root1"))
			}
		})

		It("combines parent filter with offset", func() {
			albums, err := dt.QueryAlbums(context.TODO(),
				pg.FilterAlbumByParentId("root1"),
				pg.Offset(1))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(2)) // 3 children of root1, offset 1 = 2 remaining

			for _, album := range albums {
				Expect(*album.ParentId).To(Equal("root1"))
			}
		})

		It("combines parent filter with limit and offset", func() {
			albums, err := dt.QueryAlbums(context.TODO(),
				pg.FilterAlbumByParentId("root1"),
				pg.Offset(1),
				pg.Limit(1))
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(1))

			Expect(*albums[0].ParentId).To(Equal("root1"))
		})

		It("combines limit and offset for pagination", func() {
			// Test pagination: get second page with 2 items per page
			albums, err := dt.QueryAlbums(context.TODO(),
				pg.Offset(2),
				pg.Limit(2))
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
})
