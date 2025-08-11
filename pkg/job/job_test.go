package job_test

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/pg"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/services"
	"git.tls.tupangiu.ro/cosmin/photos-ng/pkg/job"
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
)

// getSampleJPEGPath returns the absolute path to the sample JPEG file
func getSampleJPEGPath() string {
	// Get the directory where this test file is located
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	return filepath.Join(testDir, "testdata/sample.jpg")
}

func TestJob(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SyncAlbumJob Suite")
}

var _ = Describe("SyncAlbumJob", Ordered, func() {
	var (
		albumService *services.AlbumService
		mediaService *services.MediaService
		dt           *pg.Datastore
		pgPool       *pgxpool.Pool
		tmpDir       string
		rootAlbum    entity.Album
		realFs       *fs.Datastore
	)

	BeforeAll(func() {
		// Set up PostgreSQL connection
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
		if tmpDir != "" {
			os.RemoveAll(tmpDir)
		}
	})

	BeforeEach(func() {
		// Create a temporary directory structure for testing
		var err error
		tmpDir, err = os.MkdirTemp("", "photos-ng-sync-test-*")
		Expect(err).To(BeNil())

		// Set up filesystem datastore
		realFs = fs.NewFsDatastore(tmpDir)

		// Set up services
		albumService = services.NewAlbumService(dt, realFs)
		mediaService = services.NewMediaService(dt, realFs)

		// Create root album for testing - fs datastore works with relative paths
		rootAlbum = entity.NewAlbum("photos")
		rootAlbum.Description = stringPtr("Root Photos Album")
	})

	AfterEach(func() {
		// Clean up database
		_, err := pgPool.Exec(context.TODO(), "DELETE FROM media;")
		Expect(err).To(BeNil())
		_, err = pgPool.Exec(context.TODO(), "DELETE FROM albums;")
		Expect(err).To(BeNil())

		// Clean up filesystem
		if tmpDir != "" {
			os.RemoveAll(tmpDir)
		}
	})

	Context("Job Creation", func() {
		It("creates a new sync job successfully", func() {
			syncJob, err := job.NewSyncJob(rootAlbum, albumService, mediaService, realFs)

			Expect(err).To(BeNil())
			Expect(syncJob).ToNot(BeNil())

			status := syncJob.Status()
			Expect(status.Status).To(Equal(entity.StatusPending))
			Expect(status.Total).To(Equal(0))
			Expect(status.Remaining).To(Equal(0))
			Expect(status.Completed).To(HaveLen(0))
		})

		It("initializes job with correct timestamps", func() {
			beforeCreation := time.Now()
			syncJob, err := job.NewSyncJob(rootAlbum, albumService, mediaService, realFs)
			afterCreation := time.Now()

			Expect(err).To(BeNil())
			status := syncJob.Status()
			Expect(status.CreatedAt).To(BeTemporally(">=", beforeCreation))
			Expect(status.CreatedAt).To(BeTemporally("<=", afterCreation))
			Expect(status.StartedAt).To(BeNil())
			Expect(status.CompletedAt).To(BeNil())
		})
	})

	Context("Album Discovery and Creation", func() {
		It("discovers and creates albums from directory structure", func() {
			// Create directory structure - only one level of sub-albums
			createTestDirectoryStructure(tmpDir, map[string][]string{
				"photos/2023": {},
				"photos/2024": {},
				"photos/2025": {},
			})

			// Create root album in database
			_, err := albumService.CreateAlbum(context.TODO(), rootAlbum)
			Expect(err).To(BeNil())

			syncJob, err := job.NewSyncJob(rootAlbum, albumService, mediaService, realFs)
			Expect(err).To(BeNil())

			// Start the sync job
			err = syncJob.Start(context.TODO())
			Expect(err).To(BeNil())

			// Verify job completed successfully
			status := syncJob.Status()
			Expect(status.Status).To(Equal(entity.StatusCompleted))
			Expect(status.StartedAt).ToNot(BeNil())
			Expect(status.CompletedAt).ToNot(BeNil())

			// With LEFT JOIN, we get albums with children nested
			// The query returns root albums with their direct children in the Children field
			albums, err := albumService.GetAlbums(context.TODO(), &services.AlbumOptions{Limit: 100})
			Expect(err).To(BeNil())

			// We should get back the root album with its children nested
			Expect(albums).To(HaveLen(1))

			rootAlbumResult := albums[0]
			Expect(rootAlbumResult.Path).To(Equal("photos"))
			Expect(rootAlbumResult.Children).To(HaveLen(3)) // 2023, 2024, 2025

			// Verify specific child albums exist
			childPaths := make(map[string]bool)
			for _, child := range rootAlbumResult.Children {
				childPaths[child.Path] = true
			}
			Expect(childPaths).To(HaveKey("photos/2023"))
			Expect(childPaths).To(HaveKey("photos/2024"))
			Expect(childPaths).To(HaveKey("photos/2025"))
		})

		It("handles empty directory structure", func() {
			// Create only root directory
			err := os.MkdirAll(filepath.Join(tmpDir, "photos"), 0755)
			Expect(err).To(BeNil())

			// Create root album in database
			_, err = albumService.CreateAlbum(context.TODO(), rootAlbum)
			Expect(err).To(BeNil())

			syncJob, err := job.NewSyncJob(rootAlbum, albumService, mediaService, realFs)
			Expect(err).To(BeNil())

			// Start the sync job
			err = syncJob.Start(context.TODO())
			Expect(err).To(BeNil())

			// Verify job completed successfully
			status := syncJob.Status()
			Expect(status.Status).To(Equal(entity.StatusCompleted))
			Expect(status.Total).To(Equal(0)) // No subdirectories or files to process
		})

		It("handles root album with empty path (starts from root data folder)", func() {
			// Create some media files in various directories (directories will be created automatically)
			// NOTE: NO media files in root data folder - they only exist in subdirectories
			createTestMediaFiles(tmpDir, map[string]string{
				"2023/summer/beach.jpg":    "SAMPLE_JPEG",
				"2023/summer/vacation.jpg": "SAMPLE_JPEG",
				"2023/winter/skiing.jpg":   "SAMPLE_JPEG",
				"2024/spring/flowers.jpg":  "SAMPLE_JPEG",
				"documents/readme.txt":     "SAMPLE_JPEG", // Non-media file, should be ignored
			})

			// Create root album with empty path (does NOT go in database - represents root data folder)
			emptyRootAlbum := entity.NewAlbum("")
			emptyRootAlbum.Description = stringPtr("Root Data Folder")

			// DON'T create the root album in database - empty path means "start from root data folder"
			// The job will discover top-level directories as albums

			syncJob, err := job.NewSyncJob(emptyRootAlbum, albumService, mediaService, realFs)
			Expect(err).To(BeNil())

			// Start the sync job
			err = syncJob.Start(context.TODO())
			Expect(err).To(BeNil())

			// Verify job completed successfully
			status := syncJob.Status()

			Expect(status.Status).To(Equal(entity.StatusCompleted))
			// Should discover: 2023, 2023/summer, 2023/winter, 2024, 2024/spring, documents = 6 albums
			// Should process: 2 + 1 + 1 = 4 media files (readme.txt ignored as non-media, no root media)
			Expect(status.Total).To(Equal(10)) // 6 albums + 4 media files

			// Verify albums were created - should find top-level directories as albums (no parent)
			albums, err := albumService.GetAlbums(context.TODO(), &services.AlbumOptions{})
			Expect(err).To(BeNil())

			// Top-level albums should be: 2023, 2024, documents (no empty path album in database)
			topLevelAlbums := []entity.Album{}
			for _, album := range albums {
				if album.ParentId == nil { // Top-level albums have no parent
					topLevelAlbums = append(topLevelAlbums, album)
				}
			}
			Expect(topLevelAlbums).To(HaveLen(3))

			// Check paths of top-level albums
			topLevelPaths := make(map[string]bool)
			for _, album := range topLevelAlbums {
				topLevelPaths[album.Path] = true
			}
			Expect(topLevelPaths).To(HaveKey("2023"))
			Expect(topLevelPaths).To(HaveKey("2024"))
			Expect(topLevelPaths).To(HaveKey("documents"))

			// Verify that there's NO album with empty path in the database
			for _, album := range albums {
				Expect(album.Path).ToNot(Equal(""))
			}

			// Verify media files were processed (no root level media files)
			for _, result := range status.Completed {
				if result.ItemType == "media" {
					GinkgoWriter.Printf("Processed media: %s\n", result.Name)
				}
			}
		})
	})

	Context("Media Processing", func() {
		It("processes media files in album directories", func() {
			// Create directory structure with media files - only one level
			createTestDirectoryStructure(tmpDir, map[string][]string{
				"photos/2023": {"photo1.jpg", "photo2.jpeg", "document.txt", "photo3.png"},
				"photos/2024": {"vacation1.jpg", "vacation2.JPG"},
			})

			// Create actual media files
			createTestMediaFiles(tmpDir, map[string]string{
				"photos/2023/photo1.jpg":    "SAMPLE_JPEG",
				"photos/2023/photo2.jpeg":   "SAMPLE_JPEG",
				"photos/2023/photo3.png":    "SAMPLE_JPEG", // Will be treated as JPEG for simplicity
				"photos/2024/vacation1.jpg": "SAMPLE_JPEG",
				"photos/2024/vacation2.JPG": "SAMPLE_JPEG",
			})

			// Create root album in database
			_, err := albumService.CreateAlbum(context.TODO(), rootAlbum)
			Expect(err).To(BeNil())

			syncJob, err := job.NewSyncJob(rootAlbum, albumService, mediaService, realFs)
			Expect(err).To(BeNil())

			// Start the sync job
			err = syncJob.Start(context.TODO())
			Expect(err).To(BeNil())

			// Verify job completed successfully
			status := syncJob.Status()
			Expect(status.Status).To(Equal(entity.StatusCompleted))
			Expect(status.StartedAt).ToNot(BeNil())
			Expect(status.CompletedAt).ToNot(BeNil())

			// Debug: Print job status for troubleshooting
			GinkgoWriter.Printf("Job status: %+v\n", status)
			for i, result := range status.Completed {
				GinkgoWriter.Printf("Task %d: Type=%s, Name=%s, Error=%v\n", i, result.ItemType, result.Name, result.Err)
			}

			// Verify album structure
			albums, err := albumService.GetAlbums(context.TODO(), &services.AlbumOptions{Limit: 100})
			Expect(err).To(BeNil())

			// With LEFT JOIN, we get the root album with children and media nested
			Expect(albums).To(HaveLen(1)) // Just the root album with children nested
			rootAlbumResult := albums[0]
			Expect(rootAlbumResult.Children).To(HaveLen(2)) // 2023, 2024

			// Verify media files were processed (5 media files total)
			allMedia, err := mediaService.GetMedia(context.TODO(), &services.MediaOptions{MediaLimit: 100})
			Expect(err).To(BeNil())
			GinkgoWriter.Printf("Found %d media files\n", len(allMedia))
			Expect(allMedia).To(HaveLen(5))

			// Verify specific media files exist
			mediaFilenames := make(map[string]bool)
			for _, media := range allMedia {
				mediaFilenames[media.Filename] = true
			}
			Expect(mediaFilenames).To(HaveKey("photo1.jpg"))
			Expect(mediaFilenames).To(HaveKey("photo2.jpeg"))
			Expect(mediaFilenames).To(HaveKey("photo3.png"))
			Expect(mediaFilenames).To(HaveKey("vacation1.jpg"))
			Expect(mediaFilenames).To(HaveKey("vacation2.JPG"))
			// document.txt should not be processed
			Expect(mediaFilenames).ToNot(HaveKey("document.txt"))

			// Verify task results in job status (2 albums + 5 media files)
			Expect(status.Completed).To(HaveLen(7))

			// Count successful tasks
			successfulTasks := 0
			for _, result := range status.Completed {
				if result.Err == nil {
					successfulTasks++
				}
			}
			Expect(successfulTasks).To(Equal(7))
		})

	})

	Context("Multi-Level Nested Albums", func() {
		It("processes deeply nested album structures (3+ levels)", func() {
			// Create a complex nested directory structure - 3 levels deep
			createTestDirectoryStructure(tmpDir, map[string][]string{
				"photos/2023/summer":        {"vacation1.jpg", "beach.jpg"},
				"photos/2023/summer/europe": {"paris.jpg", "rome.jpeg"},
				"photos/2023/winter":        {"skiing.jpg"},
				"photos/2023/winter/alps":   {"mountain1.jpg", "mountain2.JPG"},
				"photos/2024/work":          {"conference.jpg"},
				"photos/2024/work/projects": {"demo.jpg", "presentation.png"},
			})

			// Create actual media files for all locations
			createTestMediaFiles(tmpDir, map[string]string{
				"photos/2023/summer/vacation1.jpg":           "SAMPLE_JPEG",
				"photos/2023/summer/beach.jpg":               "SAMPLE_JPEG",
				"photos/2023/summer/europe/paris.jpg":        "SAMPLE_JPEG",
				"photos/2023/summer/europe/rome.jpeg":        "SAMPLE_JPEG",
				"photos/2023/winter/skiing.jpg":              "SAMPLE_JPEG",
				"photos/2023/winter/alps/mountain1.jpg":      "SAMPLE_JPEG",
				"photos/2023/winter/alps/mountain2.JPG":      "SAMPLE_JPEG",
				"photos/2024/work/conference.jpg":            "SAMPLE_JPEG",
				"photos/2024/work/projects/demo.jpg":         "SAMPLE_JPEG",
				"photos/2024/work/projects/presentation.png": "SAMPLE_JPEG",
			})

			// Create root album in database
			_, err := albumService.CreateAlbum(context.TODO(), rootAlbum)
			Expect(err).To(BeNil())

			syncJob, err := job.NewSyncJob(rootAlbum, albumService, mediaService, realFs)
			Expect(err).To(BeNil())

			// Start the sync job
			err = syncJob.Start(context.TODO())
			Expect(err).To(BeNil())

			// Verify job completed successfully
			status := syncJob.Status()
			Expect(status.Status).To(Equal(entity.StatusCompleted))
			Expect(status.StartedAt).ToNot(BeNil())
			Expect(status.CompletedAt).ToNot(BeNil())

			// Debug: Print job status for troubleshooting
			GinkgoWriter.Printf("Multi-level job status: %+v\n", status)
			for i, result := range status.Completed {
				GinkgoWriter.Printf("Task %d: Type=%s, Name=%s, Error=%v\n", i, result.ItemType, result.Name, result.Err)
			}

			// With multi-level nesting, we should have discovered ALL albums recursively
			// Expected structure: photos/{2023,2024,2023/summer,2023/winter,2024/work,2023/summer/europe,2023/winter/alps,2024/work/projects}
			// That's 8 albums + 10 media files = 18 total tasks
			GinkgoWriter.Printf("Expected: 8 albums + 10 media files = 18 tasks, Got: %d tasks\n", status.Total)
			Expect(status.Total).To(BeNumerically(">=", 18))

			// Verify all albums were created - recursive discovery should find all levels
			albums, err := albumService.GetAlbums(context.TODO(), &services.AlbumOptions{Limit: 100})
			Expect(err).To(BeNil())
			Expect(albums).To(HaveLen(1)) // Root album with nested children

			rootAlbumResult := albums[0]
			GinkgoWriter.Printf("Root album has %d direct children\n", len(rootAlbumResult.Children))

			// The root should have immediate children (2023, 2024)
			// Each of those should have their own nested children
			Expect(rootAlbumResult.Children).To(HaveLen(2)) // 2023, 2024

			// Verify all media files were processed (10 media files total)
			allMedia, err := mediaService.GetMedia(context.TODO(), &services.MediaOptions{MediaLimit: 100})
			Expect(err).To(BeNil())
			GinkgoWriter.Printf("Found %d media files in database\n", len(allMedia))
			Expect(allMedia).To(HaveLen(10))

			// Verify some specific files exist at different nesting levels
			mediaFilenames := make(map[string]bool)
			for _, media := range allMedia {
				mediaFilenames[media.Filename] = true
			}

			// Level 2 files (photos/2023/summer/)
			Expect(mediaFilenames).To(HaveKey("vacation1.jpg"))
			Expect(mediaFilenames).To(HaveKey("beach.jpg"))

			// Level 3 files (photos/2023/summer/europe/)
			Expect(mediaFilenames).To(HaveKey("paris.jpg"))
			Expect(mediaFilenames).To(HaveKey("rome.jpeg"))

			// Level 3 files (photos/2023/winter/alps/)
			Expect(mediaFilenames).To(HaveKey("mountain1.jpg"))
			Expect(mediaFilenames).To(HaveKey("mountain2.JPG"))

			// Level 3 files (photos/2024/work/projects/)
			Expect(mediaFilenames).To(HaveKey("demo.jpg"))
			Expect(mediaFilenames).To(HaveKey("presentation.png"))

			// Verify all tasks completed successfully
			successfulTasks := 0
			for _, result := range status.Completed {
				if result.Err == nil {
					successfulTasks++
				}
			}
			Expect(successfulTasks).To(Equal(len(status.Completed)))
		})

		It("handles mixed content with some invalid media files", func() {
			// Create directory structure
			createTestDirectoryStructure(tmpDir, map[string][]string{
				"photos/mixed": {"valid.jpg", "invalid.jpg", "another_valid.jpeg"},
			})

			// Create media files - one invalid
			createTestMediaFiles(tmpDir, map[string]string{
				"photos/mixed/valid.jpg":          "SAMPLE_JPEG",
				"photos/mixed/invalid.jpg":        "invalid-image-data",
				"photos/mixed/another_valid.jpeg": "SAMPLE_JPEG",
			})

			// Create root album in database
			_, err := albumService.CreateAlbum(context.TODO(), rootAlbum)
			Expect(err).To(BeNil())

			syncJob, err := job.NewSyncJob(rootAlbum, albumService, mediaService, realFs)
			Expect(err).To(BeNil())

			// Start the sync job
			err = syncJob.Start(context.TODO())
			Expect(err).To(BeNil())

			// Job should complete even with some failed media processing
			status := syncJob.Status()
			Expect(status.Status).To(Equal(entity.StatusCompleted))

			// Check task results
			Expect(status.Completed).To(HaveLen(4)) // 1 album + 3 media files

			// Count successful vs failed tasks
			successfulTasks := 0
			failedTasks := 0
			for _, result := range status.Completed {
				if result.Err == nil {
					successfulTasks++
				} else {
					failedTasks++
				}
			}

			// Album creation should succeed, and valid media should succeed
			Expect(successfulTasks).To(BeNumerically(">=", 2)) // At least album + 1 valid media
			Expect(failedTasks).To(BeNumerically(">=", 1))     // At least 1 failed media
		})
	})

	Context("Job Progress Tracking", func() {
		It("tracks progress correctly during execution", func() {
			// Create a simple structure - one level
			createTestDirectoryStructure(tmpDir, map[string][]string{
				"photos/album1": {"photo1.jpg"},
				"photos/album2": {"photo2.jpg"},
			})

			createTestMediaFiles(tmpDir, map[string]string{
				"photos/album1/photo1.jpg": "SAMPLE_JPEG",
				"photos/album2/photo2.jpg": "SAMPLE_JPEG",
			})

			// Create root album in database
			_, err := albumService.CreateAlbum(context.TODO(), rootAlbum)
			Expect(err).To(BeNil())

			syncJob, err := job.NewSyncJob(rootAlbum, albumService, mediaService, realFs)
			Expect(err).To(BeNil())

			// Check initial status
			status := syncJob.Status()
			Expect(status.Status).To(Equal(entity.StatusPending))
			Expect(status.Total).To(Equal(0))
			Expect(status.Remaining).To(Equal(0))

			// Start the job
			err = syncJob.Start(context.TODO())
			Expect(err).To(BeNil())

			// Check final status - 2 child albums + 2 media files
			finalStatus := syncJob.Status()
			Expect(finalStatus.Status).To(Equal(entity.StatusCompleted))
			Expect(finalStatus.Total).To(Equal(4))
			Expect(finalStatus.Remaining).To(Equal(0))
			Expect(finalStatus.Completed).To(HaveLen(4))
			Expect(finalStatus.StartedAt).ToNot(BeNil())
			Expect(finalStatus.CompletedAt).ToNot(BeNil())
			Expect(finalStatus.CompletedAt.After(*finalStatus.StartedAt)).To(BeTrue())
		})
	})

	Context("Error Handling", func() {
		It("handles non-existent root album", func() {
			nonExistentAlbum := entity.NewAlbum("nonexistent")

			syncJob, err := job.NewSyncJob(nonExistentAlbum, albumService, mediaService, realFs)
			Expect(err).To(BeNil())

			// Start should fail because root album doesn't exist
			err = syncJob.Start(context.TODO())
			Expect(err).ToNot(BeNil())

			status := syncJob.Status()
			Expect(status.Status).To(Equal(entity.StatusFailed))
		})

		It("continues processing despite individual media failures", func() {
			// Create structure with problematic file that can't be processed - one level
			createTestDirectoryStructure(tmpDir, map[string][]string{
				"photos/test": {"good.jpg", "bad.jpg", "another_good.jpg"},
			})

			// Create files - one with bad content
			createTestMediaFiles(tmpDir, map[string]string{
				"photos/test/good.jpg":         "SAMPLE_JPEG",
				"photos/test/bad.jpg":          "completely-invalid-binary-data",
				"photos/test/another_good.jpg": "SAMPLE_JPEG",
			})

			// Create root album in database
			_, err := albumService.CreateAlbum(context.TODO(), rootAlbum)
			Expect(err).To(BeNil())

			syncJob, err := job.NewSyncJob(rootAlbum, albumService, mediaService, realFs)
			Expect(err).To(BeNil())

			// Start the job
			err = syncJob.Start(context.TODO())
			Expect(err).To(BeNil())

			// Job should complete despite failures
			status := syncJob.Status()
			Expect(status.Status).To(Equal(entity.StatusCompleted))

			// Some tasks should have failed, but job continues - 1 child album + 3 media files
			Expect(status.Completed).To(HaveLen(4))

			errorCount := 0
			for _, result := range status.Completed {
				if result.Err != nil {
					errorCount++
				}
			}
			Expect(errorCount).To(BeNumerically(">=", 1)) // At least one media processing failure
		})
	})

	Context("Complex Directory Structures", func() {
		It("handles deeply nested directory structures", func() {
			// Create deep nested structure
			createTestDirectoryStructure(tmpDir, map[string][]string{
				"photos/2023/summer/vacation/beach":     {"sunset.jpg"},
				"photos/2023/summer/vacation/mountain":  {"hike.jpg"},
				"photos/2023/winter/holidays/christmas": {"tree.jpg"},
				"photos/2024/work/conference":           {"presentation.jpg"},
			})

			createTestMediaFiles(tmpDir, map[string]string{
				"photos/2023/summer/vacation/beach/sunset.jpg":   "SAMPLE_JPEG",
				"photos/2023/summer/vacation/mountain/hike.jpg":  "SAMPLE_JPEG",
				"photos/2023/winter/holidays/christmas/tree.jpg": "SAMPLE_JPEG",
				"photos/2024/work/conference/presentation.jpg":   "SAMPLE_JPEG",
			})

			// Create root album in database
			_, err := albumService.CreateAlbum(context.TODO(), rootAlbum)
			Expect(err).To(BeNil())

			syncJob, err := job.NewSyncJob(rootAlbum, albumService, mediaService, realFs)
			Expect(err).To(BeNil())

			// Start the sync job
			err = syncJob.Start(context.TODO())
			Expect(err).To(BeNil())

			// Verify job completed successfully
			status := syncJob.Status()
			Expect(status.Status).To(Equal(entity.StatusCompleted))

			// The sync job processes all nested directories and creates them as albums
			// We verify this by checking that all media files were processed
			allMedia, err := mediaService.GetMedia(context.TODO(), &services.MediaOptions{MediaLimit: 100})
			Expect(err).To(BeNil())
			Expect(allMedia).To(HaveLen(4))

			// Verify that the job processed all directories and media files
			// The exact number will be: all nested directories + all media files
			Expect(status.Completed).To(HaveLen(15)) // 11 directories + 4 media files

			// Count successful tasks - should be high since these are valid structures
			successfulTasks := 0
			for _, result := range status.Completed {
				if result.Err == nil {
					successfulTasks++
				}
			}
			Expect(successfulTasks).To(BeNumerically(">=", 10)) // Most tasks should succeed
		})
	})
})

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func createTestDirectoryStructure(basePath string, structure map[string][]string) {
	for dirPath, files := range structure {
		// Clean the directory path to handle both absolute and relative paths
		cleanDirPath := strings.TrimPrefix(dirPath, "/")
		fullDirPath := filepath.Join(basePath, cleanDirPath)
		err := os.MkdirAll(fullDirPath, 0755)
		Expect(err).To(BeNil())

		// Create empty files for now
		for _, fileName := range files {
			fullFilePath := filepath.Join(fullDirPath, fileName)
			file, err := os.Create(fullFilePath)
			Expect(err).To(BeNil())
			file.Close()
		}
	}
}

func createTestMediaFiles(basePath string, files map[string]string) {
	for filePath, content := range files {
		// Clean the file path to handle both absolute and relative paths
		cleanFilePath := strings.TrimPrefix(filePath, "/")
		fullPath := filepath.Join(basePath, cleanFilePath)

		// Ensure directory exists
		dir := filepath.Dir(fullPath)
		err := os.MkdirAll(dir, 0755)
		Expect(err).To(BeNil())

		var data []byte
		if content == "SAMPLE_JPEG" {
			// Copy the real JPEG file
			data, err = os.ReadFile(getSampleJPEGPath())
			Expect(err).To(BeNil())
		} else if decoded, err := base64.StdEncoding.DecodeString(content); err == nil {
			// Decode base64 content if it looks like base64
			data = decoded
		} else {
			// Use content as raw bytes
			data = []byte(content)
		}

		err = os.WriteFile(fullPath, data, 0644)
		Expect(err).To(BeNil())
	}
}
