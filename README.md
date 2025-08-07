# Photos NG

A full-stack photo management application built with Go backend and React frontend. Photos NG provides comprehensive photo and video organization, browsing, and management capabilities with a modern, responsive user interface.

## Overview

Photos NG consists of:
- **Backend API**: RESTful API built with Go, Gin, and PostgreSQL
- **Frontend UI**: Modern React application with TypeScript and Tailwind CSS
- **Database**: PostgreSQL for metadata storage
- **File Storage**: Local filesystem for media files

## Features

### Core Functionality
- 📁 **Album Management**: Create, organize, and manage photo albums with hierarchical structure
- 🖼️ **Media Organization**: Upload, view, and organize photos and videos
- 📊 **Statistics Dashboard**: View total media counts and available years
- ⏱️ **Timeline View**: Browse media by date with year navigation
- 🔍 **Advanced Filtering**: Filter by album, media type, and date ranges
- 📱 **Responsive Design**: Works seamlessly on desktop and mobile devices

### Recent Improvements
- ✨ **Enhanced Pagination**: Client-side pagination for album media views
- 🚀 **Performance Optimization**: Efficient media loading using href-based pagination
- 📈 **Accurate Counters**: Album titles show total media count across all pages
- 🎯 **Smart Navigation**: Improved user experience with proper loading states

### Technical Features
- 🔄 **File System Sync**: Synchronize albums with file system changes
- 🏷️ **EXIF Data Support**: Extract and display photo metadata
- 🖼️ **Thumbnail Generation**: Automatic thumbnail creation for fast loading
- 📤 **Bulk Operations**: Select and manage multiple media items
- 🎨 **Dark Mode**: Full dark mode support throughout the application

## Architecture

### Frontend (React + TypeScript)
- **Location**: `ui/` directory
- **Framework**: React 18 with TypeScript
- **Styling**: Tailwind CSS for responsive design
- **State Management**: Redux Toolkit for centralized state
- **API Client**: Auto-generated from OpenAPI specification
- **Build Tool**: Webpack with custom configuration

### Backend (Go)
- **Location**: Root directory
- **Framework**: Gin web framework
- **Database**: PostgreSQL with custom query builder
- **Code Generation**: OpenAPI code generation for consistency
- **Architecture**: Clean architecture with separate layers (handlers, services, datastore)

### Pagination Implementation

Photos NG uses an innovative pagination approach for album media:

1. **Album Data**: Albums contain arrays of media href references
2. **Client-Side Pagination**: Media hrefs are paginated locally using JavaScript
3. **On-Demand Loading**: Only media objects for the current page are fetched via API
4. **Performance**: Reduces server load and improves user experience

```
Album contains: ["/api/v1/media/1", "/api/v1/media/2", ..., "/api/v1/media/100"]
Page 1: Fetch media objects for hrefs 1-20
Page 2: Fetch media objects for hrefs 21-40
```

This approach provides:
- ✅ Fast initial album loading
- ✅ Efficient memory usage
- ✅ Accurate pagination counts
- ✅ Reduced API calls

## Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- PostgreSQL 14+

### Backend Setup
```bash
# Clone and build backend
make build

# Setup database (update connection string in config)
make migrate

# Run backend server
make run
```

### Frontend Setup
```bash
# Install dependencies
cd ui && npm install

# Generate API client from OpenAPI spec
npm run generate:api

# Start development server
npm run dev

# Or build for production
npm run build
```

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

### Statistics

#### `GET /api/v1/stats`
Retrieve application statistics including media and album counts.

**Response:** Application statistics

**Example Response:**
```json
{
  "years": [2024, 2023, 2022],
  "countMedia": 1250,
  "countAlbum": 45
}
```

**Properties:**
- `years`: Array of years that contain media (sorted descending)
- `countMedia`: Total number of media items in the system
- `countAlbum`: Total number of albums in the system

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

### StatsResponse
```json
{
  "years": [2024, 2023, 2022],
  "countMedia": 1250,
  "countAlbum": 45
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
photos-ng/
├── api/v1/                    # API specification and generated code
│   ├── openapi.yaml          # OpenAPI 3.0 specification
│   ├── extensions.go         # Custom API extensions
│   └── [generated files]     # Auto-generated from spec
├── cmd/                      # CLI commands
│   ├── migrate.go           # Database migration command
│   └── serve.go            # Server startup command
├── internal/                # Internal application code
│   ├── config/             # Configuration management
│   ├── datastore/          # Database layer (PostgreSQL)
│   ├── entity/             # Domain models
│   ├── handlers/           # HTTP request handlers
│   ├── server/             # HTTP server setup
│   └── services/           # Business logic layer
├── pkg/                     # Shared packages
│   ├── encoder/            # Media processing
│   ├── logger/             # Logging utilities
│   └── processing/         # Background processing
├── ui/                      # Frontend React application
│   ├── src/
│   │   ├── generated/      # Auto-generated API client
│   │   ├── pages/          # React page components
│   │   ├── shared/         # Shared components and utilities
│   │   └── ...
│   ├── package.json        # Frontend dependencies
│   └── webpack.*.js        # Build configuration
├── main.go                  # Application entry point
├── Makefile                # Build automation
└── README.md               # This file
```

## Recent Changes

### v1.2.0 - Pagination & Performance Improvements
- ✨ **Smart Pagination**: Implemented client-side pagination for album media views
- 🚀 **Performance**: Reduced API calls by using href-based media loading
- 📊 **Statistics**: Added `/stats` endpoint replacing timeline functionality
- 🎯 **UX**: Album titles now show accurate total media counts
- 🐛 **Fixes**: Corrected album media counting issues with JOIN queries

### v1.1.0 - Album Enhancements
- 📁 **Hierarchical Albums**: Support for nested album structures
- 🖼️ **Media Management**: Improved media organization within albums
- 🔄 **File Sync**: Enhanced file system synchronization
- 🎨 **UI Polish**: Better responsive design and dark mode support

## Roadmap

### Planned Features
- 🔍 **Search**: Full-text search across media metadata and filenames
- 🏷️ **Tagging**: Add custom tags to media items for better organization
- 👥 **User Management**: Multi-user support with authentication
- ☁️ **Cloud Storage**: Support for cloud storage backends (S3, etc.)
- 📱 **Mobile App**: Native mobile applications for iOS and Android
- 🤖 **AI Features**: Automatic tagging using computer vision
- 📊 **Analytics**: Advanced usage analytics and insights

### Technical Improvements
- 🔧 **API Versioning**: Better API version management
- 🧪 **Testing**: Comprehensive test coverage
- 📦 **Containerization**: Docker support for easy deployment
- 🔒 **Security**: Enhanced security features and audit logging
- ⚡ **Caching**: Redis-based caching for improved performance

## License

[Add your license information here]

## Contributing

[Add contributing guidelines here]