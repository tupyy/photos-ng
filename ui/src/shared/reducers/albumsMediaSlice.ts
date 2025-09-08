import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { Media, ListMediaResponse } from '@generated/models';
import { ListMediaTypeEnum, ListMediaSortByEnum, ListMediaSortOrderEnum, ListMediaDirectionEnum } from '@generated/api/media-api';
import { mediaApi } from '@api/apiConfig';
import { CacheMetadata, CachedResponse, createCacheMetadata, isCacheValid, shouldMakeConditionalRequest } from '@shared/types';

// Albums-specific async thunk for fetching media filtered by album
export const fetchAlbumsMedia = createAsyncThunk(
  'albumsMedia/fetchMedia',
  async (params: {
    limit?: number;
    cursor?: string;
    direction?: ListMediaDirectionEnum;
    albumId: string; // REQUIRED for albums media
    type?: ListMediaTypeEnum;
    startDate?: string;
    endDate?: string;
    sortBy?: ListMediaSortByEnum;
    sortOrder?: ListMediaSortOrderEnum;
    forceRefresh?: boolean;
  }, { getState }) => {
    const state = getState() as any;
    const currentCache = state.albumsMedia.cache;

    // Cache disabled - always make fresh requests
    // if (!params.forceRefresh && currentCache && isCacheValid(currentCache)) {
    //   throw new Error('CACHE_HIT');
    // }


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

    // Albums media ALWAYS includes albumId
    const response = await mediaApi.listMedia(
      params.limit,
      params.cursor,
      params.direction,
      params.albumId, // Required for albums
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

// Albums media filters interface (albumId required)
export interface AlbumsMediaFilters {
  albumId: string; // Always required
  type?: ListMediaTypeEnum;
  startDate?: string;
  endDate?: string;
  sortBy?: ListMediaSortByEnum;
  sortOrder?: ListMediaSortOrderEnum;
  limit: number;
  cursor?: string;
  direction?: ListMediaDirectionEnum;
}

// Albums media state interface
interface AlbumsMediaState {
  media: Media[];
  currentAlbumId: string | null; // Track which album we're viewing
  filters: Partial<AlbumsMediaFilters>;
  loading: boolean;
  loadingMore: boolean;
  error: string | null;
  selectedMediaIds: string[];
  hasMore: boolean;
  nextCursor?: string | null;
  cache: CacheMetadata | null;
}

// Initial state
const initialState: AlbumsMediaState = {
  media: [],
  currentAlbumId: null,
  filters: {
    limit: 100,
    sortBy: ListMediaSortByEnum.CapturedAt,
    sortOrder: ListMediaSortOrderEnum.Desc,
  },
  loading: false,
  loadingMore: false,
  error: null,
  selectedMediaIds: [],
  hasMore: true,
  nextCursor: null,
  cache: null,
};

// Albums media slice
const albumsMediaSlice = createSlice({
  name: 'albumsMedia',
  initialState,
  reducers: {
    clearError: (state) => {
      state.error = null;
    },
    setCurrentAlbum: (state, action: PayloadAction<string | null>) => {
      // When album changes, clear everything
      if (state.currentAlbumId !== action.payload) {
        state.currentAlbumId = action.payload;
        state.media = [];
        state.selectedMediaIds = [];
        state.hasMore = true;
        state.nextCursor = null;
        state.cache = null;
        if (action.payload) {
          state.filters.albumId = action.payload;
        }
      }
    },
    setFilters: (state, action: PayloadAction<Partial<AlbumsMediaFilters>>) => {
      state.filters = { ...state.filters, ...action.payload };
      // Reset pagination when filters change
      if (action.payload.type !== undefined ||
          action.payload.startDate !== undefined ||
          action.payload.endDate !== undefined ||
          action.payload.sortBy !== undefined ||
          action.payload.sortOrder !== undefined ||
          action.payload.direction !== undefined) {
        state.filters.cursor = undefined;
        state.nextCursor = null;
        state.hasMore = true;
      }
    },
    clearFilters: (state) => {
      const currentAlbumId = state.currentAlbumId;
      state.filters = {
        limit: 100,
        sortBy: ListMediaSortByEnum.CapturedAt,
        sortOrder: ListMediaSortOrderEnum.Desc,
        albumId: currentAlbumId || undefined,
      };
      state.nextCursor = null;
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
    invalidateCache: (state) => {
      state.cache = null;
    },
  },
  extraReducers: (builder) => {
    // Fetch albums media
    builder
      .addCase(fetchAlbumsMedia.pending, (state, action) => {
        const isLoadingMore = !!action.meta.arg?.cursor;
        if (isLoadingMore) {
          state.loadingMore = true;
        } else {
          state.loading = true;
        }
        state.error = null;
      })
      .addCase(fetchAlbumsMedia.fulfilled, (state, action: PayloadAction<CachedResponse<ListMediaResponse>>) => {
        state.loading = false;
        state.loadingMore = false;

        // Update cache metadata
        state.cache = action.payload.cache;

        // Update basic response data
        state.filters.limit = action.payload.data.limit;

        // Determine if this is a fresh load or infinite scroll
        const requestCursor = action.meta.arg?.cursor;
        const forceRefresh = action.meta.arg?.forceRefresh;

        if (!requestCursor || forceRefresh) {
          // Fresh load - replace the media array
          state.media = action.payload.data.media;
        } else {
          // Infinite scroll - append or prepend based on direction
          const direction = action.meta.arg?.direction;
          const existingIds = new Set(state.media.map(m => m.id));
          const newMedia = action.payload.data.media.filter(m => !existingIds.has(m.id));
          
          if (direction === ListMediaDirectionEnum.Backward) {
            // Backward direction: prepend to beginning
            state.media.unshift(...newMedia);
          } else {
            // Forward direction: append to end
            state.media.push(...newMedia);
          }
        }

        // Compute cursor for next navigation
        if (state.media.length > 0) {
          const direction = action.meta.arg?.direction;
          const referenceMedia = direction === ListMediaDirectionEnum.Backward 
            ? state.media[0] 
            : state.media[state.media.length - 1];
            
          // Use the capturedAt as-is from the API response for cursor generation
          state.nextCursor = btoa(JSON.stringify({
            captured_at: referenceMedia.capturedAt,
            id: referenceMedia.id
          }));
        } else {
          state.nextCursor = null;
        }

        // Check if there are more pages available based on nextCursor
        state.hasMore = !!action.payload.data.nextCursor;
      })
      .addCase(fetchAlbumsMedia.rejected, (state, action) => {
        // Handle cache hit scenario
        if (action.error.message === 'CACHE_HIT') {
          state.loading = false;
          state.loadingMore = false;
          return;
        }

        state.loading = false;
        state.loadingMore = false;
        state.error = action.error.message || 'Failed to fetch albums media';
      });
  },
});

export const {
  clearError,
  setCurrentAlbum,
  setFilters,
  clearFilters,
  toggleMediaSelection,
  selectAllMedia,
  clearSelection,
  invalidateCache,
} = albumsMediaSlice.actions;

export default albumsMediaSlice.reducer;