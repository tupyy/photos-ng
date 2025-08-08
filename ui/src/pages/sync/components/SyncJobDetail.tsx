import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useAppDispatch } from '@shared/store';
import { stopSyncJob } from '@shared/reducers/syncSlice';
import { useSyncJobDetail } from '@shared/hooks/useSyncJobDetail';
import { ProcessedFile } from '@generated/models';

interface SyncJobDetailProps {
  jobId: string;
}

export const SyncJobDetail: React.FC<SyncJobDetailProps> = ({ jobId }) => {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  
  // Use centralized job detail hook for fetching and polling
  const { job, isActive } = useSyncJobDetail(jobId);

  // Polling is now handled by the useSyncJobDetail hook

  const handleStopJob = async () => {
    try {
      await dispatch(stopSyncJob(jobId)).unwrap();
      navigate('/sync'); // Navigate back to sync list after stopping
    } catch (error) {
      console.error('Failed to stop sync job:', error);
    }
  };

  const getProgressPercentage = () => {
    if (!job || job.totalFiles === 0) return 0;
    return Math.round(((job.totalFiles - job.filesRemaining) / job.totalFiles) * 100);
  };

  const getFileResultColor = (result: string) => {
    return result === 'ok' ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400';
  };

  const getFileResultIcon = (result: string) => {
    if (result === 'ok') {
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

  if (!job) {
    return (
      <div className="flex justify-center items-center py-12">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  const progressPercentage = getProgressPercentage();
  const successfulFiles = job.filesProcessed?.filter(f => f.result === 'ok').length || 0;
  const failedFiles = job.filesProcessed?.filter(f => f.result !== 'ok').length || 0;

  return (
    <div className="space-y-6">
      {/* Job Overview */}
      <div className="bg-white dark:bg-slate-800 shadow rounded-lg p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-medium text-gray-900 dark:text-white">Job Overview</h2>
          {isActive && (
            <button
              onClick={handleStopJob}
              className="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-red-700 bg-red-100 hover:bg-red-200 dark:bg-red-900/20 dark:text-red-400 dark:hover:bg-red-900/40 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
            >
              <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
              Stop Job
            </button>
          )}
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          <div className="text-center">
            <div className="text-2xl font-bold text-gray-900 dark:text-white">{job.totalFiles}</div>
            <div className="text-sm text-gray-500 dark:text-gray-400">Total Files</div>
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
              isActive ? 'text-blue-700 bg-blue-100 dark:text-blue-300 dark:bg-blue-900' : 'text-green-700 bg-green-100 dark:text-green-300 dark:bg-green-900'
            }`}>
              {isActive ? 'In Progress' : 'Completed'}
            </span>
            {isActive && (
              <span className="ml-2 text-sm text-gray-500 dark:text-gray-400">
                {job.filesRemaining} files remaining
              </span>
            )}
          </div>
          {isActive && (
            <div className="flex items-center text-sm text-gray-500 dark:text-gray-400">
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 mr-2"></div>
              Processing...
            </div>
          )}
        </div>
      </div>

      {/* File Results */}
      {job.filesProcessed && job.filesProcessed.length > 0 && (
        <div className="bg-white dark:bg-slate-800 shadow rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <h2 className="text-lg font-medium text-gray-900 dark:text-white">File Processing Results</h2>
          </div>
          
          <div className="max-h-96 overflow-y-auto">
            <ul className="divide-y divide-gray-200 dark:divide-gray-700">
              {job.filesProcessed.map((file, index) => (
                <li key={index} className="px-6 py-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center min-w-0 flex-1">
                      <div className="flex-shrink-0 mr-3">
                        {getFileResultIcon(file.result)}
                      </div>
                      <div className="min-w-0 flex-1">
                        <p className="text-sm font-medium text-gray-900 dark:text-white truncate">
                          {file.filename}
                        </p>
                        {file.result !== 'ok' && (
                          <p className="text-sm text-red-600 dark:text-red-400 mt-1">
                            {file.result}
                          </p>
                        )}
                      </div>
                    </div>
                    <div className="flex-shrink-0">
                      <span className={`text-sm font-medium ${getFileResultColor(file.result)}`}>
                        {file.result === 'ok' ? 'Success' : 'Error'}
                      </span>
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          </div>
        </div>
      )}
    </div>
  );
};
