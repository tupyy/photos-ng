# Photos NG - Redux Implementation

This document explains the Redux implementation using Redux Toolkit and the generated OpenAPI client.

## Architecture Overview

### Generated API Client
- Located in `src/generated/` - Auto-generated from OpenAPI spec
- Contains TypeScript models and Axios-based API clients
- Regenerated with `npm run generate:api`

### Redux Store Structure
```
src/shared/
├── api/
│   └── apiConfig.ts          # API configuration and instances
├── reducers/
│   ├── albumsSlice.ts        # Albums state management
│   ├── mediaSlice.ts         # Media state management
│   ├── timelineSlice.ts      # Timeline state management
│   └── index.ts              # Root reducer
├── hooks/
│   └── useApi.ts             # Custom hooks for API operations
└── store.ts                  # Store configuration and selectors
```

## Usage Examples

### 1. Albums Management

```tsx
import { useAlbumsApi } from '../shared/hooks/useApi';

const AlbumsComponent = () => {
  const {
    albums,
    loading,
    error,
    fetchAlbums,
    createAlbum,
    deleteAlbum,
    syncAlbum
  } = useAlbumsApi();

  useEffect(() => {
    fetchAlbums({ limit: 20, offset: 0 });
  }, [fetchAlbums]);

  const handleCreateAlbum = async () => {
    await createAlbum({ name: 'New Album' });
  };

  // Component JSX...
};
```

### 2. Media Management

```tsx
import { useMediaApi } from '../shared/hooks/useApi';

const MediaComponent = () => {
  const {
    media,
    filters,
    loading,
    fetchMedia,
    setFilters,
    toggleMediaSelection
  } = useMediaApi();

  const handleFilterChange = (newFilters) => {
    setFilters(newFilters);
    fetchMedia(newFilters);
  };

  // Component JSX...
};
```

### 3. Timeline Management

```tsx
import { useTimelineApi } from '../shared/hooks/useApi';

const TimelineComponent = () => {
  const {
    buckets,
    years,
    selectedYear,
    fetchTimeline,
    setSelectedYear
  } = useTimelineApi();

  const handleYearSelect = (year) => {
    setSelectedYear(year);
    fetchTimeline({ startDate: `01/01/${year}`, endDate: `31/12/${year}` });
  };

  // Component JSX...
};
```

## Available Scripts

### API Generation
- `npm run generate:api` - Generate API client from OpenAPI spec
- `npm run generate:api:clean` - Clean and regenerate API client
- `npm run generate:api:watch` - Watch for changes and auto-regenerate

### Development
- `npm run start:dev` - Start development server with Redux DevTools
- `npm run build:prod` - Production build with optimizations

## State Management Features

### Albums Slice
- **State**: albums list, current album, pagination, loading states
- **Actions**: CRUD operations, sync functionality, filtering
- **Async Thunks**: fetchAlbums, createAlbum, updateAlbum, deleteAlbum, syncAlbum

### Media Slice
- **State**: media list, filters, selection, view mode
- **Actions**: CRUD operations, filtering, selection management
- **Features**: Multi-select, view modes (grid/list), advanced filtering

### Timeline Slice
- **State**: timeline buckets, years, date navigation
- **Actions**: Date filtering, year/month navigation
- **Features**: Chronological organization, date range filtering

## Error Handling

All async actions include proper error handling with user-friendly messages:

```tsx
// Error states are automatically managed
const { error, clearError } = useAlbumsApi();

if (error) {
  return (
    <div className="error-message">
      <p>{error}</p>
      <button onClick={clearError}>Dismiss</button>
    </div>
  );
}
```

## TypeScript Integration

- Full TypeScript support with generated types
- Type-safe Redux actions and state
- IntelliSense support for API methods and data structures

## Environment Configuration

Set API base URL in environment variables:
```bash
# Development
REACT_APP_API_URL=http://localhost:8080

# Production  
REACT_APP_API_URL=https://api.photos-ng.com
```

## Performance Optimizations

- **Memoized selectors** for efficient state derivation
- **Lazy loading** for large media lists
- **Pagination support** for all list endpoints
- **Background sync** for album operations