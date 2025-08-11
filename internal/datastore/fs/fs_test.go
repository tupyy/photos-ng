package fs_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/datastore/fs"
	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFsDatastore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Filesystem Datastore Suite")
}

var _ = Describe("Filesystem Datastore", func() {
	var (
		datastore *fs.Datastore
		tmpDir    string
		ctx       context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Create temporary directory for tests
		var err error
		tmpDir, err = os.MkdirTemp("", "fs-datastore-test-*")
		Expect(err).To(BeNil())

		// Initialize datastore
		datastore = fs.NewFsDatastore(tmpDir)
	})

	AfterEach(func() {
		// Clean up temporary directory
		if tmpDir != "" {
			os.RemoveAll(tmpDir)
		}
	})

	// Helper function to create test directory structure
	createTestStructure := func(structure map[string][]string) {
		for dirPath, files := range structure {
			fullDirPath := filepath.Join(tmpDir, dirPath)
			err := os.MkdirAll(fullDirPath, 0755)
			Expect(err).To(BeNil())

			for _, fileName := range files {
				fullFilePath := filepath.Join(fullDirPath, fileName)
				err := os.WriteFile(fullFilePath, []byte("test content"), 0644)
				Expect(err).To(BeNil())
			}
		}
	}

	Describe("WalkTree", func() {
		Context("Single level directory with media files", func() {
			It("creates correct tree structure", func() {
				// Create test structure
				createTestStructure(map[string][]string{
					"photos": {"photo1.jpg", "photo2.png", "document.txt"},
				})

				// Walk the tree
				tree, err := datastore.WalkTree(ctx, "photos")
				Expect(err).To(BeNil())
				Expect(tree).ToNot(BeNil())

				// Verify root node
				Expect(tree.Path).To(Equal("photos"))
				Expect(tree.Children).To(HaveLen(0))
				Expect(tree.MediaFiles).To(HaveLen(2)) // Only jpg and png
				Expect(tree.MediaFiles).To(ContainElements("photos/photo1.jpg", "photos/photo2.png"))
				Expect(tree.Parent).To(BeNil())
			})
		})

		Context("Multi-level nested directory structure", func() {
			It("creates correct hierarchical tree", func() {
				// Create complex test structure
				createTestStructure(map[string][]string{
					"photos":                    {"root1.jpg", "root2.jpeg"},
					"photos/2023":               {"year1.jpg"},
					"photos/2023/summer":        {"vacation1.jpg", "vacation2.png", "notes.txt"},
					"photos/2023/summer/europe": {"paris.jpg", "rome.jpeg"},
					"photos/2023/winter":        {"skiing.jpg"},
					"photos/2023/winter/alps":   {"mountain1.jpg", "mountain2.JPG"},
					"photos/2024":               {"recent.jpg"},
					"photos/2024/work":          {"conference.jpg"},
					"photos/2024/work/projects": {"demo.jpg", "presentation.png"},
				})

				// Walk the tree
				tree, err := datastore.WalkTree(ctx, "photos")
				Expect(err).To(BeNil())
				Expect(tree).ToNot(BeNil())

				// Verify root node
				Expect(tree.Path).To(Equal("photos"))
				Expect(tree.MediaFiles).To(HaveLen(2)) // root1.jpg, root2.jpeg
				Expect(tree.Children).To(HaveLen(2))   // 2023, 2024

				// Find 2023 child
				var child2023 *entity.FolderNode
				for _, child := range tree.Children {
					if child.Path == "photos/2023" {
						child2023 = child
						break
					}
				}
				Expect(child2023).ToNot(BeNil())
				Expect(child2023.MediaFiles).To(HaveLen(1)) // year1.jpg
				Expect(child2023.Children).To(HaveLen(2))   // summer, winter
				Expect(child2023.Parent).To(Equal(tree))

				// Find summer child of 2023
				var summer *entity.FolderNode
				for _, child := range child2023.Children {
					if child.Path == "photos/2023/summer" {
						summer = child
						break
					}
				}
				Expect(summer).ToNot(BeNil())
				Expect(summer.MediaFiles).To(HaveLen(2)) // vacation1.jpg, vacation2.png (notes.txt excluded)
				Expect(summer.Children).To(HaveLen(1))   // europe
				Expect(summer.Parent).To(Equal(child2023))

				// Find europe child of summer
				europe := summer.Children[0]
				Expect(europe.Path).To(Equal("photos/2023/summer/europe"))
				Expect(europe.MediaFiles).To(HaveLen(2)) // paris.jpg, rome.jpeg
				Expect(europe.Children).To(HaveLen(0))
				Expect(europe.Parent).To(Equal(summer))

				// Verify total counts
				totalFolders := tree.GetTotalFolderCount()
				Expect(totalFolders).To(Equal(9)) // photos, 2023, summer, europe, winter, alps, 2024, work, projects

				totalMedia := tree.GetTotalMediaCount()
				Expect(totalMedia).To(Equal(14)) // All jpg, jpeg, JPG, png files (2+1+2+2+1+2+1+1+2)
			})
		})

		Context("Empty directory", func() {
			It("creates tree with no children or media", func() {
				// Create empty directory
				createTestStructure(map[string][]string{
					"empty": {},
				})

				tree, err := datastore.WalkTree(ctx, "empty")
				Expect(err).To(BeNil())
				Expect(tree).ToNot(BeNil())
				Expect(tree.Path).To(Equal("empty"))
				Expect(tree.Children).To(HaveLen(0))
				Expect(tree.MediaFiles).To(HaveLen(0))
			})
		})

		Context("Directory with only subdirectories (no media)", func() {
			It("creates tree with children but no media files", func() {
				createTestStructure(map[string][]string{
					"folders":      {"document.txt"}, // Non-media file
					"folders/sub1": {"readme.md"},    // Non-media file
					"folders/sub2": {},               // Empty
					"folders/sub3": {"config.json"},  // Non-media file
				})

				tree, err := datastore.WalkTree(ctx, "folders")
				Expect(err).To(BeNil())
				Expect(tree.Path).To(Equal("folders"))
				Expect(tree.MediaFiles).To(HaveLen(0)) // No media files
				Expect(tree.Children).To(HaveLen(3))   // sub1, sub2, sub3

				// Verify all children have no media files
				for _, child := range tree.Children {
					Expect(child.MediaFiles).To(HaveLen(0))
				}
			})
		})

		Context("Directory with mixed file types", func() {
			It("only includes supported media files", func() {
				createTestStructure(map[string][]string{
					"mixed": {
						"photo1.jpg",   // ✓ included
						"photo2.JPEG",  // ✓ included (case insensitive)
						"photo3.png",   // ✓ included
						"photo4.PNG",   // ✓ included (case insensitive)
						"video.mp4",    // ✗ not supported
						"document.pdf", // ✗ not supported
						"archive.zip",  // ✗ not supported
						"script.sh",    // ✗ not supported
						"image.gif",    // ✗ not supported yet
						"photo5.JPG",   // ✓ included
					},
				})

				tree, err := datastore.WalkTree(ctx, "mixed")
				Expect(err).To(BeNil())
				Expect(tree.MediaFiles).To(HaveLen(5)) // Only jpg, jpeg, png files

				// Verify specific files are included
				mediaFiles := tree.MediaFiles
				Expect(mediaFiles).To(ContainElement("mixed/photo1.jpg"))
				Expect(mediaFiles).To(ContainElement("mixed/photo2.JPEG"))
				Expect(mediaFiles).To(ContainElement("mixed/photo3.png"))
				Expect(mediaFiles).To(ContainElement("mixed/photo4.PNG"))
				Expect(mediaFiles).To(ContainElement("mixed/photo5.JPG"))
			})
		})

		Context("Error cases", func() {
			It("returns error for non-existent directory", func() {
				tree, err := datastore.WalkTree(ctx, "non-existent")
				Expect(err).ToNot(BeNil())
				Expect(tree).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("path does not exist"))
			})

			It("returns error for file instead of directory", func() {
				// Create a file
				filePath := filepath.Join(tmpDir, "file.txt")
				err := os.WriteFile(filePath, []byte("content"), 0644)
				Expect(err).To(BeNil())

				tree, err := datastore.WalkTree(ctx, "file.txt")
				Expect(err).ToNot(BeNil())
				Expect(tree).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("path is not a directory"))
			})
		})

		Context("Tree structure verification", func() {
			It("maintains correct parent-child relationships", func() {
				createTestStructure(map[string][]string{
					"root":               {"root.jpg"},
					"root/level1":        {"level1.jpg"},
					"root/level1/level2": {"level2.jpg"},
				})

				tree, err := datastore.WalkTree(ctx, "root")
				Expect(err).To(BeNil())

				// Verify parent-child relationships
				level1 := tree.Children[0]
				Expect(level1.Parent).To(Equal(tree))

				level2 := level1.Children[0]
				Expect(level2.Parent).To(Equal(level1))

				// Verify no circular references
				Expect(tree.Parent).To(BeNil())
			})

			It("provides correct tree traversal methods", func() {
				createTestStructure(map[string][]string{
					"traverse":      {"1.jpg", "2.jpg"},
					"traverse/a":    {"a1.jpg"},
					"traverse/a/a1": {"a1a.jpg"},
					"traverse/b":    {"b1.jpg", "b2.jpg"},
				})

				tree, err := datastore.WalkTree(ctx, "traverse")
				Expect(err).To(BeNil())

				// Test GetAllNodes
				allNodes := tree.GetAllNodes()
				Expect(allNodes).To(HaveLen(4)) // traverse, a, a1, b (4 folders total)

				// Test GetTotalFolderCount
				totalFolders := tree.GetTotalFolderCount()
				Expect(totalFolders).To(Equal(4)) // Same as GetAllNodes().length

				// Test GetTotalMediaCount
				totalMedia := tree.GetTotalMediaCount()
				Expect(totalMedia).To(Equal(6)) // 2 + 1 + 1 + 2
			})
		})
	})
})
