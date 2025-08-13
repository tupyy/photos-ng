import { useEffect } from 'react';
import { useAppDispatch, useAppSelector, selectSyncJobs } from '@shared/store';
import { fetchSyncJob } from '@shared/reducers/syncSlice';

/**
 * Custom hook for managing individual sync job details and polling.
 * This centralizes the logic for fetching and monitoring a specific sync job.
 */
export const useSyncJobDetail = (jobId: string, options: {
  interval?: number;
} = {}) => {
  const { interval = 500 } = options;
  const dispatch = useAppDispatch();
  const jobs = useAppSelector(selectSyncJobs);
  
  // Find the specific job in the Redux state
  const job = jobs.find(j => j.id === jobId);
  const isActive = job ? job.status === 'running' : false;
  // Always poll unless the job is completed, failed, or stopped
  const shouldPoll = job ? !['completed', 'failed', 'stopped'].includes(job.status) : true;

  useEffect(() => {
    // Initial fetch when hook is mounted (not silent for first load)
    console.log('useSyncJobDetail: Initial fetch for job', jobId);
    dispatch(fetchSyncJob({ jobId }));
  }, [dispatch, jobId]);

  useEffect(() => {
    // Start polling immediately and continue until job is finished
    console.log('useSyncJobDetail: Starting polling for job', jobId, 'status:', job?.status);
    
    const pollInterval = setInterval(() => {
      dispatch(fetchSyncJob({ jobId, silent: true }));
    }, interval);

    return () => {
      console.log('useSyncJobDetail: Cleanup polling for job', jobId);
      clearInterval(pollInterval);
    };
  }, [dispatch, jobId, interval]); // Don't depend on job status - just keep polling

  return {
    job,
    isActive,
    isPolling: shouldPoll,
  };
};
