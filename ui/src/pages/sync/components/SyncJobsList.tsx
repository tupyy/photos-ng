import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAppDispatch, useAppSelector, selectSyncJobs, selectSyncLoading, selectSyncError } from '@shared/store';
import { stopSyncJob, stopAllSyncJobs, clearFinishedSyncJobs } from '@shared/reducers/syncSlice';
import { SyncJob } from '@generated/models';

// SyncJob now includes path field from the API

export const SyncJobsList: React.FC = () => {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const jobs = useAppSelector(selectSyncJobs);
  const loading = useAppSelector(selectSyncLoading);
  const error = useAppSelector(selectSyncError);
  const [stoppingJobs, setStoppingJobs] = useState<Set<string>>(new Set());
  const [isStoppingAll, setIsStoppingAll] = useState(false);
  const [isClearingFinished, setIsClearingFinished] = useState(false);

  // This component is purely presentational - it only consumes Redux state
  // Data fetching and polling is handled by parent components

  // Clear stopping state for jobs that have finished (stopped, failed, or completed)
  useEffect(() => {
    const finishedJobIds = new Set(jobs.filter(job => job.status === 'stopped' || job.status === 'failed' || job.status === 'completed').map(job => job.id));
    setStoppingJobs(prev => {
      const newSet = new Set();
      for (const jobId of prev) {
        if (!finishedJobIds.has(jobId)) {
          newSet.add(jobId);
        }
      }
      return newSet;
    });
  }, [jobs]);

  const handleJobClick = (jobId: string) => {
    console.log('Job clicked:', jobId);
    navigate(`/sync?jobId=${jobId}`);
  };

  const handleStopJob = async (jobId: string) => {
    if (stoppingJobs.has(jobId)) return; // Prevent multiple clicks
    
    console.log('Stop button clicked for job:', jobId);
    setStoppingJobs(prev => new Set(prev).add(jobId));
    try {
      await dispatch(stopSyncJob(jobId)).unwrap();
      console.log('Stop job dispatched successfully for:', jobId);
      // Don't clear stopping state immediately - wait for job status to update
    } catch (error) {
      console.error('Failed to stop sync job:', error);
      // Only clear on error
      setStoppingJobs(prev => {
        const newSet = new Set(prev);
        newSet.delete(jobId);
        return newSet;
      });
    }
  };

  const handleStopAllJobs = async () => {
    if (isStoppingAll) return; // Prevent multiple clicks
    
    setIsStoppingAll(true);
    try {
      await dispatch(stopAllSyncJobs()).unwrap();
    } catch (error) {
      console.error('Failed to stop all sync jobs:', error);
    } finally {
      setIsStoppingAll(false);
    }
  };

  const handleClearFinishedJobs = async () => {
    if (isClearingFinished) return; // Prevent multiple clicks
    
    setIsClearingFinished(true);
    try {
      await dispatch(clearFinishedSyncJobs()).unwrap();
    } catch (error) {
      console.error('Failed to clear finished sync jobs:', error);
    } finally {
      setIsClearingFinished(false);
    }
  };

  const getJobStatusColor = (job: SyncJob) => {
    if (stoppingJobs.has(job.id)) {
      return 'text-orange-600 dark:text-orange-400';
    }
    switch (job.status) {
      case 'pending':
        return 'text-yellow-600 dark:text-yellow-400';
      case 'running':
        return 'text-blue-600 dark:text-blue-400';
      case 'completed':
        return 'text-green-600 dark:text-green-400';
      case 'stopped':
        return 'text-red-600 dark:text-red-400';
      case 'failed':
        return 'text-red-600 dark:text-red-400';
      default:
        return 'text-gray-600 dark:text-gray-400';
    }
  };

  const getJobStatusText = (job: SyncJob) => {
    if (stoppingJobs.has(job.id)) {
      return job.status === 'pending' ? 'Canceling' : 'Stopping';
    }
    switch (job.status) {
      case 'pending':
        return 'Pending';
      case 'running':
        return 'In Progress';
      case 'completed':
        return 'Completed';
      case 'stopped':
        return 'Stopped';
      case 'failed':
        return 'Failed';
      default:
        return 'Unknown';
    }
  };

  const getProgressPercentage = (job: SyncJob) => {
    if (job.totalTasks === 0) return 0;
    return Math.round(((job.totalTasks - job.remainingTasks) / job.totalTasks) * 100);
  };

  const formatDuration = (seconds?: number) => {
    if (!seconds || seconds < 1) return '0s';
    
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    
    if (hours > 0) {
      return `${hours}h ${minutes}m ${secs}s`;
    } else if (minutes > 0) {
      return `${minutes}m ${secs}s`;
    } else {
      return `${secs}s`;
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4">
        <div className="flex">
          <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
          </svg>
          <div className="ml-3">
            <h3 className="text-sm font-medium text-red-800 dark:text-red-200">Error</h3>
            <div className="mt-2 text-sm text-red-700 dark:text-red-300">{error}</div>
          </div>
        </div>
      </div>
    );
  }

  if (jobs.length === 0) {
    return (
      <div className="text-center py-12">
        <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
        </svg>
        <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-white">No sync jobs</h3>
        <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
          Start your first sync operation to see it here.
        </p>
      </div>
    );
  }

  // Sort jobs: running first, then pending, then stopped/failed, then completed, then by creation time (newest first)
  const sortedJobs = [...jobs].sort((a, b) => {
    // Priority order: running > pending > stopped/failed > completed
    const getStatusPriority = (status: string) => {
      switch (status) {
        case 'running': return 4;
        case 'pending': return 3;
        case 'stopped': return 2;
        case 'failed': return 2;
        case 'completed': return 1;
        default: return 0;
      }
    };
    
    const aPriority = getStatusPriority(a.status);
    const bPriority = getStatusPriority(b.status);
    
    if (aPriority !== bPriority) {
      return bPriority - aPriority; // Higher priority first
    }
    
    // Within same priority, sort by ID (which contains timestamp) descending
    return b.id.localeCompare(a.id);
  });

  const activeJobs = sortedJobs.filter(job => job.status === 'running' || job.status === 'pending');
  const hasActiveJobs = activeJobs.length > 0;
  
  const finishedJobs = sortedJobs.filter(job => job.status === 'completed' || job.status === 'stopped' || job.status === 'failed');
  const hasFinishedJobs = finishedJobs.length > 0;

  return (
    <div className="flex flex-col h-full">
      {/* Fixed header section */}
      <div className="flex-shrink-0 mb-4">
        {(hasActiveJobs || hasFinishedJobs) && (
          <div className="bg-white dark:bg-slate-800 shadow rounded-lg p-4">
            <div className="flex items-center justify-between">
              <div className="text-sm text-gray-900 dark:text-white">
                {hasActiveJobs && (
                  <div className="mb-1">
                    <span className="font-medium">
                      {activeJobs.length} active job{activeJobs.length !== 1 ? 's' : ''}
                    </span>
                    <span className="text-gray-500 dark:text-gray-400 font-normal ml-1">
                      ({sortedJobs.filter(job => job.status === 'running').length} running, {sortedJobs.filter(job => job.status === 'pending').length} pending)
                    </span>
                  </div>
                )}
                {hasFinishedJobs && (
                  <div>
                    <span className="font-medium">
                      {finishedJobs.length} finished job{finishedJobs.length !== 1 ? 's' : ''}
                    </span>
                    <span className="text-gray-500 dark:text-gray-400 font-normal ml-1">
                      ({sortedJobs.filter(job => job.status === 'completed').length} completed, {sortedJobs.filter(job => job.status === 'stopped').length} stopped, {sortedJobs.filter(job => job.status === 'failed').length} failed)
                    </span>
                  </div>
                )}
              </div>
              <div className="flex items-center space-x-2">
                {hasActiveJobs && (
                  <button
                    onClick={handleStopAllJobs}
                    disabled={isStoppingAll}
                    className="inline-flex items-center justify-center px-3 py-1.5 min-w-[84px] border border-transparent text-xs font-medium rounded-md text-red-700 bg-red-100 hover:bg-red-200 dark:bg-red-900/20 dark:text-red-400 dark:hover:bg-red-900/40 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {isStoppingAll ? (
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-red-600 mr-1"></div>
                    ) : (
                      <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 10l2 2 4-4" />
                      </svg>
                    )}
                    {isStoppingAll ? 'Stopping...' : 'Stop All'}
                  </button>
                )}
                {hasFinishedJobs && (
                  <button
                    onClick={handleClearFinishedJobs}
                    disabled={isClearingFinished}
                    className="inline-flex items-center justify-center px-3 py-1.5 min-w-[84px] border border-transparent text-xs font-medium rounded-md text-gray-700 bg-gray-100 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {isClearingFinished ? (
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-gray-600 mr-1"></div>
                    ) : (
                      <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    )}
                    {isClearingFinished ? 'Clearing...' : 'Clear Finished'}
                  </button>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
      
      {/* Scrollable job list */}
      <div className="flex-1 overflow-y-auto space-y-4">
        {sortedJobs.map((job) => {
          const progressPercentage = getProgressPercentage(job);
          const isActive = job.status === 'running';
          const isPending = job.status === 'pending';
          const isStopping = stoppingJobs.has(job.id);
          
          return (
            <div key={job.id} className="bg-white dark:bg-slate-800 shadow rounded-lg overflow-hidden">
              <div className="group flex hover:bg-gray-50 dark:hover:bg-slate-700 transition-colors">
                {/* Main clickable area */}
                <div
                  onClick={() => handleJobClick(job.id)}
                  className="flex-1 cursor-pointer px-4 py-4 sm:px-6"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between">
                        <div className="flex flex-col">
                          <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                            {job.path || 'root data folder'}
                          </p>
                          <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                            Job #{job.id.substring(0, 8)}
                          </p>
                        </div>
                        <div className="ml-2 flex-shrink-0 flex items-center space-x-2">
                          {isActive && !isStopping && (
                            <div className="animate-pulse">
                              <div className="h-2 w-2 bg-blue-500 rounded-full"></div>
                            </div>
                          )}
                          <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getJobStatusColor(job)}`}>
                            {getJobStatusText(job)}
                          </span>
                        </div>
                      </div>
                    
                    <div className="mt-2">
                      <div className="flex items-center justify-between text-sm text-gray-500 dark:text-gray-400">
                        <span>
                          {job.totalTasks - job.remainingTasks} of {job.totalTasks} tasks processed
                        </span>
                        <span>{progressPercentage}%</span>
                      </div>
                      
                      {/* Progress bar */}
                      <div className="mt-1 w-full bg-gray-200 dark:bg-gray-600 rounded-full h-2">
                        <div 
                          className={`h-2 rounded-full transition-all duration-300 ${
                            isActive && !isStopping ? 'bg-blue-500' : 'bg-green-500'
                          }`}
                          style={{ width: `${progressPercentage}%` }}
                        />
                      </div>
                    </div>

                    {job.remainingTasks > 0 && isActive && !isStopping && (
                      <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                        {job.remainingTasks} tasks remaining
                      </p>
                    )}

                    {/* Duration for completed/failed jobs */}
                    {(job.status === 'completed' || job.status === 'failed' || job.status === 'stopped') && job.duration !== undefined && (
                      <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                        Duration: {formatDuration(job.duration)}
                      </p>
                    )}

                    {/* Message for any job status */}
                    {job.message && (
                      <div className={`mt-2 p-2 border rounded ${
                        job.status === 'failed'
                          ? 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800 text-red-700 dark:text-red-300'
                          : 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800 text-blue-700 dark:text-blue-300'
                      }`}>
                        <div className="flex items-start">
                          <svg className="w-4 h-4 mr-2 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d={
                              job.status === 'failed'
                                ? "M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
                                : "M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                            } />
                          </svg>
                          <p className="text-xs truncate">
                            {job.message}
                          </p>
                        </div>
                      </div>
                    )}
                  </div>
                  
                    <div className="ml-4 flex-shrink-0">
                      <svg className="h-5 w-5 text-gray-400 group-hover:text-gray-600 dark:group-hover:text-gray-300 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5l7 7-7 7" />
                      </svg>
                    </div>
                  </div>
                </div>
                
                {/* Stop/Cancel button for active or pending jobs */}
                {(isActive || isPending) && (
                  <div className="flex items-center px-4 py-4">
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleStopJob(job.id);
                      }}
                      disabled={stoppingJobs.has(job.id)}
                      className="inline-flex items-center px-2.5 py-1.5 border border-transparent text-xs font-medium rounded text-red-700 bg-red-100 hover:bg-red-200 dark:bg-red-900/20 dark:text-red-400 dark:hover:bg-red-900/40 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 relative z-10 disabled:opacity-50 disabled:cursor-not-allowed"
                      title={isPending ? "Cancel this sync job" : "Stop this sync job"}
                    >
                      {stoppingJobs.has(job.id) ? (
                        <div className="animate-spin rounded-full h-3 w-3 border-b-2 border-red-600 mr-1"></div>
                      ) : (
                        <svg className="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      )}
                      {stoppingJobs.has(job.id) 
                        ? (isPending ? 'Canceling...' : 'Stopping...')
                        : (isPending ? 'Cancel' : 'Stop')
                      }
                    </button>
                  </div>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
