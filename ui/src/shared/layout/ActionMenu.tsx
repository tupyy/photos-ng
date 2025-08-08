import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAppDispatch, useAppSelector, selectSyncStarting } from '@shared/store';
import { startSyncJob } from '@shared/reducers/syncSlice';

export interface ActionMenuProps {
  isOpen: boolean;
  onToggle: () => void;
  onClose: () => void;
  onCreateAlbum: () => void;
  onUploadMedia?: () => void;
  showUploadMedia?: boolean;
  albumPath?: string; // Current album path for sync operations
}

const ActionMenu: React.FC<ActionMenuProps> = ({ 
  isOpen, 
  onToggle, 
  onClose, 
  onCreateAlbum, 
  onUploadMedia, 
  showUploadMedia = false,
  albumPath = ''
}) => {
  const navigate = useNavigate();
  const location = useLocation();
  const dispatch = useAppDispatch();
  const isStartingSync = useAppSelector(selectSyncStarting);

  const handleSync = async () => {
    onClose();
    
    // Determine sync path based on current location
    const isRoot = location.pathname === '/' || location.pathname === '/albums';
    const syncPath = isRoot ? '' : albumPath;
    
    try {
      // Dispatch Redux action to start sync job
      const result = await dispatch(startSyncJob(syncPath)).unwrap();
      
      // Navigate to sync page showing the new job
      navigate(`/sync?jobId=${result.jobId}`);
    } catch (error) {
      console.error('Failed to start sync job:', error);
      // On error, navigate to sync page to show error or allow manual retry
      navigate('/sync');
    }
  };

  const handleCreateAlbum = () => {
    onClose();
    onCreateAlbum();
  };

  const handleUploadMedia = () => {
    onClose();
    if (onUploadMedia) {
      onUploadMedia();
    }
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
            <button
              onClick={handleSync}
              disabled={isStartingSync}
              className="flex items-center w-full px-4 py-2 text-sm text-gray-700 dark:text-slate-300 hover:bg-gray-100 dark:hover:bg-slate-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isStartingSync ? (
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-current mr-3"></div>
              ) : (
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor" className="w-4 h-4 mr-3">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
                </svg>
              )}
              {isStartingSync ? 'Starting...' : 'Sync'}
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
            
            {showUploadMedia && onUploadMedia && (
              <button
                onClick={handleUploadMedia}
                className="flex items-center w-full px-4 py-2 text-sm text-gray-700 dark:text-slate-300 hover:bg-gray-100 dark:hover:bg-slate-700 transition-colors"
              >
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor" className="w-4 h-4 mr-3">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5m-13.5-9L12 3m0 0 4.5 4.5M12 3v13.5" />
                </svg>
                Upload Media
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default ActionMenu;