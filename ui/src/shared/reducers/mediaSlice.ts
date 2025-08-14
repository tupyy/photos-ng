import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { Media, UpdateMediaRequest, ListMediaResponse } from '@generated/models';
import { ListMediaTypeEnum, ListMediaSortByEnum, ListMediaSortOrderEnum } from '@generated/api/media-api';
import { mediaApi } from '@api/apiConfig';
import { CacheMetadata, CachedResponse, createCacheMetadata, isCacheValid, shouldMakeConditionalRequest } from '@shared/types';

// Async thunks
export const fetchMedia = createAsyncThunk(
  'media/fetchMedia',
  async (params: {
    limit?: number;
    offset?: number;
    albumId?: string;
    type?: ListMediaTypeEnum;
    startDate?: string;
    endDate?: string;
    sortBy?: ListMediaSortByEnum;
    sortOrder?: ListMediaSortOrderEnum;
    forceRefresh?: boolean;
  } = {}, { getState }) => {
    const state = getState() as any;
    const currentCache = state.media.cache;
    
    // Check if we can use cached data (unless force refresh is requested)
    if (!params.forceRefresh && currentCache && isCacheValid(currentCache)) {
      // Return a special action to indicate cache hit
      throw new Error('CACHE_HIT');
    }
    
    // If we have expired cache, clear it automatically
    if (currentCache && !isCacheValid(currentCache)) {
      // Cache will be replaced with fresh data after the request
    }
    
    // Prepare conditional request headers if we have cache metadata
    const conditionalHeaders: Record<string, string> = {};
    if (currentCache && shouldMakeConditionalRequest(currentCache)) {
      if (currentCache.etag) {
        conditionalHeaders['If-None-Match'] = currentCache.etag;
      }
      if (currentCache.lastModified) {
        conditionalHeaders['If-Modified-Since'] = currentCache.lastModified;
      }
    }
    
    const response = await mediaApi.listMedia(
      params.limit,
      params.offset,
      params.albumId,
      params.type,
      params.startDate,
      params.endDate,
      params.sortBy,
      params.sortOrder,
      {
        headers: conditionalHeaders
      }
    );
    
    // Create cache metadata from response headers
    const cache = createCacheMetadata(response.headers || {});
    
    return {
      data: response.data,
      cache
    } as CachedResponse<ListMediaResponse>;
  }
);

export const fetchMediaById = createAsyncThunk(
  'media/fetchMediaById',
  async (params: { id: string; forceRefresh?: boolean }, { getState }) => {
    const state = getState() as any;
    const currentMedia = state.media.currentMedia;
    const currentMediaCache = state.media.currentMediaCache;
    
    // Check if we can use cached data for this specific media item
    if (!params.forceRefresh && 
        currentMedia && 
        currentMedia.id === params.id && 
        currentMediaCache && 
        isCacheValid(currentMediaCache)) {
      throw new Error('CACHE_HIT');
    }
    
    // If we have expired cache for this media, it will be replaced after the request
    
    // Prepare conditional request headers
    const conditionalHeaders: Record<string, string> = {};
    if (currentMediaCache && shouldMakeConditionalRequest(currentMediaCache)) {
      if (currentMediaCache.etag) {
        conditionalHeaders['If-None-Match'] = currentMediaCache.etag;
      }
      if (currentMediaCache.lastModified) {
        conditionalHeaders['If-Modified-Since'] = currentMediaCache.lastModified;
      }
    }
    
    const response = await mediaApi.getMedia(params.id, {
      headers: conditionalHeaders
    });
    
    const cache = createCacheMetadata(response.headers || {});
    
    return {
      data: response.data,
      cache
    } as CachedResponse<Media>;
  }
);

export const updateMedia = createAsyncThunk(
  'media/updateMedia',
  async ({ id, mediaData }: { id: string; mediaData: UpdateMediaRequest }) => {
    const response = await mediaApi.updateMedia(id, mediaData);
    return response.data;
  }
);

