/**
 * Cache metadata for HTTP responses
 */
export interface CacheMetadata {
  /** When the data was fetched */
  fetchedAt: number;
  /** Cache-Control max-age in seconds */
  maxAge?: number;
  /** ETag for conditional requests */
  etag?: string;
  /** Last-Modified date */
  lastModified?: string;
  /** Whether the resource can be cached */
  cacheable: boolean;
  /** When the cache expires (computed from fetchedAt + maxAge) */
  expiresAt?: number;
}

/**
 * Enhanced response type that includes cache metadata
 */
export interface CachedResponse<T> {
  data: T;
  cache: CacheMetadata;
}

/**
 * Parse Cache-Control header and extract max-age
 */
export function parseCacheControl(cacheControl?: string): { maxAge?: number; cacheable: boolean } {
  if (!cacheControl) {
    return { cacheable: false };
  }

  const directives = cacheControl.toLowerCase().split(',').map(d => d.trim());
  
  // Check if caching is disabled
  if (directives.includes('no-cache') || directives.includes('no-store')) {
    return { cacheable: false };
  }

  // Extract max-age
  const maxAgeDirective = directives.find(d => d.startsWith('max-age='));
  if (maxAgeDirective) {
    const maxAge = parseInt(maxAgeDirective.split('=')[1], 10);
    return { maxAge, cacheable: true };
  }

  // If public or private directive exists, it's cacheable
  const hasPublic = directives.includes('public');
  const hasPrivate = directives.includes('private');
  
  return { cacheable: hasPublic || hasPrivate };
}

/**
 * Create cache metadata from response headers
 */
export function createCacheMetadata(headers: any): CacheMetadata {
  const cacheControl = (headers['cache-control'] || headers['Cache-Control']) as string;
  const etag = (headers['etag'] || headers['ETag']) as string;
  const lastModified = (headers['last-modified'] || headers['Last-Modified']) as string;
  
  const { maxAge, cacheable } = parseCacheControl(cacheControl);
  const fetchedAt = Date.now();
  
  return {
    fetchedAt,
    maxAge,
    etag,
    lastModified,
    cacheable,
    expiresAt: maxAge ? fetchedAt + (maxAge * 1000) : undefined,
  };
}

/**
 * Check if cached data is still valid
 */
export function isCacheValid(cache: CacheMetadata): boolean {
  if (!cache.cacheable) {
    return false;
  }

  if (cache.expiresAt) {
    return Date.now() < cache.expiresAt;
  }

  // If no explicit expiration, consider it valid for a default period (5 minutes)
  const defaultMaxAge = 5 * 60 * 1000; // 5 minutes in milliseconds
  return Date.now() < (cache.fetchedAt + defaultMaxAge);
}

/**
 * Check if we should make a conditional request
 */
export function shouldMakeConditionalRequest(cache: CacheMetadata): boolean {
  return !isCacheValid(cache) && (!!cache.etag || !!cache.lastModified);
}
