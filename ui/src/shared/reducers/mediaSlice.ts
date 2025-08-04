import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { Media, UpdateMediaRequest, ListMediaResponse } from '@generated/models';
import { ListMediaTypeEnum, ListMediaSortByEnum, ListMediaSortOrderEnum } from '@generated/api/media-api';
import { mediaApi } from '@api/apiConfig';

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
  } = {}) => {
    const response = await mediaApi.listMedia(
      params.limit,
      params.offset,
      params.albumId,
      params.type,
      params.startDate,
      params.endDate,
      params.sortBy,
      params.sortOrder
    );
    return response.data;
  }
);

export const fetchMediaById = createAsyncThunk(
  'media/fetchMediaById',
  async (id: string) => {
    const response = await mediaApi.getMedia(id);
    return response.data;
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
  error: string | null;
  selectedMediaIds: string[];
  viewMode: 'grid' | 'list';
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
  error: null,
  selectedMediaIds: [],
  viewMode: 'grid',
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
      }
    },
    clearFilters: (state) => {
      state.filters = {
        limit: 50,
        offset: 0,
        sortBy: ListMediaSortByEnum.CapturedAt,
        sortOrder: ListMediaSortOrderEnum.Desc,
      };
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
  },
  extraReducers: (builder) => {
    // Fetch media
    builder
      .addCase(fetchMedia.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchMedia.fulfilled, (state, action: PayloadAction<ListMediaResponse>) => {
        state.loading = false;
        
        // If offset is 0, replace the media array; otherwise, append (for pagination)
        if (state.filters.offset === 0) {
          state.media = action.payload.media;
        } else {
          state.media.push(...action.payload.media);
        }
        
        state.total = action.payload.total;
        state.filters.limit = action.payload.limit;
        state.filters.offset = action.payload.offset;
      })
      .addCase(fetchMedia.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Failed to fetch media';
      });

    // Fetch media by ID
    builder
      .addCase(fetchMediaById.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchMediaById.fulfilled, (state, action: PayloadAction<Media>) => {
        state.loading = false;
        state.currentMedia = action.payload;
      })
      .addCase(fetchMediaById.rejected, (state, action) => {
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
        }
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
        }
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
} = mediaSlice.actions;

export default mediaSlice.reducer;