/*
Package job provides the sync job system for Photos NG.

# Overview

The job package implements a synchronization system that scans filesystem directories
and creates corresponding album structures in the database, processing media files
along the way. The core component is SyncAlbumJob, which handles the discovery and
processing of photos organized in directory hierarchies.

# Key Concepts

## Root Album and Data Folder

The sync job operates on a "root album" concept that determines the starting point
for directory scanning:

  - **Non-empty root path** (e.g., "photos/2023"): The job treats this as an existing
    album in the database and discovers subdirectories as child albums.

  - **Empty root path** (""): The job starts from the root data folder itself. This
    represents scanning the entire data directory, where top-level directories become
    root albums in the database. The root data folder itself is NOT an album.

## Important Constraints

  - **Root data folder never contains media files**: By design, media files only exist
    in subdirectories that become albums. The root data folder serves purely as a
    container for album directories.

  - **Relative paths**: All services work with paths relative to the configured root
    data folder. The filesystem datastore handles the translation to absolute paths.

# Architecture

## SyncAlbumJob Structure

The SyncAlbumJob coordinates between multiple services:

- **AlbumService**: Creates and manages album records in the database
- **MediaService**: Processes and stores media file records
- **Filesystem Datastore**: Provides directory traversal and file access
- **No ProcessingService**: Media processing is handled directly by MediaService

## Discovery Process

 1. **Tree Walking**: Uses fs.Datastore.WalkTree() to build a complete FolderNode tree
    representing the directory structure with all subdirectories and media files.

2. **Task Generation**: Converts the tree into discrete tasks:
  - Album creation tasks for each discovered directory
  - Media processing tasks for each media file found

3. **Parent-Child Relationships**: Maintains proper album hierarchy:
  - For empty root path: top-level directories become root albums (ParentId = nil)
  - For non-empty root path: discovered albums become children of the root album

## Task Execution

Tasks are executed sequentially with proper error handling:

- **Album Tasks**: Create album records with correct parent relationships
- **Media Tasks**: Process media files and associate them with their containing album
- **Error Isolation**: Individual task failures don't stop the entire job

# Usage Examples

## Sync from Root Data Folder

```go
// Start from root data folder - discovers all top-level directories as albums
rootAlbum := entity.NewAlbum("")  // Empty path = root data folder
job, err := NewSyncJob(rootAlbum, albumService, mediaService, fsDatastore)
err = job.Start(ctx)
```

This will:
- Scan the entire data directory
- Create albums for top-level directories (2023, 2024, etc.)
- Create sub-albums for nested directories
- Process all media files in subdirectories
- Skip any files in the root data folder (there shouldn't be any)

## Sync Specific Album

```go
// Start from specific album - discovers subdirectories as child albums
rootAlbum := entity.NewAlbum("photos/2023")
job, err := NewSyncJob(rootAlbum, albumService, mediaService, fsDatastore)
err = job.Start(ctx)
```

This will:
- Scan only the photos/2023 directory
- Create child albums for subdirectories (summer, winter, etc.)
- Process media files in the root album and all subdirectories

# Directory Structure Example

Given this filesystem structure:
```
/data/root/
├── 2023/
│   ├── summer/
│   │   ├── beach.jpg
│   │   └── vacation.jpg
│   └── winter/
│       └── skiing.jpg
├── 2024/
│   └── spring/
│       └── flowers.jpg
└── documents/

	└── readme.txt (ignored - not media)

```

## Empty Root Path Sync

Starting with `entity.NewAlbum("")`:
- Creates albums: "2023", "2023/summer", "2023/winter", "2024", "2024/spring", "documents"
- Top-level albums (2023, 2024, documents) have ParentId = nil
- Processes 4 media files
- Ignores non-media files

## Specific Album Sync

Starting with `entity.NewAlbum("2023")`:
- Creates albums: "2023/summer", "2023/winter" (children of existing "2023" album)
- Processes 3 media files in 2023 and its subdirectories

# File Organization

- `job.go`: Core SyncAlbumJob implementation and task processing
- `scheduler.go`: Job scheduling interface (separate from SyncAlbumJob)
- `job_test.go`: Comprehensive tests covering various directory structures
- `doc.go`: This documentation file

# Error Handling

The job system is designed to be resilient:

- **Discovery failures**: Stop the job immediately (e.g., root path doesn't exist)
- **Individual task failures**: Log errors but continue processing other tasks
- **Database constraints**: Handle foreign key violations gracefully
- **File access errors**: Skip problematic files and continue

# Testing

The test suite covers:
- Empty directory structures
- Single-level and multi-level nested albums
- Empty root path scenarios (scanning from data folder)
- Media file processing and error handling
- Parent-child relationship validation
- Edge cases and error conditions

See `job_test.go` for comprehensive examples and test cases.
*/
package job
