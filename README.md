# Photos NG

Photos NG is a self-hosted photo and video management application designed for organizing and browsing personal media collections. The application provides a web-based interface for managing hierarchical album structures, viewing media with metadata, and synchronizing with filesystem changes.

Photos and videos are stored in folders on the local filesystem. Network-attached storage such as NFS can be mounted and used as the storage backend, allowing flexible deployment options.

Built with a Go backend and React frontend, Photos NG emphasizes performance through efficient pagination strategies and on-demand media loading. The system stores metadata in PostgreSQL while media files remain on the filesystem, enabling fast queries while preserving original file organization.

The backend exposes two API interfaces: a RESTful HTTP API and a gRPC API. Both APIs provide the same functionality. The HTTP API is consumed by the web frontend, while the gRPC API can be used for mobile applications on Android and iOS platforms.

## Authentication and Authorization

This application does not implement authentication or authorization mechanisms. It is intended for deployment in trusted, internal networks only. Access control should be managed at the network or reverse proxy level if needed.

## Features

### Album Management
- Hierarchical album organization with parent-child relationships
- File system synchronization to keep albums in sync with directory structures

### Media Organization
- Support for photos and videos with automatic type detection
- EXIF metadata extraction and display
- Automatic thumbnail generation for fast browsing
- Sortable views by capture date, filename, or type

### User Interface
- Responsive design for desktop and mobile browsers
- Dark mode support
- Client-side pagination with on-demand media loading
- Timeline view with year-based navigation

## Road map

- improve http header's middleware. Currently, there's almost none. see https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Headers_Cheat_Sheet.html
- authentication. Only oauth2 is considered.
- authorization. not really sure I need this.

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

**Media Pagination**: The API implements cursor-based pagination for media endpoints to efficiently handle large collections. The cursor encodes the position in the result set using two fields: `captured_at` timestamp and `id`.

Example cursor structure (base64 encoded):
```json
{
  "captured_at": "2024-01-15T10:30:00Z",
  "id": "abc123"
}
```

When listing media, clients receive a `nextCursor` in the response that can be used to fetch the next page:
```
GET /api/v1/media?limit=20
GET /api/v1/media?limit=20&cursor=eyJjYXB0dXJlZF9hdCI6IjIwMjQtMDEtMTV...
```

**Album Pagination**: Albums use simple offset-based pagination with `limit` and `offset` query parameters:
```
GET /api/v1/albums?limit=20&offset=0
GET /api/v1/albums?limit=20&offset=20
```

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

## Server Configuration

The `serve` command starts both HTTP and gRPC servers simultaneously. Configuration can be provided through command-line flags.

### Database Flags

- `--db-conn-uri` - PostgreSQL connection string (default: `postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable`)
- `--db-ssl-mode` - Enable SSL mode for database connection

### Server Flags

- `--http-port` - Port for HTTP server (default: 8080)
- `--grpc-port` - Port for gRPC server (default: 9090)
- `--server-mode` - Server mode: `dev` or `prod` (default: `dev`)
- `--server-gin-mode` - Gin mode: `release` or `debug` (default: `debug`)
- `--data-root-folder` - Path to root folder containing media files (required)
- `--statics-folder` - Path to static files for web UI (required in prod mode)

### Global Flags

- `--log-format` - Log format: `console` or `json` (default: `console`)
- `--log-level` - Log level: `debug`, `info`, `warn`, `error` (default: `debug`)

### Example

```bash
./photos-ng serve \
  --db-conn-uri "postgres://user:pass@localhost:5432/photos" \
  --http-port 8080 \
  --grpc-port 9090 \
  --data-root-folder /mnt/photos \
  --server-mode prod \
  --statics-folder ./ui/dist \
  --log-level info
```

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.
