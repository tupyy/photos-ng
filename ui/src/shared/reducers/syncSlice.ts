import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { SyncApi } from '@generated/api/sync-api';
import { SyncJob, StartSyncRequest, StopSyncJob200Response, StopAllSyncJobs200Response } from '@generated/models';
import { apiConfig } from '@shared/api/apiConfig';

export interface SyncState {
  jobs: SyncJob[];
  loading: boolean;
  error: string | null;
  startingSync: boolean;
}

const initialState: SyncState = {
  jobs: [],
  loading: false,
  error: null,
  startingSync: false,
};

const syncApi = new SyncApi(apiConfig);

// Async thunk to start a new sync job
export const startSyncJob = createAsyncThunk(
  'sync/startJob',
  async (path: string, { rejectWithValue }) => {
    try {
      const request: StartSyncRequest = { path };
      const response = await syncApi.startSyncJob(request);
      return {
        jobId: response.data.id,
        path,
      };
    } catch (error: any) {
      return rejectWithValue(
        error.response?.data?.message || 'Failed to start sync job'
      );
    }
  }
);

// Async thunk to fetch all sync jobs
export const fetchSyncJobs = createAsyncThunk(
  'sync/fetchJobs',
  async (options: { silent?: boolean } = {}, { rejectWithValue }) => {
    try {
      const response = await syncApi.listSyncJobs();
      return { 
        jobs: response.data.jobs || [], 
        silent: options.silent || false 
      };
    } catch (error: any) {
      return rejectWithValue(
        error.response?.data?.message || 'Failed to fetch sync jobs'
      );
    }
  }
);

// Async thunk to fetch a specific sync job
export const fetchSyncJob = createAsyncThunk(
  'sync/fetchJob',
  async (params: { jobId: string; silent?: boolean }, { rejectWithValue }) => {
    try {
      const response = await syncApi.getSyncJob(params.jobId);
      return { 
        job: response.data, 
        silent: params.silent || false 
      };
    } catch (error: any) {
      return rejectWithValue(
        error.response?.data?.message || 'Failed to fetch sync job'
      );
    }
  }
);

// Async thunk to stop a specific sync job
export const stopSyncJob = createAsyncThunk(
  'sync/stopJob',
  async (jobId: string, { rejectWithValue }) => {
    try {
      const response = await syncApi.stopSyncJob(jobId);
      return {
        jobId,
        message: response.data.message || 'Sync job stopped successfully',
      };
    } catch (error: any) {
      return rejectWithValue(
        error.response?.data?.message || 'Failed to stop sync job'
      );
    }
  }
);

// Async thunk to stop all sync jobs
export const stopAllSyncJobs = createAsyncThunk(
  'sync/stopAllJobs',
  async (_, { rejectWithValue }) => {
    try {
      const response = await syncApi.stopAllSyncJobs();
      return {
        message: response.data.message || 'All sync jobs stopped successfully',
        stoppedCount: response.data.stoppedCount || 0,
      };
    } catch (error: any) {
      return rejectWithValue(
        error.response?.data?.message || 'Failed to stop sync jobs'
      );
    }
  }
);

const syncSlice = createSlice({
  name: 'sync',
  initialState,
  reducers: {
    clearError: (state) => {
      state.error = null;
    },
    updateJob: (state, action: PayloadAction<SyncJob>) => {
      const index = state.jobs.findIndex(job => job.id === action.payload.id);
      if (index !== -1) {
        state.jobs[index] = action.payload;
      } else {
        state.jobs.push(action.payload);
      }
    },
  },
  extraReducers: (builder) => {
    builder
      // Start sync job
      .addCase(startSyncJob.pending, (state) => {
        state.startingSync = true;
        state.error = null;
      })
      .addCase(startSyncJob.fulfilled, (state) => {
        state.startingSync = false;
        state.error = null;
      })
      .addCase(startSyncJob.rejected, (state, action) => {
        state.startingSync = false;
        state.error = action.payload as string;
      })
      
      // Fetch sync jobs
      .addCase(fetchSyncJobs.pending, (state, action) => {
        // Only show loading state if not a silent polling request
        if (!action.meta.arg?.silent) {
          state.loading = true;
        }
        state.error = null;
      })
      .addCase(fetchSyncJobs.fulfilled, (state, action) => {
        state.loading = false;
        state.jobs = action.payload.jobs;
        state.error = null;
      })
      .addCase(fetchSyncJobs.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload as string;
      })
      
      // Fetch single sync job
      .addCase(fetchSyncJob.fulfilled, (state, action) => {
        const index = state.jobs.findIndex(job => job.id === action.payload.job.id);
        if (index !== -1) {
          state.jobs[index] = action.payload.job;
        } else {
          state.jobs.push(action.payload.job);
        }
      })
      
      // Stop sync job
      .addCase(stopSyncJob.fulfilled, (state, action) => {
        // Update the job status in the list (backend will return updated job data via polling)
        // Don't remove the job, just clear any errors
        state.error = null;
      })
      .addCase(stopSyncJob.rejected, (state, action) => {
        state.error = action.payload as string;
      })
      
      // Stop all sync jobs
      .addCase(stopAllSyncJobs.fulfilled, (state, action) => {
        // Jobs will be updated via polling with new status
        // Don't remove jobs, just clear errors
        state.error = null;
      })
      .addCase(stopAllSyncJobs.rejected, (state, action) => {
        state.error = action.payload as string;
      });
  },
});

export const { clearError, updateJob } = syncSlice.actions;
export default syncSlice.reducer;