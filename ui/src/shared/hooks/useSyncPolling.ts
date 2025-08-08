import { useEffect } from 'react';
import { useAppDispatch, useAppSelector, selectSyncJobs } from '@shared/store';
import { fetchSyncJobs } from '@shared/reducers/syncSlice';

/**
 * Custom hook that manages sync job polling logic.
 * This centralizes all polling behavior and automatically starts/stops
 * based on whether there are active jobs.
 */
export const useSyncPolling = (options: {
  enabled?: boolean;
  interval?: number;
} = {}) => {
  const { enabled = true, interval = 1000 } = options;
  const dispatch = useAppDispatch();
  const jobs = useAppSelector(selectSyncJobs);

  // Check if there are any active jobs (running status)
  const hasActiveJobs = jobs.some(job => job.status === 'running');

  useEffect(() => {
    if (!enabled) return;

    // Initial fetch when hook is enabled (not silent for first load)
    dispatch(fetchSyncJobs());
  }, [enabled, dispatch]);

  useEffect(() => {
    if (!enabled || !hasActiveJobs) {
      // Don't poll if disabled or no active jobs
      return;
    }

    // Start polling while there are active jobs (silent to prevent flickering)
    const pollInterval = setInterval(() => {
      dispatch(fetchSyncJobs({ silent: true }));
    }, interval);

    return () => clearInterval(pollInterval);
  }, [enabled, hasActiveJobs, interval, dispatch]);

  return {
    jobs,
    hasActiveJobs,
    isPolling: enabled && hasActiveJobs,
  };
};
