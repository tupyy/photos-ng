package fs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

// WalkResult represents an item found during filesystem traversal
type WalkResult struct {
	Path        string // Relative path from the root data folder
	IsDirectory bool   // True if this is a directory, false if it's a file
}

// Common filter functions for Walk method
var (
	// FilterDirectories returns only directories
	FilterDirectories = func(result WalkResult) bool {
		return result.IsDirectory
	}

	// FilterFiles returns only files (not directories)
	FilterFiles = func(result WalkResult) bool {
		return !result.IsDirectory
	}

	// FilterMediaFiles returns only supported media files (jpg, jpeg, png)
	FilterMediaFiles = func(result WalkResult) bool {
		if result.IsDirectory {
			return false
		}

		supportedExtensions := map[string]bool{
			".jpg":  true,
			".jpeg": true,
			".png":  true,
		}

		ext := strings.ToLower(filepath.Ext(result.Path))
		return supportedExtensions[ext]
	}

	// FilterAll returns all items (directories and files)
	FilterAll = func(result WalkResult) bool {
		return true
	}
)

type Datastore struct {
	rootFolder string
}

func NewFsDatastore(rootFolder string) *Datastore {
	return &Datastore{rootFolder: rootFolder}
}

func (fs *Datastore) Read(ctx context.Context, filepath string) entity.MediaContentFn {
	return func() (io.Reader, error) {
		data, err := os.ReadFile(path.Join(fs.rootFolder, filepath))
		if err != nil {
			return nil, err
		}

		return bytes.NewReader(data), nil
	}
}

func (fs *Datastore) Write(ctx context.Context, filePath string, r io.Reader) error {
	fullPath := filepath.Join(fs.rootFolder, filePath)

	// Create the file (album folder should already exist)
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy content to file
	_, err = io.Copy(file, r)
	return err
}

// CreateFolder creates a directory on the filesystem.
// This operation is idempotent - it will not fail if the directory already exists.
func (fs *Datastore) CreateFolder(ctx context.Context, folderPath string) error {
	fullPath := filepath.Join(fs.rootFolder, folderPath)

	// os.MkdirAll is idempotent - it creates the directory and any necessary parents,
	// and returns nil if the path already exists as a directory
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return err
	}

	// Verify the path exists and is actually a directory
	if info, err := os.Stat(fullPath); err != nil {
		return err
	} else if !info.IsDir() {
		return os.ErrExist // Path exists but is not a directory
	}

	return nil
}

// DeleteFolder removes a directory from the filesystem.
// This operation is idempotent - it will not fail if the directory doesn't exist.
func (fs *Datastore) DeleteFolder(ctx context.Context, folderPath string) error {
	fullPath := filepath.Join(fs.rootFolder, folderPath)

	// Check if the path exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// Directory doesn't exist, nothing to delete (idempotent)
		return nil
	} else if err != nil {
		// Other error occurred
		return err
	}

	// Remove the directory and all its contents
	return os.RemoveAll(fullPath)
}

func (fs *Datastore) DeleteMedia(ctx context.Context, mediapath string) error {
	fullPath := filepath.Join(fs.rootFolder, mediapath)

	// Check if the path exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// Directory doesn't exist, nothing to delete (idempotent)
		return nil
	} else if err != nil {
		// Other error occurred
		return err
	}

	// Remove the directory and all its contents
	return os.Remove(fullPath)
}

// Walk recursively traverses the filesystem starting from the given relative path
// and returns a list of items (directories and files) as WalkResult structs that pass the filter
// The filter function receives each WalkResult and returns true if the item should be included

func (fs *Datastore) Walk(ctx context.Context, relativePath string, filter func(WalkResult) bool) ([]WalkResult, error) {
	fullPath := filepath.Join(fs.rootFolder, relativePath)

	// Check if the path exists and is a directory
	if info, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return []WalkResult{}, nil // Path doesn't exist, return empty list
		}
		return nil, err
	} else if !info.IsDir() {
		return []WalkResult{}, nil // Path is not a directory, return empty list
	}

	var results []WalkResult

	// Recursively walk the directory tree
	err := filepath.WalkDir(fullPath, func(path string, d os.DirEntry, err error) error {
		// Handle walk errors
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == fullPath {
			return nil
		}

		// Create relative path from the root data folder
		itemRelativePath, err := filepath.Rel(fs.rootFolder, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		result := WalkResult{
			Path:        itemRelativePath,
			IsDirectory: d.IsDir(),
		}

		// Apply filter - only add if filter returns true
		if filter == nil || filter(result) {
			results = append(results, result)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// WalkTree recursively walks the directory tree and returns a FolderNode tree structure
// The tree includes all directories and media files organized in a hierarchical structure
func (fs *Datastore) WalkTree(ctx context.Context, relativePath string) (*entity.FolderNode, error) {
	fullPath := filepath.Join(fs.rootFolder, relativePath)

	// Check if the path exists and is a directory
	if info, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("path does not exist: %s", relativePath)
		}
		return nil, err
	} else if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", relativePath)
	}

	// Create the root node
	root := entity.NewFolderNode(relativePath)

	// Build a map to track all folder nodes by their path
	folderNodes := make(map[string]*entity.FolderNode)
	folderNodes[relativePath] = root

	// Walk the directory tree
	err := filepath.WalkDir(fullPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == fullPath {
			return nil
		}

		// Get relative path from fs root
		itemRelativePath, err := filepath.Rel(fs.rootFolder, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		if d.IsDir() {
			// Create a new folder node
			folderNode := entity.NewFolderNode(itemRelativePath)
			folderNodes[itemRelativePath] = folderNode

			// Find the parent folder and add this as a child
			parentPath := filepath.Dir(itemRelativePath)
			// Handle the case where parent is the root - filepath.Dir returns "." for top-level items
			if parentPath == "." {
				parentPath = relativePath // Use the original relative path (could be "")
			}
			if parentNode, exists := folderNodes[parentPath]; exists {
				parentNode.AddChild(folderNode)
			}
		} else {
			// This is a file - check if it's a media file and add to parent folder
			if fs.isMediaFile(itemRelativePath) {
				parentPath := filepath.Dir(itemRelativePath)
				// Handle the case where parent is the root - filepath.Dir returns "." for top-level items
				if parentPath == "." {
					parentPath = relativePath // Use the original relative path (could be "")
				}
				if parentNode, exists := folderNodes[parentPath]; exists {
					parentNode.AddMediaFile(itemRelativePath)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return root, nil
}

// isMediaFile checks if a file is a supported media file based on extension
func (fs *Datastore) isMediaFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".jpg", ".jpeg", ".png":
		return true
	default:
		return false
	}
}
