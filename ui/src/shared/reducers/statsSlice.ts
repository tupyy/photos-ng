import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { StatsApi } from '@generated/api/stats-api';
import { StatsResponse } from '@generated/models/stats-response';
import { apiConfig } from '@shared/api/apiConfig';

// Initialize API instance
const statsApi = new StatsApi(undefined, apiConfig.basePath);

// Async thunks
export const fetchStats = createAsyncThunk(
  'stats/fetchStats',
  async (_, { rejectWithValue }) => {
    try {
      const response = await statsApi.getStats();
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || 'Failed to fetch statistics');
    }
  }
);

// State interface
interface StatsState {
  data: StatsResponse | null;
  loading: boolean;
  error: string | null;
}

// Initial state
const initialState: StatsState = {
  data: null,
  loading: false,
  error: null,
};

// Slice
const statsSlice = createSlice({
  name: 'stats',
  initialState,
  reducers: {
    clearError: (state) => {
      state.error = null;
    },
    resetStats: (state) => {
      state.data = null;
      state.loading = false;
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    // Fetch stats
    builder
      .addCase(fetchStats.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchStats.fulfilled, (state, action: PayloadAction<StatsResponse>) => {
        state.loading = false;
        state.data = action.payload;
      })
      .addCase(fetchStats.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string || 'Failed to fetch statistics';
      });
  },
});

export const {
  clearError,
  resetStats,
} = statsSlice.actions;

export default statsSlice.reducer;
