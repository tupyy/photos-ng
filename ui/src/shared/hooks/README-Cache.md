# Cache-Aware Media Reducers

This solution implements intelligent HTTP caching for the media reducers, respecting server-side `Cache-Control` headers and preventing unnecessary API calls.

## Features

- **Automatic cache validation**: Respects `Cache-Control`, `ETag`, and `Last-Modified` headers
- **Conditional requests**: Uses `If-None-Match` and `If-Modified-Since` headers when appropriate
- **Smart fetching**: Automatically skips API calls when cached data is still valid
- **Cache invalidation**: Automatically invalidates cache when data is modified
- **Background cleanup**: Periodically removes expired cache entries

## Usage Examples

### Basic Usage with Smart Caching

```tsx
import { useCachedMedia } from '@shared/hooks/useCachedMedia';

function MediaGallery() {
  const { 
    media, 
    loading, 
    fetchMediaSmart, 
    isCached,
    getCacheInfo 
  } = useCachedMedia();

  useEffect(() => {
    // This will use cache if valid, otherwise fetch from server
    fetchMediaSmart({ limit: 20 });
  }, []);

  const cacheInfo = getCacheInfo('list');

  return (
    <div>
      {cacheInfo?.isValid && (
        <div className="cache-indicator">
          ✓ Cached data (expires in {Math.round(cacheInfo.timeUntilExpiry / 1000)}s)
        </div>
      )}
      
      <button onClick={() => fetchMediaSmart({}, true)}>
        Force Refresh
      </button>
      
      {loading ? 'Loading...' : media.map(item => (
        <MediaItem key={item.id} media={item} />
      ))}
    </div>
  );
}
```

### Auto-refresh with Cache Awareness

```tsx
import { useAutoRefreshMedia } from '@shared/hooks/useCachedMedia';

function LiveMediaFeed() {
  const { isCached, cacheInfo } = useAutoRefreshMedia(
    { limit: 10, sortBy: 'capturedAt' },
    30000 // Check every 30 seconds
  );

  return (
    <div>
      <div className="status">
        {isCached ? '🟢 Using cached data' : '🔴 Fetching fresh data'}
        {cacheInfo && (
          <span> • Expires in {Math.round(cacheInfo.timeUntilExpiry / 1000)}s</span>
        )}
      </div>
      {/* Media content */}
    </div>
  );
}
```

### Manual Cache Management

```tsx
import { useMediaApi } from '@shared/hooks/useApi';

function MediaManager() {
  const { 
    updateMedia, 
    deleteMedia, 
    invalidateCache,
    invalidateMediaCache 
  } = useMediaApi();

  const handleUpdateMedia = async (id: string, data: UpdateMediaRequest) => {
    await updateMedia(id, data);
    // Cache is automatically invalidated after updates
  };

  const handleDeleteMedia = async (id: string) => {
    await deleteMedia(id);
    // Cache is automatically invalidated after deletion
  };

  const handleManualCacheInvalidation = () => {
    invalidateCache(); // Invalidate entire media list cache
  };



  return (
    <div>
      {/* Your component UI */}
      <button onClick={handleManualCacheInvalidation}>
        Clear All Cache
      </button>
    </div>
  );
}
```

## How It Works

### 1. Cache Metadata Storage

The Redux state now includes cache metadata:

```typescript
interface MediaState {
  // ... existing fields
  cache: CacheMetadata | null;           // Cache for media list
  currentMediaCache: CacheMetadata | null; // Cache for current media item
}
```

### 2. Enhanced Async Thunks

Async thunks now:
- Check cache validity before making requests
- Include conditional headers (`If-None-Match`, `If-Modified-Since`)
- Extract and store cache metadata from response headers
- Handle 304 Not Modified responses

### 3. Automatic Cache Validation

Cache validation happens automatically when:
- `fetchMediaSmart()` or `fetchMediaByIdSmart()` is called
- `isCacheValid()` function checks expiration inline
- Expired cache is replaced with fresh data from the server

### 4. Server Integration

The server already sets appropriate cache headers:

```go
// In media handler
c.Header("Cache-Control", "public, max-age=86400") // 24 hours
```

## Cache Headers Supported

- **Cache-Control**: `max-age`, `public`, `private`, `no-cache`, `no-store`
- **ETag**: For conditional requests
- **Last-Modified**: For conditional requests

## Best Practices

1. **Use smart fetch methods**: Prefer `fetchMediaSmart()` over `fetchMedia()`
2. **Force refresh when needed**: Use `forceRefresh` parameter for user-initiated refreshes
3. **Monitor cache status**: Display cache indicators in your UI
4. **Handle offline scenarios**: Cache helps when server is temporarily unavailable
5. **Invalidate appropriately**: Cache is auto-invalidated on mutations, but you can manually invalidate if needed

## Performance Benefits

- **Reduced network requests**: Valid cached data prevents unnecessary API calls
- **Faster page loads**: Cached data is instantly available
- **Better UX**: Smooth transitions with cached data while fresh data loads in background
- **Bandwidth savings**: Conditional requests reduce data transfer when content hasn't changed

## Migration Guide

For existing components using `useMediaApi()`:

1. Replace `fetchMedia()` calls with `fetchMediaSmart()`
2. Replace `fetchMediaById()` calls with `fetchMediaByIdSmart()`
3. Add cache status indicators to your UI
4. Consider using `useAutoRefreshMedia()` for real-time data components

The existing API remains compatible, but the new methods provide better performance.
