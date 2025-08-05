import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';

export interface SyncState {
  isInProgress: boolean;
  progress: number;
  error: string | null;
}

const initialState: SyncState = {
  isInProgress: false,
  progress: 0,
  error: null,
};

// Async thunk for sync operation
export const startSync = createAsyncThunk(
  'sync/start',
  async (_, { dispatch, signal, rejectWithValue }) => {
    try {
      // TODO: Replace with actual sync API call
      // Simulating sync progress
      for (let i = 0; i <= 100; i += 10) {
        if (signal.aborted) {
          return rejectWithValue('Sync cancelled');
        }
        
        dispatch(updateProgress(i));
        await new Promise(resolve => setTimeout(resolve, 500)); // Simulate work
      }
      
      return { success: true };
    } catch (error) {
      return rejectWithValue(error instanceof Error ? error.message : 'Sync failed');
    }
  }
);

const syncSlice = createSlice({
  name: 'sync',
  initialState,
  reducers: {
    updateProgress: (state, action: PayloadAction<number>) => {
      state.progress = action.payload;
    },
    cancelSync: (state) => {
      state.isInProgress = false;
      state.progress = 0;
      state.error = null;
    },
    clearError: (state) => {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(startSync.pending, (state) => {
        state.isInProgress = true;
        state.progress = 0;
        state.error = null;
      })
      .addCase(startSync.fulfilled, (state) => {
        state.isInProgress = false;
        state.progress = 100;
        state.error = null;
      })
      .addCase(startSync.rejected, (state, action) => {
        state.isInProgress = false;
        state.progress = 0;
        state.error = action.error.message || 'Sync failed';
      });
  },
});

export const { updateProgress, cancelSync, clearError } = syncSlice.actions;
export default syncSlice.reducer;