export const deleteMedia = createAsyncThunk(
  'media/deleteMedia',
  async (id: string) => {
    await mediaApi.deleteMedia(id);
    return id;
  }
);

// Media filters interface
export interface MediaFilters {
  albumId?: string;
  type?: ListMediaTypeEnum;
  startDate?: string;
  endDate?: string;
  sortBy?: ListMediaSortByEnum;
  sortOrder?: ListMediaSortOrderEnum;
  limit: number;
  offset: number;
}

// State interface
interface MediaState {
  media: Media[];
  currentMedia: Media | null;
  total: number;
  filters: MediaFilters;
  loading: boolean;
  loadingMore: boolean; // For infinite scroll loading indicator
  error: string | null;
  selectedMediaIds: string[];
  viewMode: 'grid' | 'list';
  hasMore: boolean; // Track if there are more items to load
  // Cache metadata for the media list
  cache: CacheMetadata | null;
  // Cache metadata for the current media item
  currentMediaCache: CacheMetadata | null;
}

// Initial state
const initialState: MediaState = {
  media: [],
  currentMedia: null,
  total: 0,
  filters: {
    limit: 50,
    offset: 0,
    sortBy: ListMediaSortByEnum.CapturedAt,
    sortOrder: ListMediaSortOrderEnum.Desc,
  },
  loading: false,
  loadingMore: false,
  error: null,
  selectedMediaIds: [],
  viewMode: 'grid',
  hasMore: true,
  cache: null,
  currentMediaCache: null,
};

