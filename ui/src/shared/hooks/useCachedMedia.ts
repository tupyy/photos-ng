import { useCallback, useEffect } from 'react';
import { useTimelineApi, useAlbumsMediaApi } from './useApi';
import { useAppSelector } from '@shared/store';
import { isCacheValid } from '@shared/types';
import { TimelineFilters } from '@reducers/timelineSlice';
import { AlbumsMediaFilters } from '@reducers/albumsMediaSlice';

/**
 * Enhanced hook for timeline media operations with intelligent caching
 */
export const useCachedTimelineMedia = () => {
  const timelineApi = useTimelineApi();
  const cache = useAppSelector(state => state.timeline.cache);

  // Smart fetch media that respects cache
  const fetchMediaSmart = useCallback(
    (params?: Partial<TimelineFilters>, forceRefresh = false) => {
      // If we have valid cache and not forcing refresh, don't fetch
      if (!forceRefresh && cache && isCacheValid(cache)) {
        return Promise.resolve();
      }
      
      return timelineApi.fetchMedia({ ...params, forceRefresh });
    },
    [timelineApi, cache]
  );

  // Note: Timeline doesn't support individual media fetching

  // Force refresh function that bypasses cache
  const forceRefreshMedia = useCallback(
    (params?: Partial<TimelineFilters>) => {
      return fetchMediaSmart(params, true);
    },
    [fetchMediaSmart]
  );

  // Note: Timeline doesn't support individual media fetching

  // Check if data is cached and valid
  const isCached = useCallback(
    () => {
      return cache ? isCacheValid(cache) : false;
    },
    [cache]
  );

  // Get cache expiration info
  const getCacheInfo = useCallback(
    () => {
      if (!cache) return null;

      const now = Date.now();
      const expiresAt = cache.expiresAt || (cache.fetchedAt + (5 * 60 * 1000)); // default 5 min
      const timeUntilExpiry = expiresAt - now;
      
      return {
        isValid: isCacheValid(cache),
        expiresAt,
        timeUntilExpiry: Math.max(0, timeUntilExpiry),
        fetchedAt: cache.fetchedAt,
        maxAge: cache.maxAge,
      };
    },
    [cache]
  );

  return {
    // Original API methods
    ...timelineApi,
    
    // Enhanced methods with caching
    fetchMediaSmart,
    forceRefreshMedia,
    
    // Cache utilities
    isCached,
    getCacheInfo,
    
    // Cache invalidation (from original API)
    invalidateCache: timelineApi.invalidateCache,
  };
};

/**
 * Enhanced hook for albums media operations with intelligent caching
 */
export const useCachedAlbumsMedia = () => {
  const albumsMediaApi = useAlbumsMediaApi();
  const cache = useAppSelector(state => state.albumsMedia.cache);

  // Smart fetch media that respects cache
  const fetchMediaSmart = useCallback(
    (params: Partial<AlbumsMediaFilters> & { forceRefresh?: boolean }) => {
      // If we have valid cache and not forcing refresh, don't fetch
      if (!params.forceRefresh && cache && isCacheValid(cache)) {
        return Promise.resolve();
      }
      
      return albumsMediaApi.fetchMedia(params);
    },
    [albumsMediaApi, cache]
  );

  // Force refresh function that bypasses cache
  const forceRefreshMedia = useCallback(
    (params: Partial<AlbumsMediaFilters>) => {
      return fetchMediaSmart({ ...params, forceRefresh: true });
    },
    [fetchMediaSmart]
  );

  // Check if data is cached and valid
  const isCached = useCallback(
    () => {
      return cache ? isCacheValid(cache) : false;
    },
    [cache]
  );

  // Get cache expiration info
  const getCacheInfo = useCallback(
    () => {
      if (!cache) return null;

      const now = Date.now();
      const expiresAt = cache.expiresAt || (cache.fetchedAt + (5 * 60 * 1000)); // default 5 min
      const timeUntilExpiry = expiresAt - now;
      
      return {
        isValid: isCacheValid(cache),
        expiresAt,
        timeUntilExpiry: Math.max(0, timeUntilExpiry),
        fetchedAt: cache.fetchedAt,
        maxAge: cache.maxAge,
      };
    },
    [cache]
  );

  return {
    // Original API methods
    ...albumsMediaApi,
    
    // Enhanced methods with caching
    fetchMediaSmart,
    forceRefreshMedia,
    
    // Cache utilities
    isCached,
    getCacheInfo,
    
    // Cache invalidation (from original API)
    invalidateCache: albumsMediaApi.invalidateCache,
  };
};

/**
 * Hook for components that need to auto-refresh timeline data when cache expires
 */
export const useAutoRefreshTimelineMedia = (
  params?: Partial<TimelineFilters>,
  refreshInterval?: number
) => {
  const { fetchMediaSmart, getCacheInfo, isCached } = useCachedTimelineMedia();

  useEffect(() => {
    // Initial fetch
    fetchMediaSmart(params);

    // Set up auto-refresh if specified
    if (refreshInterval) {
      const interval = setInterval(() => {
        const cacheInfo = getCacheInfo();
        
        // Only refresh if cache is invalid or about to expire (within 1 minute)
        if (!cacheInfo?.isValid || (cacheInfo.timeUntilExpiry < 60000)) {
          fetchMediaSmart(params, true);
        }
      }, refreshInterval);

      return () => clearInterval(interval);
    }
  }, [fetchMediaSmart, params, refreshInterval, getCacheInfo]);

  return {
    isCached: isCached(),
    cacheInfo: getCacheInfo(),
  };
};

// Legacy export for backward compatibility
export const useCachedMedia = useCachedTimelineMedia;
export const useAutoRefreshMedia = useAutoRefreshTimelineMedia;
