/**
 * Sync Page Component
 *
 * Main page for sync operation management in the Photos NG application.
 * Provides functionality for:
 * - Viewing all sync jobs and their progress
 * - Starting new sync operations
 * - Monitoring detailed progress of individual sync jobs
 * - Viewing sync results and error messages
 */

import React, { useEffect, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useAppDispatch } from '@shared/store';
import { fetchSyncJobs } from '@shared/reducers/syncSlice';
import { useSyncPolling } from '@shared/hooks/useSyncPolling';
import { SyncJobsList } from './components/SyncJobsList';
import { SyncJobDetail } from './components/SyncJobDetail';
import { StartSyncModal } from './components/StartSyncModal';

const SyncPage: React.FC = () => {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const [searchParams] = useSearchParams();
  
  // Get jobId from URL params if viewing a specific job
  const jobId = searchParams.get('jobId');
  const [showStartModal, setShowStartModal] = useState(false);
  const [resetPolling, setResetPolling] = useState(false);
  
  // Use centralized sync polling hook - only poll when viewing jobs list
  const { jobs, hasActiveJobs, isPolling } = useSyncPolling({
    enabled: !jobId, // Only poll when viewing jobs list, not individual job detail
    interval: 2000, // Poll every 2 seconds
    resetExponentialPolling: resetPolling,
  });

  // Initial fetch of sync jobs when component mounts
  useEffect(() => {
    dispatch(fetchSyncJobs({}));
  }, [dispatch]);

  // Reset the polling reset flag after it's been applied
  useEffect(() => {
    if (resetPolling) {
      setResetPolling(false);
    }
  }, [resetPolling]);

  // Auto-start sync if path is provided in URL params
  const autoStartPath = searchParams.get('autoStart');
  
  useEffect(() => {
    // Only show modal if autoStart is requested and we're not viewing a specific job
    if (autoStartPath !== null && !jobId) {
      setShowStartModal(true);
    }
  }, [autoStartPath, jobId]);

  // Calculate job statistics
  const pendingJobs = jobs.filter(job => job.status === 'pending');
  const activeJobs = jobs.filter(job => job.status === 'running');
  const completedJobs = jobs.filter(job => job.status === 'completed');
  const stoppedJobs = jobs.filter(job => job.status === 'stopped');
  const failedJobs = jobs.filter(job => job.status === 'failed');

  const handleBack = () => {
    if (jobId) {
      // If viewing a specific job, go back to job list
      navigate('/sync');
    } else {
      // If viewing job list, go back to main page
      navigate('/');
    }
  };

  const handleStartSync = () => {
    setShowStartModal(true);
  };

  const handleStartSyncClose = () => {
    setShowStartModal(false);
    // Remove autoStart param if it exists
    if (autoStartPath !== null) {
      const newParams = new URLSearchParams(searchParams);
      newParams.delete('autoStart');
      navigate({ search: newParams.toString() }, { replace: true });
    }
  };

  const handleJobCreated = () => {
    setShowStartModal(false);
    // Fetch updated job list after creation to ensure we have the latest data
    dispatch(fetchSyncJobs({}));
    // Reset exponential polling to start from 2 seconds again
    setResetPolling(true);
  };

  return (
    <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
      <div className="px-4 py-6 sm:px-0">
        {/* Header */}
        <div className="mb-6">
          <div className="flex items-center justify-between">
            <button
              onClick={handleBack}
              className="flex items-center text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors"
            >
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 19l-7-7 7-7" />
              </svg>
              Back
            </button>

            {!jobId && (
              <button
                onClick={handleStartSync}
                className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors"
              >
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor" className="w-4 h-4 mr-2">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
                </svg>
                Start Sync
              </button>
            )}
          </div>

          <div className="mt-4">
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
              {jobId ? 'Sync Job Details' : 'Sync Operations'}
            </h1>
            <p className="mt-1 text-gray-600 dark:text-gray-400">
              {jobId 
                ? 'Monitor the progress and results of your sync operation'
                : 'Manage and monitor your sync operations'
              }
            </p>
            
            {/* Jobs Summary - only show on main sync page */}
            {!jobId && jobs.length > 0 && (
              <div className="mt-3 flex items-center space-x-4 text-sm flex-wrap">
                {pendingJobs.length > 0 && (
                  <div className="flex items-center">
                    <div className="h-2 w-2 bg-yellow-500 rounded-full mr-2"></div>
                    <span className="text-gray-600 dark:text-gray-400">
                      {pendingJobs.length} pending
                    </span>
                  </div>
                )}
                {activeJobs.length > 0 && (
                  <div className="flex items-center">
                    <div className="h-2 w-2 bg-blue-500 rounded-full mr-2"></div>
                    <span className="text-gray-600 dark:text-gray-400">
                      {activeJobs.length} running
                    </span>
                  </div>
                )}
                {completedJobs.length > 0 && (
                  <div className="flex items-center">
                    <div className="h-2 w-2 bg-green-500 rounded-full mr-2"></div>
                    <span className="text-gray-600 dark:text-gray-400">
                      {completedJobs.length} completed
                    </span>
                  </div>
                )}
                {stoppedJobs.length > 0 && (
                  <div className="flex items-center">
                    <div className="h-2 w-2 bg-red-500 rounded-full mr-2"></div>
                    <span className="text-gray-600 dark:text-gray-400">
                      {stoppedJobs.length} stopped
                    </span>
                  </div>
                )}
                {failedJobs.length > 0 && (
                  <div className="flex items-center">
                    <div className="h-2 w-2 bg-red-600 rounded-full mr-2"></div>
                    <span className="text-gray-600 dark:text-gray-400">
                      {failedJobs.length} failed
                    </span>
                  </div>
                )}
                {(pendingJobs.length > 0 || activeJobs.length > 0) && (
                  <div className={`flex items-center ${pendingJobs.length > 0 && activeJobs.length === 0 ? 'text-yellow-600 dark:text-yellow-400' : 'text-blue-600 dark:text-blue-400'}`}>
                    {pendingJobs.length > 0 && activeJobs.length === 0 ? (
                      <div className="animate-pulse h-3 w-3 bg-yellow-500 rounded-full mr-1"></div>
                    ) : (
                      <svg className="animate-spin h-3 w-3 mr-1" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                        <path className="opacity-75" fill="currentColor" d="m100 50.5908c0 27.6142-22.3858 50-50 50s-50-22.3858-50-50 22.3858-50 50-50 50 22.3858 50 50zm-9.08144 0c0-22.5981-18.4013-40.9186-40.9186-40.9186s-40.9186 18.3205-40.9186 40.9186 18.3205 40.9186 40.9186 40.9186 40.9186-18.3205 40.9186-40.9186zm-90.77316-5.909c-2.425 0-4.392-1.967-4.392-4.392s1.967-4.392 4.392-4.392 4.392 1.967 4.392 4.392-1.967 4.392-4.392 4.392z"></path>
                      </svg>
                    )}
                    <span>{pendingJobs.length > 0 && activeJobs.length === 0 ? 'Waiting to start...' : 'Processing...'}</span>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>

        {/* Content */}
        {jobId ? (
          <SyncJobDetail jobId={jobId} />
        ) : (
          <div className="h-[calc(100vh-16rem)]">
            <SyncJobsList />
          </div>
        )}

        {/* Start Sync Modal */}
        <StartSyncModal 
          isOpen={showStartModal}
          onClose={handleStartSyncClose}
          onJobCreated={handleJobCreated}
          initialPath={autoStartPath || ''}
        />
      </div>
    </div>
  );
};

export default SyncPage;
