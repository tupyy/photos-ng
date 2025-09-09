import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAppDispatch, useAppSelector, selectSyncJobs, selectSyncLoading, selectSyncError } from '@shared/store';
import { stopSyncJob, stopAllSyncJobs } from '@shared/reducers/syncSlice';
import { SyncJob } from '@generated/models';

// Extended SyncJob interface to include path information
interface SyncJobWithPath extends SyncJob {
  path?: string;
}

export const SyncJobsList: React.FC = () => {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const jobs = useAppSelector(selectSyncJobs);
  const loading = useAppSelector(selectSyncLoading);
  const error = useAppSelector(selectSyncError);
  const [stoppingJobs, setStoppingJobs] = useState<Set<string>>(new Set());
  const [isStoppingAll, setIsStoppingAll] = useState(false);

  // This component is purely presentational - it only consumes Redux state
  // Data fetching and polling is handled by parent components

  // Clear stopping state for jobs that are no longer active
  useEffect(() => {
    const activeJobIds = new Set(jobs.filter(job => job.status === 'running').map(job => job.id));
    setStoppingJobs(prev => {
      const newSet = new Set();
      for (const jobId of prev) {
        if (activeJobIds.has(jobId)) {
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

  const getJobStatusColor = (job: SyncJobWithPath) => {
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

  const getJobStatusText = (job: SyncJobWithPath) => {
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

  const getProgressPercentage = (job: SyncJobWithPath) => {
    if (job.totalTasks === 0) return 0;
    return Math.round(((job.totalTasks - job.remainingTasks) / job.totalTasks) * 100);
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

  // Sort jobs: pending/running first, then stopped/failed, then completed, then by creation time (newest first)
  const sortedJobs = [...jobs].sort((a, b) => {
    // Priority order: pending/running > stopped/failed > completed
    const getStatusPriority = (status: string) => {
      switch (status) {
        case 'pending': return 4;
        case 'running': return 3;
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

  const activeJobs = sortedJobs.filter(job => job.status === 'running');
  const hasActiveJobs = activeJobs.length > 0;

  return (
    <div className="bg-white dark:bg-slate-800 shadow overflow-hidden sm:rounded-md">
      {/* Header with Stop All button */}
      {hasActiveJobs && (
        <div className="bg-gray-50 dark:bg-slate-700 px-4 py-3 border-b border-gray-200 dark:border-gray-600">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-medium text-gray-900 dark:text-white">
              {activeJobs.length} active job{activeJobs.length !== 1 ? 's' : ''}
            </h3>
            <button
              onClick={handleStopAllJobs}
              disabled={isStoppingAll}
              className="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded-md text-red-700 bg-red-100 hover:bg-red-200 dark:bg-red-900/20 dark:text-red-400 dark:hover:bg-red-900/40 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed"
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
          </div>
        </div>
      )}
      
      <ul className="divide-y divide-gray-200 dark:divide-gray-700">
        {sortedJobs.map((job) => {
          const progressPercentage = getProgressPercentage(job);
          const isActive = job.status === 'running';
          
          return (
            <li key={job.id}>
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
                            Sync Job #{job.id.substring(5, 13)}
                          </p>
                          <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                            Path: {job.path === '' || !job.path ? 'root data folder' : job.path}
                          </p>
                        </div>
                        <div className="ml-2 flex-shrink-0 flex items-center space-x-2">
                          {isActive && (
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
                            isActive ? 'bg-blue-500' : 'bg-green-500'
                          }`}
                          style={{ width: `${progressPercentage}%` }}
                        />
                      </div>
                    </div>

                    {job.remainingTasks > 0 && (
                      <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                        {job.remainingTasks} tasks remaining
                      </p>
                    )}
                  </div>
                  
                    <div className="ml-4 flex-shrink-0">
                      <svg className="h-5 w-5 text-gray-400 group-hover:text-gray-600 dark:group-hover:text-gray-300 transition-colors" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5l7 7-7 7" />
                      </svg>
                    </div>
                  </div>
                </div>
                
                {/* Stop button for active jobs */}
                {isActive && (
                  <div className="flex items-center px-4 py-4">
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleStopJob(job.id);
                      }}
                      disabled={stoppingJobs.has(job.id)}
                      className="inline-flex items-center px-2.5 py-1.5 border border-transparent text-xs font-medium rounded text-red-700 bg-red-100 hover:bg-red-200 dark:bg-red-900/20 dark:text-red-400 dark:hover:bg-red-900/40 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 relative z-10 disabled:opacity-50 disabled:cursor-not-allowed"
                      title="Stop this sync job"
                    >
                      {stoppingJobs.has(job.id) ? (
                        <div className="animate-spin rounded-full h-3 w-3 border-b-2 border-red-600 mr-1"></div>
                      ) : (
                        <svg className="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                      )}
                      {stoppingJobs.has(job.id) ? 'Stopping...' : 'Stop'}
                    </button>
                  </div>
                )}
              </div>
            </li>
          );
        })}
      </ul>
    </div>
  );
};
