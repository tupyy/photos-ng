import React, { useEffect, useRef } from 'react';
import { useSync } from '@hooks/useSync';

export interface ActionMenuProps {
  isOpen: boolean;
  onToggle: () => void;
  onClose: () => void;
  onCreateAlbum: () => void;
}

const ActionMenu: React.FC<ActionMenuProps> = ({ isOpen, onToggle, onClose, onCreateAlbum }) => {
  const { isInProgress: syncInProgress, progress: syncProgress, start: startSyncAction, cancel: cancelSyncAction } = useSync();
  const syncPromiseRef = useRef<any>(null);



  // Cleanup sync on component unmount
  useEffect(() => {
    return () => {
      if (syncPromiseRef.current) {
        syncPromiseRef.current.abort();
      }
    };
  }, []);

  const handleSync = () => {
    onClose();
    syncPromiseRef.current = startSyncAction();
  };

  const handleStopSync = () => {
    if (syncPromiseRef.current) {
      syncPromiseRef.current.abort();
      syncPromiseRef.current = null;
    }
    cancelSyncAction();
  };

  const handleCreateAlbum = () => {
    onClose();
    onCreateAlbum();
  };

  return (
    <div className="relative">
      <button
        type="button"
        onClick={onToggle}
        className="p-2 text-gray-500 hover:text-gray-700 dark:text-slate-400 dark:hover:text-gray-200 transition-colors"
        title="Actions"
      >
        {/* Three dots icon */}
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor" className="w-5 h-5">
          <path strokeLinecap="round" strokeLinejoin="round" d="M12 6.75a.75.75 0 1 1 0-1.5.75.75 0 0 1 0 1.5ZM12 12.75a.75.75 0 1 1 0-1.5.75.75 0 0 1 0 1.5ZM12 18.75a.75.75 0 1 1 0-1.5.75.75 0 0 1 0 1.5Z" />
        </svg>
      </button>

      {/* Dropdown Menu */}
      {isOpen && (
        <div className="absolute right-0 mt-2 w-48 bg-white dark:bg-slate-800 rounded-md shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none z-50">
          <div className="py-1">
            {syncInProgress ? (
              // Progress bar with stop button when sync is in progress
              <div className="flex items-center px-4 py-2 space-x-3">
                <div className="flex-1">
                  <div className="text-xs text-gray-600 dark:text-slate-400 mb-1">
                    Syncing... {syncProgress}%
                  </div>
                  {/* Progress bar */}
                  <div className="w-full h-2 bg-gray-200 dark:bg-slate-600 rounded-full overflow-hidden">
                    <div 
                      className="h-full bg-blue-500 transition-all duration-300 ease-out"
                      style={{ width: `${syncProgress}%` }}
                    />
                  </div>
                </div>
                
                {/* Stop button */}
                <button
                  type="button"
                  onClick={handleStopSync}
                  className="p-1 text-gray-500 hover:text-red-600 dark:text-slate-400 dark:hover:text-red-400 transition-colors"
                  title="Stop sync"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" className="w-4 h-4">
                    <circle cx="12" cy="12" r="10" fill="none" stroke="currentColor" strokeWidth="1.5" />
                    <rect x="8" y="8" width="8" height="8" fill="currentColor" />
                  </svg>
                </button>
              </div>
            ) : (
              // Normal action buttons when not in progress
              <>
                <button
                  onClick={handleSync}
                  className="flex items-center w-full px-4 py-2 text-sm text-gray-700 dark:text-slate-300 hover:bg-gray-100 dark:hover:bg-slate-700 transition-colors"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor" className="w-4 h-4 mr-3">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
                  </svg>
                  Sync
                </button>
                
                <button
                  onClick={handleCreateAlbum}
                  className="flex items-center w-full px-4 py-2 text-sm text-gray-700 dark:text-slate-300 hover:bg-gray-100 dark:hover:bg-slate-700 transition-colors"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor" className="w-4 h-4 mr-3">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M12 10.5v6m3-3H9m4.06-7.19-2.12-2.12a1.5 1.5 0 0 0-1.061-.44H4.5A2.25 2.25 0 0 0 2.25 6v12a2.25 2.25 0 0 0 2.25 2.25h15A2.25 2.25 0 0 0 21.75 18V9a2.25 2.25 0 0 0-2.25-2.25h-5.379a1.5 1.5 0 0 1-1.06-.44Z" />
                  </svg>
                  Create Album
                </button>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default ActionMenu;