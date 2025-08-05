import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { Album, CreateAlbumRequest, UpdateAlbumRequest, ListAlbumsResponse } from '@generated/models';
import { albumsApi } from '@api/apiConfig';

// Async thunks
export const fetchAlbums = createAsyncThunk(
  'albums/fetchAlbums',
  async (params: { limit?: number; offset?: number } = {}) => {
    const response = await albumsApi.listAlbums(params.limit, params.offset);
    return response.data;
  }
);

export const fetchAlbumById = createAsyncThunk(
  'albums/fetchAlbumById',
  async (id: string) => {
    const response = await albumsApi.getAlbum(id);
    return response.data;
  }
);

export const createAlbum = createAsyncThunk(
  'albums/createAlbum',
  async (albumData: CreateAlbumRequest) => {
    const response = await albumsApi.createAlbum(albumData);
    return response.data;
  }
);

export const updateAlbum = createAsyncThunk(
  'albums/updateAlbum',
  async ({ id, albumData }: { id: string; albumData: UpdateAlbumRequest }) => {
    const response = await albumsApi.updateAlbum(id, albumData);
    return response.data;
  }
);

export const deleteAlbum = createAsyncThunk(
  'albums/deleteAlbum',
  async (id: string) => {
    await albumsApi.deleteAlbum(id);
    return id;
  }
);

export const syncAlbum = createAsyncThunk(
  'albums/syncAlbum',
  async (id: string) => {
    const response = await albumsApi.syncAlbum(id);
    return response.data;
  }
);

// State interface
interface AlbumsState {
  albums: Album[];
  currentAlbum: Album | null;
  isPageActive: boolean;
  isCreateFormOpen: boolean;
  total: number;
  limit: number;
  offset: number;
  loading: boolean;
  error: string | null;
  syncStatus: {
    [albumId: string]: {
      syncing: boolean;
      lastSyncedItems?: number;
      error?: string;
    };
  };
}

// Initial state
const initialState: AlbumsState = {
  albums: [],
  currentAlbum: null,
  isPageActive: false,
  isCreateFormOpen: false,
  total: 0,
  limit: 20,
  offset: 0,
  loading: false,
  error: null,
  syncStatus: {},
};

// Slice
const albumsSlice = createSlice({
  name: 'albums',
  initialState,
  reducers: {
    setPageActive: (state, action: PayloadAction<boolean>) => {
      state.isPageActive = action.payload;
    },
    setCreateFormOpen: (state, action: PayloadAction<boolean>) => {
      state.isCreateFormOpen = action.payload;
    },
    setCurrentAlbum: (state, action: PayloadAction<Album | null>) => {
      state.currentAlbum = action.payload;
    },
    clearCurrentAlbum: (state) => {
      state.currentAlbum = null;
    },
    clearError: (state) => {
      state.error = null;
    },
    setFilters: (state, action: PayloadAction<{ limit?: number; offset?: number }>) => {
      if (action.payload.limit !== undefined) {
        state.limit = action.payload.limit;
      }
      if (action.payload.offset !== undefined) {
        state.offset = action.payload.offset;
      }
    },
  },
  extraReducers: (builder) => {
    // Fetch albums
    builder
      .addCase(fetchAlbums.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchAlbums.fulfilled, (state, action: PayloadAction<ListAlbumsResponse>) => {
        state.loading = false;
        state.albums = action.payload.albums;
        state.total = action.payload.total;
        state.limit = action.payload.limit;
        state.offset = action.payload.offset;
      })
      .addCase(fetchAlbums.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Failed to fetch albums';
      });

    // Fetch album by ID
    builder
      .addCase(fetchAlbumById.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchAlbumById.fulfilled, (state, action: PayloadAction<Album>) => {
        state.loading = false;
        state.currentAlbum = action.payload;
      })
      .addCase(fetchAlbumById.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Failed to fetch album';
      });

    // Create album
    builder
      .addCase(createAlbum.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(createAlbum.fulfilled, (state, action: PayloadAction<Album>) => {
        state.loading = false;
        state.albums.unshift(action.payload);
        state.total += 1;
      })
      .addCase(createAlbum.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Failed to create album';
      });

    // Update album
    builder
      .addCase(updateAlbum.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(updateAlbum.fulfilled, (state, action: PayloadAction<Album>) => {
        state.loading = false;
        const index = state.albums.findIndex(album => album.id === action.payload.id);
        if (index !== -1) {
          state.albums[index] = action.payload;
        }
        if (state.currentAlbum && state.currentAlbum.id === action.payload.id) {
          state.currentAlbum = action.payload;
        }
      })
      .addCase(updateAlbum.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Failed to update album';
      });

    // Delete album
    builder
      .addCase(deleteAlbum.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(deleteAlbum.fulfilled, (state, action: PayloadAction<string>) => {
        state.loading = false;
        state.albums = state.albums.filter(album => album.id !== action.payload);
        state.total -= 1;
        if (state.currentAlbum && state.currentAlbum.id === action.payload) {
          state.currentAlbum = null;
        }
      })
      .addCase(deleteAlbum.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Failed to delete album';
      });

    // Sync album
    builder
      .addCase(syncAlbum.pending, (state, action) => {
        const albumId = action.meta.arg;
        state.syncStatus[albumId] = {
          syncing: true,
          error: undefined,
        };
      })
      .addCase(syncAlbum.fulfilled, (state, action) => {
        const albumId = action.meta.arg;
        state.syncStatus[albumId] = {
          syncing: false,
          lastSyncedItems: action.payload.synced_items,
          error: undefined,
        };
      })
      .addCase(syncAlbum.rejected, (state, action) => {
        const albumId = action.meta.arg;
        state.syncStatus[albumId] = {
          syncing: false,
          error: action.error.message || 'Failed to sync album',
        };
      });
  },
});

export const { setPageActive, setCreateFormOpen, setCurrentAlbum, clearCurrentAlbum, clearError, setFilters } = albumsSlice.actions;
export default albumsSlice.reducer;