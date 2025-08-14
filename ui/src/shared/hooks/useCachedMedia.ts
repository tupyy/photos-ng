import { useCallback, useEffect } from 'react';
import { useMediaApi } from './useApi';
import { useAppSelector } from '@shared/store';
import { isCacheValid } from '@shared/types';
import { MediaFilters } from '@reducers/mediaSlice';

/**
 * Enhanced hook for media operations with intelligent caching
 */
export const useCachedMedia = () => {
  const mediaApi = useMediaApi();
  const cache = useAppSelector(state => state.media.cache);
  const currentMediaCache = useAppSelector(state => state.media.currentMediaCache);

  // Smart fetch media that respects cache
  const fetchMediaSmart = useCallback(
    (params?: Partial<MediaFilters>, forceRefresh = false) => {
      // If we have valid cache and not forcing refresh, don't fetch
      if (!forceRefresh && cache && isCacheValid(cache)) {
        return Promise.resolve();
      }
      
      return mediaApi.fetchMedia({ ...params, forceRefresh });
    },
    [mediaApi, cache]
  );

  // Smart fetch media by ID that respects cache
  const fetchMediaByIdSmart = useCallback(
    (id: string, forceRefresh = false) => {
      // If we have valid cache for this specific media and not forcing refresh, don't fetch
      if (!forceRefresh && 
          mediaApi.currentMedia?.id === id && 
          currentMediaCache && 
          isCacheValid(currentMediaCache)) {
        return Promise.resolve();
      }
      
      return mediaApi.fetchMediaById(id, forceRefresh);
    },
    [mediaApi, currentMediaCache]
  );

  // Force refresh function that bypasses cache
  const forceRefreshMedia = useCallback(
    (params?: Partial<MediaFilters>) => {
      return fetchMediaSmart(params, true);
    },
    [fetchMediaSmart]
  );

  // Force refresh specific media
  const forceRefreshMediaById = useCallback(
    (id: string) => {
      return fetchMediaByIdSmart(id, true);
    },
    [fetchMediaByIdSmart]
  );

  // Check if data is cached and valid
  const isCached = useCallback(
    (type: 'list' | 'current') => {
      if (type === 'list') {
        return cache ? isCacheValid(cache) : false;
      } else {
        return currentMediaCache ? isCacheValid(currentMediaCache) : false;
      }
    },
    [cache, currentMediaCache]
  );

  // Get cache expiration info
  const getCacheInfo = useCallback(
    (type: 'list' | 'current') => {
      const targetCache = type === 'list' ? cache : currentMediaCache;
      if (!targetCache) return null;

      const now = Date.now();
      const expiresAt = targetCache.expiresAt || (targetCache.fetchedAt + (5 * 60 * 1000)); // default 5 min
      const timeUntilExpiry = expiresAt - now;
      
      return {
        isValid: isCacheValid(targetCache),
        expiresAt,
        timeUntilExpiry: Math.max(0, timeUntilExpiry),
        fetchedAt: targetCache.fetchedAt,
        maxAge: targetCache.maxAge,
      };
    },
    [cache, currentMediaCache]
  );

  return {
    // Original API methods
    ...mediaApi,
    
    // Enhanced methods with caching
    fetchMediaSmart,
    fetchMediaByIdSmart,
    forceRefreshMedia,
    forceRefreshMediaById,
    
    // Cache utilities
    isCached,
    getCacheInfo,
    
    // Cache invalidation (from original API)
    invalidateCache: mediaApi.invalidateCache,
  };
};

/**
 * Hook for components that need to auto-refresh data when cache expires
 */
export const useAutoRefreshMedia = (
  params?: Partial<MediaFilters>,
  refreshInterval?: number
) => {
  const { fetchMediaSmart, getCacheInfo, isCached } = useCachedMedia();

  useEffect(() => {
    // Initial fetch
    fetchMediaSmart(params);

    // Set up auto-refresh if specified
    if (refreshInterval) {
      const interval = setInterval(() => {
        const cacheInfo = getCacheInfo('list');
        
        // Only refresh if cache is invalid or about to expire (within 1 minute)
        if (!cacheInfo?.isValid || (cacheInfo.timeUntilExpiry < 60000)) {
          fetchMediaSmart(params, true);
        }
      }, refreshInterval);

      return () => clearInterval(interval);
    }
  }, [fetchMediaSmart, params, refreshInterval, getCacheInfo]);

  return {
    isCached: isCached('list'),
    cacheInfo: getCacheInfo('list'),
  };
};
