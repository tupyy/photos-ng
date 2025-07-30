# Photos NG API

A RESTful API for managing albums and media in the Photos NG application. This API provides comprehensive CRUD operations for organizing and accessing photo and video collections.

## Overview

The Photos NG API is built using OpenAPI 3.0 specification and provides endpoints for:
- Managing photo albums and their metadata
- Accessing and organizing media files (photos and videos)
- Timeline-based media browsing
- File system synchronization

## API Specification

The full OpenAPI specification is available at: `api/v1/openapi.yaml`

### Base URLs
- **Development**: `http://localhost:8080`
- **Production**: `https://api.photos-ng.com`

## Endpoints

### Albums

#### `GET /api/v1/albums`
List all albums with pagination support.

**Parameters:**
- `limit` (optional): Maximum number of albums (1-100, default: 20)
- `offset` (optional): Number of albums to skip (default: 0)

**Response:** Paginated list of albums

#### `POST /api/v1/albums`
Create a new album.

**Request Body:**
```json
{
  "name": "string",
  "path": "string"
}
```

**Response:** Created album object

#### `GET /api/v1/albums/{id}`
Retrieve a specific album by ID.

**Parameters:**
- `id` (path): Album UUID

**Response:** Album object

#### `PUT /api/v1/albums/{id}`
Update an album's metadata.

**Request Body:**
```json
{
  "name": "string"
}
```

**Response:** Updated album object

#### `DELETE /api/v1/albums/{id}`
Delete an album.

**Parameters:**
- `id` (path): Album UUID

**Response:** 204 No Content

#### `POST /api/v1/albums/{id}/sync`
Synchronize an album with the file system.

**Parameters:**
- `id` (path): Album UUID

**Response:**
```json
{
  "message": "Album sync completed",
  "synced_items": 42
}
```

### Media

#### `GET /api/v1/media`
List all media items with advanced filtering and sorting.

**Parameters:**
- `limit` (optional): Maximum number of media items (1-100, default: 20)
- `offset` (optional): Number of media items to skip (default: 0)
- `album_id` (optional): Filter by album UUID
- `type` (optional): Filter by media type (`photo` | `video`)
- `startDate` (optional): Filter media captured on/after date (DD/MM/YYYY)
- `endDate` (optional): Filter media captured on/before date (DD/MM/YYYY)
- `sortBy` (optional): Sort field (`capturedAt` | `filename` | `type`, default: `capturedAt`)
- `sortOrder` (optional): Sort direction (`asc` | `desc`, default: `desc`)

**Response:** Paginated list of media items

#### `GET /api/v1/media/{id}`
Retrieve a specific media item by ID.

**Parameters:**
- `id` (path): Media UUID

**Response:** Media object

#### `PUT /api/v1/media/{id}`
Update media metadata.

**Request Body:**
```json
{
  "capturedAt": "01/01/2024",
  "exif": [
    {
      "key": "string",
      "value": "string"
    }
  ]
}
```

**Response:** Updated media object

#### `DELETE /api/v1/media/{id}`
Delete a media item.

**Parameters:**
- `id` (path): Media UUID

**Response:** 204 No Content

#### `GET /api/v1/media/{id}/thumbnail`
Retrieve the thumbnail image for a media item.

**Parameters:**
- `id` (path): Media UUID

**Response:** Binary image data

#### `GET /api/v1/media/{id}/content`
Retrieve the full content of a media item.

**Parameters:**
- `id` (path): Media UUID

**Response:** Binary image/video data

### Timeline

#### `GET /api/v1/timeline`
Retrieve a timeline of media organized in buckets.

**Parameters:**
- `startDate` (required): Start date for timeline (DD/MM/YYYY)
- `limit` (optional): Maximum number of buckets (1-100, default: 20)
- `offset` (optional): Number of buckets to skip (default: 0)

**Response:** Paginated list of timeline buckets with available years

**Example Response:**
```json
{
  "buckets": [
    {
      "date": "01/01/2024",
      "media": ["media_href_1", "media_href_2"]
    }
  ],
  "years": [2020, 2021, 2022, 2023, 2024],
  "total": 25,
  "limit": 20,
  "offset": 0
}
```

## Data Models

### Album
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "href": "/api/v1/albums/album_id",
  "name": "string",
  "path": "string",
  "parentHref": "string",
  "children": [
    {
      "href": "string",
      "name": "string"
    }
  ],
  "media": ["string"]
}
```

### Media
```json
{
  "id": "string",
  "href": "/api/v1/media/some_id",
  "albumHref": "/api/v1/albums/album_id",
  "capturedAt": "01/01/2024",
  "type": "photo | video",
  "filename": "string",
  "thumbnail": "/api/v1/media/{id}/thumbnail",
  "content": "/api/v1/media/{id}/content",
  "exif": [
    {
      "key": "string",
      "value": "string"
    }
  ]
}
```

### Bucket
```json
{
  "date": "01/01/2024",
  "media": ["string"]
}
```

### ExifHeader
```json
{
  "key": "string",
  "value": "string"
}
```

## Features

### Pagination
All list endpoints support pagination through `limit` and `offset` parameters:
- `limit`: Controls the number of items returned (1-100)
- `offset`: Controls how many items to skip
- Responses include `total`, `limit`, and `offset` metadata

### Filtering
Media endpoints support advanced filtering:
- **Album filtering**: Filter media by album ID
- **Type filtering**: Filter by media type (photo/video)
- **Date range filtering**: Filter by capture date range
- **Combination filtering**: All filters can be combined

### Sorting
Media list endpoint supports sorting by:
- `capturedAt`: Date the media was captured (default)
- `filename`: Filename of the media
- `type`: Media type (photo/video)

Sort direction can be `asc` (ascending) or `desc` (descending, default).

### Error Handling

The API uses standard HTTP status codes:
- `200`: Success
- `201`: Created
- `204`: No Content (successful deletion)
- `400`: Bad Request
- `404`: Not Found
- `500`: Internal Server Error

Error responses include structured error objects:
```json
{
  "message": "Error description",
  "code": "ERROR_CODE",
  "details": {}
}
```

## Development

### Build and Run

Use the provided Makefile for common tasks:

```bash
# Generate code from OpenAPI spec
make generate

# Build the application
make build

# Run the application
make run

# Clean build artifacts
make clean

# Show help
make help
```

### API Documentation

The OpenAPI specification can be used to generate:
- Interactive API documentation (Swagger UI, Redoc)
- Client SDKs in various languages
- Server stubs and mock servers
- API testing tools

### File Organization

```
api/v1/
├── openapi.yaml          # OpenAPI 3.0 specification
└── [generated files]     # Auto-generated from spec

cmd/
├── migrate.go           # Database migration command
└── serve.go            # Server command

config/                 # Configuration management
internal/               # Internal application code
pkg/                   # Shared packages
server/                # HTTP server implementation
```

## License

[Add your license information here]

## Contributing

[Add contributing guidelines here]