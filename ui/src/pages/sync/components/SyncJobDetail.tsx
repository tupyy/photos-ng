import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAppDispatch } from '@shared/store';
import { stopSyncJob } from '@shared/reducers/syncSlice';
import { useSyncJobDetail } from '@shared/hooks/useSyncJobDetail';
import { TaskResult } from '@generated/models';

type TaskFilter = 'all' | 'success' | 'error';

interface SyncJobDetailProps {
  jobId: string;
}

export const SyncJobDetail: React.FC<SyncJobDetailProps> = ({ jobId }) => {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const [taskFilter, setTaskFilter] = useState<TaskFilter>('all');
  const [isStoppingJob, setIsStoppingJob] = useState(false);
  
  // Use centralized job detail hook for fetching and polling
  const { job, isActive } = useSyncJobDetail(jobId);
  const isPending = job?.status === 'pending';

  // Polling is now handled by the useSyncJobDetail hook

  const handleStopJob = async () => {
    if (isStoppingJob) return; // Prevent multiple clicks
    
    setIsStoppingJob(true);
    try {
      await dispatch(stopSyncJob(jobId)).unwrap();
      // Don't clear isStoppingJob immediately - wait for the job status to update to 'stopped'
      // The polling will pick up the status change and UI will update accordingly
    } catch (error) {
      console.error('Failed to stop sync job:', error);
      setIsStoppingJob(false); // Only clear on error
    }
  };

  // Clear stopping state when job is no longer active or pending
  useEffect(() => {
    if (!isActive && !isPending && isStoppingJob) {
      setIsStoppingJob(false);
    }
  }, [isActive, isPending, isStoppingJob]);

  const getProgressPercentage = () => {
    if (!job || job.totalTasks === 0) return 0;
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

  const isTaskSuccessful = (taskResult: TaskResult) => {
    // TaskResult.result is a string - "ok" means success, anything else is an error message
    return taskResult.result === 'ok';
  };

  const getTaskResultText = (taskResult: TaskResult) => {
    return taskResult.result;
  };

  const getFileResultColor = (taskResult: TaskResult) => {
    return isTaskSuccessful(taskResult) ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400';
  };

  const getFileResultIcon = (taskResult: TaskResult) => {
    if (isTaskSuccessful(taskResult)) {
      return (
        <svg className="h-4 w-4 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 13l4 4L19 7" />
        </svg>
      );
    } else {
      return (
        <svg className="h-4 w-4 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      );
    }
  };

  const getFilteredTasks = () => {
    if (!job?.completedTasks) return [];
    
    switch (taskFilter) {
      case 'success':
        return job.completedTasks.filter(task => isTaskSuccessful(task));
      case 'error':
        return job.completedTasks.filter(task => !isTaskSuccessful(task));
      default:
        return job.completedTasks;
    }
  };

  if (!job) {
    return (
      <div className="flex justify-center items-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  const progressPercentage = getProgressPercentage();
  const successfulFiles = job.completedTasks?.filter(task => isTaskSuccessful(task)).length || 0;
  const failedFiles = job.completedTasks?.filter(task => !isTaskSuccessful(task)).length || 0;

  return (
    <div className="space-y-6">
      {/* Job Overview */}
      <div className="bg-white dark:bg-slate-800 shadow rounded-lg p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-medium text-gray-900 dark:text-white">Job Overview</h2>
          {(isActive || isPending) && (
            <button
              onClick={handleStopJob}
              disabled={isStoppingJob}
              className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-red-700 bg-red-100 hover:bg-red-200 dark:bg-red-900/20 dark:text-red-400 dark:hover:bg-red-900/40 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isStoppingJob ? (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-red-600 mr-2"></div>
              ) : (
                <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              )}
              {isStoppingJob 
                ? (isPending ? 'Canceling...' : 'Stopping...')
                : (isPending ? 'Cancel Job' : 'Stop Job')
              }
            </button>
          )}
        </div>

        {/* Folder Path */}
        <div className="mb-4 flex items-center">
          <svg className="w-5 h-5 text-gray-400 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-5l-2-2H5a2 2 0 00-2 2z" />
          </svg>
          <p className="text-sm text-gray-900 dark:text-white">
            <span className="font-medium">Syncing Folder:</span> 
            <span className="ml-2 font-mono text-gray-600 dark:text-gray-300">
              {job.path || 'root data folder'}
            </span>
          </p>
        </div>

        {/* Duration */}
        {job.duration !== undefined && (
          <div className="mb-4 flex items-center">
            <svg className="w-5 h-5 text-gray-400 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <p className="text-sm text-gray-900 dark:text-white">
              <span className="font-medium">Duration:</span> 
              <span className="ml-2 text-gray-600 dark:text-gray-300">
                {formatDuration(job.duration)}
              </span>
            </p>
          </div>
        )}

        {/* Job Message */}
        {job.message && (
          <div className={`mb-6 p-4 border rounded-md ${
            job.status === 'failed'
              ? 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800'
              : 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800'
          }`}>
            <div className="flex">
              <svg className="w-5 h-5 mr-3 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d={
                  job.status === 'failed'
                    ? "M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
                    : "M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                } />
              </svg>
              <div className="flex-1">
                <h3 className={`text-sm font-medium ${
                  job.status === 'failed'
                    ? 'text-red-800 dark:text-red-200'
                    : 'text-blue-800 dark:text-blue-200'
                }`}>
                  {job.status === 'failed' ? 'Job Failed' : 'Job Information'}
                </h3>
                <p className={`mt-1 text-sm ${
                  job.status === 'failed'
                    ? 'text-red-700 dark:text-red-300'
                    : 'text-blue-700 dark:text-blue-300'
                }`}>
                  {job.message}
                </p>
              </div>
            </div>
          </div>
        )}
        
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          <div className="text-center">
            <div className="text-2xl font-bold text-gray-900 dark:text-white">{job.totalTasks}</div>
            <div className="text-sm text-gray-500 dark:text-gray-400">Total Tasks</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-green-600 dark:text-green-400">{successfulFiles}</div>
            <div className="text-sm text-gray-500 dark:text-gray-400">Processed</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-red-600 dark:text-red-400">{failedFiles}</div>
            <div className="text-sm text-gray-500 dark:text-gray-400">Errors</div>
          </div>
        </div>

        {/* Progress Bar */}
        <div className="mb-4">
          <div className="flex items-center justify-between text-sm text-gray-600 dark:text-gray-400 mb-2">
            <span>Progress</span>
            <span>{progressPercentage}%</span>
          </div>
          <div className="w-full bg-gray-200 dark:bg-gray-600 rounded-full h-3">
            <div 
              className={`h-3 rounded-full transition-all duration-300 ${
                isActive ? 'bg-blue-500' : 'bg-green-500'
              }`}
              style={{ width: `${progressPercentage}%` }}
            />
          </div>
        </div>

        {/* Status */}
        <div className="flex items-center justify-between">
          <div className="flex items-center">
            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
              job.status === 'pending' && isStoppingJob
                ? 'text-orange-700 bg-orange-100 dark:text-orange-300 dark:bg-orange-900'
                : job.status === 'pending' 
                ? 'text-yellow-700 bg-yellow-100 dark:text-yellow-300 dark:bg-yellow-900'
                : job.status === 'running' && isStoppingJob
                ? 'text-orange-700 bg-orange-100 dark:text-orange-300 dark:bg-orange-900'
                : job.status === 'running'
                ? 'text-blue-700 bg-blue-100 dark:text-blue-300 dark:bg-blue-900'
                : job.status === 'completed'
                ? 'text-green-700 bg-green-100 dark:text-green-300 dark:bg-green-900'
                : job.status === 'failed'
                ? 'text-red-700 bg-red-100 dark:text-red-300 dark:bg-red-900'
                : 'text-gray-700 bg-gray-100 dark:text-gray-300 dark:bg-gray-900'
            }`}>
              {job.status === 'pending' && isStoppingJob ? 'Canceling' :
               job.status === 'pending' ? 'Pending' : 
               job.status === 'running' && isStoppingJob ? 'Stopping' :
               job.status === 'running' ? 'In Progress' : 
               job.status === 'completed' ? 'Completed' :
               job.status === 'failed' ? 'Failed' :
               job.status === 'stopped' ? 'Stopped' : 'Unknown'}
            </span>
            {(job.status === 'pending' || (job.status === 'running' && !isStoppingJob)) && (
              <span className="ml-2 text-sm text-gray-500 dark:text-gray-400">
                {job.status === 'pending' ? 'Waiting to start...' : `${job.remainingTasks} tasks remaining`}
              </span>
            )}
          </div>
          {job.status === 'running' && !isStoppingJob && (
            <div className="flex items-center text-sm text-gray-500 dark:text-gray-400">
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 mr-2"></div>
              Processing...
            </div>
          )}
          {job.status === 'pending' && (
            <div className="flex items-center text-sm text-gray-500 dark:text-gray-400">
              <div className="animate-pulse rounded-full h-4 w-4 bg-yellow-500 mr-2"></div>
              Waiting to start...
            </div>
          )}
        </div>
      </div>

      {/* Task Results */}
      {job.completedTasks && job.completedTasks.length > 0 && (
        <div className="bg-white dark:bg-slate-800 shadow rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-medium text-gray-900 dark:text-white">Task Processing Results</h2>
              
              {/* Filter Buttons */}
              <div className="flex space-x-1">
                <button
                  onClick={() => setTaskFilter('all')}
                  className={`px-3 py-1 text-xs font-medium rounded-full transition-colors ${
                    taskFilter === 'all'
                      ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300'
                      : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-400 dark:hover:bg-gray-600'
                  }`}
                >
                  All ({job.completedTasks.length})
                </button>
                <button
                  onClick={() => setTaskFilter('success')}
                  className={`px-3 py-1 text-xs font-medium rounded-full transition-colors ${
                    taskFilter === 'success'
                      ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
                      : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-400 dark:hover:bg-gray-600'
                  }`}
                >
                  Success ({successfulFiles})
                </button>
                <button
                  onClick={() => setTaskFilter('error')}
                  className={`px-3 py-1 text-xs font-medium rounded-full transition-colors ${
                    taskFilter === 'error'
                      ? 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300'
                      : 'bg-gray-100 text-gray-600 hover:bg-gray-200 dark:bg-gray-700 dark:text-gray-400 dark:hover:bg-gray-600'
                  }`}
                >
                  Error ({failedFiles})
                </button>
              </div>
            </div>
          </div>
          
          <div className="max-h-96 overflow-y-auto">
            {getFilteredTasks().length > 0 ? (
              <ul className="divide-y divide-gray-200 dark:divide-gray-700">
                {getFilteredTasks().map((task, index) => (
                <li key={index} className="px-6 py-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center min-w-0 flex-1">
                      <div className="flex-shrink-0 mr-3">
                        {getFileResultIcon(task)}
                      </div>
                      <div className="min-w-0 flex-1">
                        <div className="flex items-center gap-2">
                          <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                            {task.item}
                          </p>
                          <span className={`inline-flex items-center px-1.5 py-0.5 rounded-full text-xs font-medium ${
                            task.itemType === 'file' 
                              ? 'text-blue-700 bg-blue-100 dark:text-blue-300 dark:bg-blue-900' 
                              : 'text-purple-700 bg-purple-100 dark:text-purple-300 dark:bg-purple-900'
                          }`}>
                            {task.itemType}
                          </span>
                        </div>
                        {!isTaskSuccessful(task) && (
                          <p className="text-sm text-red-600 dark:text-red-400 mt-1">
                            {getTaskResultText(task)}
                          </p>
                        )}
                        <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                          Duration: {task.duration}ms
                        </p>
                      </div>
                    </div>
                    <div className="flex-shrink-0">
                      <span className={`text-sm font-medium ${getFileResultColor(task)}`}>
                        {isTaskSuccessful(task) ? 'Success' : 'Error'}
                      </span>
                    </div>
                  </div>
                </li>
              ))}
              </ul>
            ) : (
              <div className="px-6 py-8 text-center">
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  No tasks match the selected filter.
                </p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};
