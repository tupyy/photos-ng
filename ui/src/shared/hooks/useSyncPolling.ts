import { useEffect, useRef } from 'react';
import { useAppDispatch, useAppSelector, selectSyncJobs } from '@shared/store';
import { fetchSyncJobs } from '@shared/reducers/syncSlice';

/**
 * Custom hook that manages sync job polling logic.
 * This centralizes polling behavior and automatically starts/stops
 * based on whether there are active jobs. Does not perform initial fetching.
 * Uses exponential backoff polling when no active jobs (2s to 1min).
 */
export const useSyncPolling = (options: {
  enabled?: boolean;
  interval?: number;
  resetExponentialPolling?: boolean;
} = {}) => {
  const { enabled = true, interval = 1000, resetExponentialPolling = false } = options;
  const dispatch = useAppDispatch();
  const jobs = useAppSelector(selectSyncJobs);
  
  // Exponential polling state
  const exponentialIntervalRef = useRef(2000); // Start at 2 seconds
  const maxExponentialInterval = 60000; // Max 1 minute

  // Check if there are any active jobs (pending or running status)
  const hasActiveJobs = jobs.some(job => job.status === 'running' || job.status === 'pending');

  // Reset exponential polling interval when requested
  useEffect(() => {
    if (resetExponentialPolling) {
      exponentialIntervalRef.current = 2000;
    }
  }, [resetExponentialPolling]);


  // Polling effect for active jobs (constant interval)
  useEffect(() => {
    if (!enabled || !hasActiveJobs) {
      return;
    }

    // Reset exponential interval when active jobs are detected
    exponentialIntervalRef.current = 2000;

    // Start polling while there are active jobs (silent to prevent flickering)
    const pollInterval = setInterval(() => {
      dispatch(fetchSyncJobs({ silent: true }));
    }, interval);

    return () => clearInterval(pollInterval);
  }, [enabled, hasActiveJobs, interval, dispatch]);

  // Exponential polling effect when no active jobs
  useEffect(() => {
    if (!enabled || hasActiveJobs) {
      return;
    }

    const startExponentialPolling = () => {
      const pollInterval = setTimeout(() => {
        dispatch(fetchSyncJobs({ silent: true }));
        
        // Increase interval for next poll (exponential backoff)
        exponentialIntervalRef.current = Math.min(
          exponentialIntervalRef.current * 2,
          maxExponentialInterval
        );
        
        // Schedule next poll
        startExponentialPolling();
      }, exponentialIntervalRef.current);

      return pollInterval;
    };

    const timeoutId = startExponentialPolling();
    
    return () => clearTimeout(timeoutId);
  }, [enabled, hasActiveJobs, dispatch]);

  return {
    jobs,
    hasActiveJobs,
    isPolling: enabled && (hasActiveJobs || jobs.length > 0),
    currentPollingInterval: hasActiveJobs ? interval : exponentialIntervalRef.current,
  };
};