// Slice
const mediaSlice = createSlice({
  name: 'media',
  initialState,
  reducers: {
    clearCurrentMedia: (state) => {
      state.currentMedia = null;
    },
    clearError: (state) => {
      state.error = null;
    },
    setFilters: (state, action: PayloadAction<Partial<MediaFilters>>) => {
      state.filters = { ...state.filters, ...action.payload };
      // Reset to first page when filters change
      if (action.payload.albumId !== undefined || 
          action.payload.type !== undefined ||
          action.payload.startDate !== undefined ||
          action.payload.endDate !== undefined ||
          action.payload.sortBy !== undefined ||
          action.payload.sortOrder !== undefined) {
        state.filters.offset = 0;
        state.hasMore = true;
      }
    },
    clearFilters: (state) => {
      state.filters = {
        limit: 50,
        offset: 0,
        sortBy: ListMediaSortByEnum.CapturedAt,
        sortOrder: ListMediaSortOrderEnum.Desc,
      };
      state.hasMore = true;
    },
    toggleMediaSelection: (state, action: PayloadAction<string>) => {
      const mediaId = action.payload;
      const index = state.selectedMediaIds.indexOf(mediaId);
      if (index === -1) {
        state.selectedMediaIds.push(mediaId);
      } else {
        state.selectedMediaIds.splice(index, 1);
      }
    },
    selectAllMedia: (state) => {
      state.selectedMediaIds = state.media.map(media => media.id);
    },
    clearSelection: (state) => {
      state.selectedMediaIds = [];
    },
    setViewMode: (state, action: PayloadAction<'grid' | 'list'>) => {
      state.viewMode = action.payload;
    },
    invalidateCache: (state) => {
      state.cache = null;
      state.currentMediaCache = null;
    },
    loadNextPage: (state) => {
      if (!state.loadingMore && state.hasMore) {
        state.filters.offset = state.media.length;
      }
    },
  },
  extraReducers: (builder) => {
    // Fetch media
    builder
      .addCase(fetchMedia.pending, (state, action) => {
        // For offset > 0, this is loading more data (infinite scroll)
        const isLoadingMore = action.meta.arg?.offset && action.meta.arg.offset > 0;
        if (isLoadingMore) {
          state.loadingMore = true;
        } else {
          state.loading = true;
        }
        state.error = null;
      })
      .addCase(fetchMedia.fulfilled, (state, action: PayloadAction<CachedResponse<ListMediaResponse>>) => {
        state.loading = false;
        state.loadingMore = false;
        
        // Update cache metadata
        state.cache = action.payload.cache;
        
        // Update basic response data first
        state.total = action.payload.data.total;
        state.filters.limit = action.payload.data.limit;
        state.filters.offset = action.payload.data.offset;
        
        // Determine if this is a fresh load or infinite scroll based on the request offset
        const requestOffset = action.meta.arg?.offset || 0;
        
        if (requestOffset === 0) {
          // Fresh load - replace the media array
          state.media = action.payload.data.media;
        } else {
          // Infinite scroll - append new media items
          const existingIds = new Set(state.media.map(m => m.id));
          const newMedia = action.payload.data.media.filter(m => !existingIds.has(m.id));
          state.media.push(...newMedia);
        }
        
        // If we got no items, we've reached the end
        state.hasMore = action.payload.data.media.length > 0;
        
        // Debug logging (dev only)
        if (process.env.NODE_ENV === 'development') {
          console.log('📥 API Response:', {
            requestParams: action.meta.arg,
            receivedItems: action.payload.data.media.length,
            total: action.payload.data.total,
            limit: action.payload.data.limit,
            offset: action.payload.data.offset,
            hasMore: state.hasMore,
          });
        }
      })
      .addCase(fetchMedia.rejected, (state, action) => {
        // Handle cache hit scenario
        if (action.error.message === 'CACHE_HIT') {
          state.loading = false;
          state.loadingMore = false;
          return; // Keep existing data and cache
        }
        
        state.loading = false;
        state.loadingMore = false;
        state.error = action.error.message || 'Failed to fetch media';
      });

    // Fetch media by ID
    builder
      .addCase(fetchMediaById.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchMediaById.fulfilled, (state, action: PayloadAction<CachedResponse<Media>>) => {
        state.loading = false;
        state.currentMedia = action.payload.data;
        state.currentMediaCache = action.payload.cache;
        
        // Also update the media in the main list if it exists
        const index = state.media.findIndex(media => media.id === action.payload.data.id);
        if (index !== -1) {
          state.media[index] = action.payload.data;
        }
      })
      .addCase(fetchMediaById.rejected, (state, action) => {
        // Handle cache hit scenario
        if (action.error.message === 'CACHE_HIT') {
          state.loading = false;
          return; // Keep existing data and cache
        }
        
        state.loading = false;
        state.error = action.error.message || 'Failed to fetch media';
      });

    // Update media
    builder
      .addCase(updateMedia.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(updateMedia.fulfilled, (state, action: PayloadAction<Media>) => {
        state.loading = false;
        const index = state.media.findIndex(media => media.id === action.payload.id);
        if (index !== -1) {
          state.media[index] = action.payload;
        }
        if (state.currentMedia && state.currentMedia.id === action.payload.id) {
          state.currentMedia = action.payload;
          // Invalidate cache for updated media
          state.currentMediaCache = null;
        }
        // Invalidate the main cache since data has changed
        state.cache = null;
      })
      .addCase(updateMedia.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Failed to update media';
      });

    // Delete media
    builder
      .addCase(deleteMedia.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(deleteMedia.fulfilled, (state, action: PayloadAction<string>) => {
        state.loading = false;
        state.media = state.media.filter(media => media.id !== action.payload);
        state.total -= 1;
        state.selectedMediaIds = state.selectedMediaIds.filter(id => id !== action.payload);
        if (state.currentMedia && state.currentMedia.id === action.payload) {
          state.currentMedia = null;
          state.currentMediaCache = null;
        }
        // Invalidate cache since data has changed
        state.cache = null;
      })
      .addCase(deleteMedia.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Failed to delete media';
      });
  },
});

export const {
  clearCurrentMedia,
  clearError,
  setFilters,
  clearFilters,
  toggleMediaSelection,
  selectAllMedia,
  clearSelection,
  setViewMode,
  invalidateCache,
  loadNextPage,
} = mediaSlice.actions;

export default mediaSlice.reducer;