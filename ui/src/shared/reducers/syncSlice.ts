import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { SyncApi } from '@generated/api/sync-api';
import { SyncJob, StartSyncRequest, StopSyncJob200Response, ClearFinishedSyncJobsResponse, SyncJobActionRequestActionEnum } from '@generated/models';
import { apiConfig } from '@shared/api/apiConfig';

// Extended SyncJob interface to include path information
interface SyncJobWithPath extends Omit<SyncJob, 'path'> {
  path?: string;
}

export interface SyncState {
  jobs: SyncJobWithPath[];
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

// Async thunk to start a new sync job (now creates multiple jobs)
export const startSyncJob = createAsyncThunk(
  'sync/startJob',
  async (path: string, { rejectWithValue }) => {
    try {
      const request: StartSyncRequest = { path };
      const response = await syncApi.startSyncJob(request);
      const albumPath = response.data.id; // Now returns album path instead of single job ID
      
      // Fetch all jobs to get the newly created jobs for this sync operation
      const allJobsResponse = await syncApi.listSyncJobs();
      const allJobs = allJobsResponse.data.jobs || [];
      
      // Filter jobs that are related to this album path (jobs for folders within this path)
      const relatedJobs = allJobs.filter(job => 
        job.path === path || job.path?.startsWith(path + '/')
      );
      
      return {
        albumPath,
        path,
        jobs: relatedJobs,
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

// Async thunk to cancel a specific sync job
export const cancelSyncJob = createAsyncThunk(
  'sync/cancelJob',
  async (jobId: string, { rejectWithValue }) => {
    try {
      const response = await syncApi.actionSyncJob(jobId, { action: SyncJobActionRequestActionEnum.Cancel });
      return {
        jobId,
        message: response.data.message || 'Sync job cancelled successfully',
      };
    } catch (error: any) {
      return rejectWithValue(
        error.response?.data?.message || 'Failed to cancel sync job'
      );
    }
  }
);

// Async thunk to pause/resume a specific sync job
export const pauseSyncJob = createAsyncThunk(
  'sync/pauseJob',
  async (jobId: string, { rejectWithValue }) => {
    try {
      const response = await syncApi.actionSyncJob(jobId, { action: SyncJobActionRequestActionEnum.Pause });
      return {
        jobId,
        message: response.data.message || 'Sync job pause/resume toggled successfully',
      };
    } catch (error: any) {
      return rejectWithValue(
        error.response?.data?.message || 'Failed to pause/resume sync job'
      );
    }
  }
);

// Async thunk to cancel all sync jobs
export const cancelAllSyncJobs = createAsyncThunk(
  'sync/cancelAllJobs',
  async (_, { rejectWithValue }) => {
    try {
      const response = await syncApi.actionAllSyncJobs({ action: SyncJobActionRequestActionEnum.Cancel });
      return {
        message: response.data.message || 'All sync jobs cancelled successfully',
        cancelledCount: response.data.affectedCount || 0,
      };
    } catch (error: any) {
      return rejectWithValue(
        error.response?.data?.message || 'Failed to cancel sync jobs'
      );
    }
  }
);

// Async thunk to clear finished sync jobs
export const clearFinishedSyncJobs = createAsyncThunk(
  'sync/clearFinishedJobs',
  async (_, { rejectWithValue }) => {
    try {
      const response = await syncApi.clearFinishedSyncJobs();
      return {
        message: response.data.message || 'Finished sync jobs cleared successfully',
        clearedCount: response.data.clearedCount || 0,
      };
    } catch (error: any) {
      return rejectWithValue(
        error.response?.data?.message || 'Failed to clear finished sync jobs'
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
      .addCase(startSyncJob.fulfilled, (state, action) => {
        state.startingSync = false;
        state.error = null;
        // Add all newly created jobs to the jobs list with path information
        const newJobs = action.payload.jobs.map(job => ({
          ...job,
          path: action.payload.path,
        }));
        
        // Remove any existing jobs for this path to avoid duplicates
        state.jobs = state.jobs.filter(job => 
          job.path !== action.payload.path && 
          !job.path?.startsWith(action.payload.path + '/')
        );
        
        // Add the new jobs
        state.jobs.push(...newJobs);
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
        // Debug logging
        if (!action.payload.silent) {
          console.log('Sync jobs updated:', action.payload.jobs);
        }
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
        // Debug logging for individual job updates
        if (!action.payload.silent) {
          console.log('Individual sync job updated:', action.payload.job);
        } else {
          console.log('Silent job update:', action.payload.job.id, action.payload.job.status, `${action.payload.job.totalTasks - action.payload.job.remainingTasks}/${action.payload.job.totalTasks} tasks`);
        }
      })
      
      // Cancel sync job
      .addCase(cancelSyncJob.fulfilled, (state, action) => {
        // Update the job status in the list (backend will return updated job data via polling)
        // Don't remove the job, just clear any errors
        state.error = null;
      })
      .addCase(cancelSyncJob.rejected, (state, action) => {
        state.error = action.payload as string;
      })
      
      // Pause sync job
      .addCase(pauseSyncJob.fulfilled, (state, action) => {
        // Update the job status in the list (backend will return updated job data via polling)
        // Don't remove the job, just clear any errors
        state.error = null;
      })
      .addCase(pauseSyncJob.rejected, (state, action) => {
        state.error = action.payload as string;
      })
      
      // Cancel all sync jobs
      .addCase(cancelAllSyncJobs.fulfilled, (state, action) => {
        // Jobs will be updated via polling with new status
        // Don't remove jobs, just clear errors
        state.error = null;
      })
      .addCase(cancelAllSyncJobs.rejected, (state, action) => {
        state.error = action.payload as string;
      })
      
      // Clear finished sync jobs
      .addCase(clearFinishedSyncJobs.fulfilled, (state, action) => {
        // Remove completed, stopped, and failed jobs from the list
        state.jobs = state.jobs.filter(job => 
          job.status !== 'completed' && 
          job.status !== 'stopped' && 
          job.status !== 'failed'
        );
        state.error = null;
      })
      .addCase(clearFinishedSyncJobs.rejected, (state, action) => {
        state.error = action.payload as string;
      });
  },
});

export const { clearError, updateJob } = syncSlice.actions;

// Backward compatibility exports
export const stopSyncJob = cancelSyncJob;
export const stopAllSyncJobs = cancelAllSyncJobs;

// New cleaner exports
export const stopJob = cancelSyncJob;
export const pauseJob = pauseSyncJob;
export const stopAllJobs = cancelAllSyncJobs;

export default syncSlice.reducer;