import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { Bucket, GetTimelineResponse } from '@generated/models';
import { timelineApi } from '@api/apiConfig';

// Async thunks
export const fetchTimeline = createAsyncThunk(
  'timeline/fetchTimeline',
  async (params: {
    startDate?: string;
    endDate?: string;
    limit?: number;
    offset?: number;
  } = {}) => {
    const response = await timelineApi.getTimeline(
      params.startDate,
      params.endDate,
      params.limit,
      params.offset
    );
    return response.data;
  }
);

// Timeline filters interface
export interface TimelineFilters {
  startDate?: string;
  endDate?: string;
  limit: number;
  offset: number;
}

// State interface
interface TimelineState {
  buckets: Bucket[];
  years: number[];
  total: number;
  filters: TimelineFilters;
  loading: boolean;
  error: string | null;
  selectedYear?: number;
  selectedMonth?: number;
}

// Initial state
const initialState: TimelineState = {
  buckets: [],
  years: [],
  total: 0,
  filters: {
    limit: 20,
    offset: 0,
  },
  loading: false,
  error: null,
};

// Slice
const timelineSlice = createSlice({
  name: 'timeline',
  initialState,
  reducers: {
    clearError: (state) => {
      state.error = null;
    },
    setFilters: (state, action: PayloadAction<Partial<TimelineFilters>>) => {
      state.filters = { ...state.filters, ...action.payload };
      // Reset to first page when filters change
      if (action.payload.startDate !== undefined || action.payload.endDate !== undefined) {
        state.filters.offset = 0;
      }
    },
    clearFilters: (state) => {
      state.filters = {
        limit: 20,
        offset: 0,
      };
    },
    setSelectedYear: (state, action: PayloadAction<number | undefined>) => {
      state.selectedYear = action.payload;
      state.selectedMonth = undefined; // Clear month when year changes
    },
    setSelectedMonth: (state, action: PayloadAction<number | undefined>) => {
      state.selectedMonth = action.payload;
    },
    navigateToDate: (state, action: PayloadAction<{ year?: number; month?: number }>) => {
      if (action.payload.year !== undefined) {
        state.selectedYear = action.payload.year;
      }
      if (action.payload.month !== undefined) {
        state.selectedMonth = action.payload.month;
      }
    },
  },
  extraReducers: (builder) => {
    // Fetch timeline
    builder
      .addCase(fetchTimeline.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchTimeline.fulfilled, (state, action: PayloadAction<GetTimelineResponse>) => {
        state.loading = false;
        
        // If offset is 0, replace the buckets array; otherwise, append (for pagination)
        if (state.filters.offset === 0) {
          state.buckets = action.payload.buckets;
        } else {
          state.buckets.push(...action.payload.buckets);
        }
        
        state.years = action.payload.years;
        state.total = action.payload.total;
        state.filters.limit = action.payload.limit;
        state.filters.offset = action.payload.offset;
      })
      .addCase(fetchTimeline.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Failed to fetch timeline';
      });
  },
});

export const {
  clearError,
  setFilters,
  clearFilters,
  setSelectedYear,
  setSelectedMonth,
  navigateToDate,
} = timelineSlice.actions;

export default timelineSlice.reducer;