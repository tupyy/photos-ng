package fs

import (
	"bytes"
	"context"
	"io"
	"os"
	"path"
	"path/filepath"

	"git.tls.tupangiu.ro/cosmin/photos-ng/internal/entity"
)

type FsDatastore struct {
	rootFolder string
}

func NewFsDatastore(rootFolder string) *FsDatastore {
	return &FsDatastore{rootFolder: rootFolder}
}

func (fs *FsDatastore) Read(ctx context.Context, filepath string) entity.MediaContentFn {
	return func() (io.Reader, error) {
		data, err := os.ReadFile(path.Join(fs.rootFolder, filepath))
		if err != nil {
			return nil, err
		}

		return bytes.NewReader(data), nil
	}
}

func (fs *FsDatastore) Write(ctx context.Context, filePath string, r io.Reader) error {
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
func (fs *FsDatastore) CreateFolder(ctx context.Context, folderPath string) error {
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
func (fs *FsDatastore) DeleteFolder(ctx context.Context, folderPath string) error {
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
