import React, { useState, useEffect, useRef } from 'react';
import { useAppDispatch, useAppSelector, selectSyncStarting, selectSyncError } from '@shared/store';
import { startSyncJob, clearError } from '@shared/reducers/syncSlice';

interface StartSyncModalProps {
  isOpen: boolean;
  onClose: () => void;
  onJobCreated: (jobId: string) => void;
  initialPath?: string;
}

export const StartSyncModal: React.FC<StartSyncModalProps> = ({ 
  isOpen, 
  onClose, 
  onJobCreated, 
  initialPath = '' 
}) => {
  const dispatch = useAppDispatch();
  const [path, setPath] = useState(initialPath);
  const isSubmitting = useAppSelector(selectSyncStarting);
  const error = useAppSelector(selectSyncError);
  const inputRef = useRef<HTMLInputElement>(null);

  // Update path when initialPath changes
  useEffect(() => {
    setPath(initialPath);
  }, [initialPath]);

  // Focus input when modal opens
  useEffect(() => {
    if (isOpen && inputRef.current) {
      // Small delay to ensure modal is fully rendered
      const timeoutId = setTimeout(() => {
        inputRef.current?.focus();
      }, 100);
      return () => clearTimeout(timeoutId);
    }
  }, [isOpen]);

  // Auto-submit if initialPath is provided (for autoStart functionality)
  useEffect(() => {
    if (isOpen && initialPath && !isSubmitting) {
      handleStartSync();
    }
  }, [isOpen, initialPath, isSubmitting]);

  const handleStartSync = async () => {
    try {
      const resultAction = await dispatch(startSyncJob(path.trim() || ''));
      if (startSyncJob.fulfilled.match(resultAction)) {
        onJobCreated(resultAction.payload.jobId);
      }
    } catch (err: any) {
      console.error('Failed to start sync job:', err);
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    handleStartSync();
  };

  const handleClose = () => {
    if (!isSubmitting) {
      setPath('');
      dispatch(clearError());
      onClose();
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        {/* Background overlay */}
        <div className="fixed inset-0 transition-opacity" onClick={handleClose}>
          <div className="absolute inset-0 bg-gray-500 dark:bg-gray-900 opacity-75"></div>
        </div>

        {/* Modal panel */}
        <div className="relative inline-block align-bottom bg-white dark:bg-slate-800 rounded-lg px-4 pt-5 pb-4 text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full sm:p-6">
          <div>
            <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-blue-100 dark:bg-blue-900">
              <svg className="h-6 w-6 text-blue-600 dark:text-blue-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
              </svg>
            </div>
            <div className="mt-3 text-center sm:mt-5">
              <h3 className="text-lg leading-6 font-medium text-gray-900 dark:text-white">
                Start Sync Operation
              </h3>
              <div className="mt-2">
                <p className="text-sm text-gray-500 dark:text-gray-400">
                  Start a new sync operation to process files from the specified path. 
                  Multiple sync jobs can run simultaneously for different paths.
                </p>
                <p className="text-xs text-gray-400 dark:text-gray-500 mt-1">
                  Leave path empty to sync from the root directory.
                </p>
              </div>
            </div>
          </div>

          <form onSubmit={handleSubmit} className="mt-5 sm:mt-6">
            <div className="mb-4">
              <label htmlFor="path" className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Path (relative to root)
              </label>
              <input
                type="text"
                id="path"
                ref={inputRef}
                value={path}
                onChange={(e) => setPath(e.target.value)}
                placeholder="/subfolder or leave empty for root"
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent dark:bg-gray-700 dark:text-white"
                disabled={isSubmitting}
              />
              <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                Examples: "" (root), "photos/2024", "albums/vacation"
              </p>
            </div>

            {error && (
              <div className="mb-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-3">
                <div className="flex">
                  <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                  </svg>
                  <div className="ml-3">
                    <p className="text-sm text-red-800 dark:text-red-200">{error}</p>
                  </div>
                </div>
              </div>
            )}

            <div className="flex space-x-3">
              <button
                type="button"
                onClick={handleClose}
                disabled={isSubmitting}
                className="flex-1 inline-flex justify-center rounded-md border border-gray-300 dark:border-gray-600 shadow-sm px-4 py-2 bg-white dark:bg-slate-700 text-base font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-slate-600 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={isSubmitting}
                className="flex-1 inline-flex justify-center items-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-blue-600 text-base font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isSubmitting ? (
                  <>
                    <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                    Starting...
                  </>
                ) : (
                  'Start Sync'
                )}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};
