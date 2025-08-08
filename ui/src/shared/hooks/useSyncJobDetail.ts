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

  useEffect(() => {
    // Initial fetch when hook is mounted (not silent for first load)
    dispatch(fetchSyncJob({ jobId }));
  }, [dispatch, jobId]);

  useEffect(() => {
    if (!isActive) {
      // Job is complete or doesn't exist, stop polling
      return;
    }

    // Poll while job is active (silent to prevent flickering)
    const pollInterval = setInterval(() => {
      dispatch(fetchSyncJob({ jobId, silent: true }));
    }, interval);

    return () => clearInterval(pollInterval);
  }, [isActive, interval, dispatch, jobId]);

  return {
    job,
    isActive,
    isPolling: isActive,
  };
};
