import React, { useState, useEffect, useRef } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useTheme } from '@shared/contexts';
import { useAppSelector, useAppDispatch, selectAlbumsPageActive, selectCurrentAlbum } from '@shared/store';
import { setCreateFormOpen } from '@shared/reducers/albumsSlice';
import { useSyncPolling } from '@shared/hooks/useSyncPolling';
import { BuildInfo } from '@shared/components';
import ActionMenu from './ActionMenu';

export interface NavbarProps {}

const Navbar: React.FC<NavbarProps> = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const { theme, toggleTheme } = useTheme();
  const isAlbumsPageActive = useAppSelector(selectAlbumsPageActive);
  const currentAlbum = useAppSelector(selectCurrentAlbum);
  const [actionMenuOpen, setActionMenuOpen] = useState(false);
  const actionMenuRef = useRef<HTMLDivElement>(null);
  
  // Use centralized sync polling for navbar status
  const { hasActiveJobs } = useSyncPolling({
    interval: 2000, // Poll every 2 seconds for navbar
  });

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (actionMenuRef.current && !actionMenuRef.current.contains(event.target as Node)) {
        setActionMenuOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  // Polling is now handled by the useSyncPolling hook

  const handleCreateAlbumFormOpen = () => {
    dispatch(setCreateFormOpen(true));
  };

  const handleUploadMedia = () => {
    if (currentAlbum) {
      navigate(`/upload/${currentAlbum.id}`);
    }
  };

  const navItems = [
    {
      path: '/albums',
      label: 'Albums',
      icon: (
        <svg className="w-6 h-6" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="none" viewBox="0 0 24 24">
          <path stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13.5 8H4m0-2v13a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1V9a1 1 0 0 0-1-1h-5.032a1 1 0 0 1-.768-.36l-1.9-2.28a1 1 0 0 0-.768-.36H5a1 1 0 0 0-1 1Z" />
        </svg>
      )
    },
    {
      path: '/sync',
      label: 'Sync',
      icon: (
        <svg className={`w-6 h-6 ${hasActiveJobs ? 'animate-spin' : ''}`} xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99" />
        </svg>
      )
    }
  ];

  return (
    <div className="sticky top-0 z-30 w-full backdrop-blur flex-none transition-colors duration-500 lg:z-30 supports-backdrop-blur:bg-white/60 dark:bg-slate-900">
      <nav>
        <div className="flex flex-wrap items-center justify-between mx-auto p-4 lg:border-b lg:border-slate-900/10">
          {/* Left - Brand and Navigation Icons */}
          <div className="flex items-center space-x-6">
            <div className="flex items-center">
              <Link to="/" className="flex items-center text-gray-900 dark:text-slate-400 hover:text-blue-600 dark:hover:text-blue-400 transition-colors" title="Photos NG">
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"></path>
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z"></path>
                </svg>
              </Link>
            </div>
            
            <div className="flex dark:border-slate-50/[0.06] items-center space-x-3">
              {navItems.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  className={`flex items-center p-2 rounded-lg transition-colors ${
                    location.pathname === item.path
                      ? 'text-blue-600 bg-blue-50 dark:bg-blue-900/20'
                      : 'text-gray-700 hover:text-blue-600 hover:bg-gray-50 dark:text-slate-400 dark:hover:text-blue-400 dark:hover:bg-gray-800'
                  }`}
                  title={item.label}
                >
                  <span className="flex-shrink-0">
                    {item.icon}
                  </span>
                </Link>
              ))}
            </div>
          </div>

          {/* Right side - Build Info, Action Menu and Theme Toggle */}
          <div className="flex items-center md:order-2 space-x-3 md:space-x-2">
            {/* Build Info */}
            <BuildInfo />
            
            {/* Action Menu - Only show when albums page is active */}
            {isAlbumsPageActive && (
              <div className="flex items-center space-x-1" ref={actionMenuRef}>
                <ActionMenu 
                  isOpen={actionMenuOpen}
                  onToggle={() => setActionMenuOpen(!actionMenuOpen)}
                  onClose={() => setActionMenuOpen(false)}
                  onCreateAlbum={handleCreateAlbumFormOpen}
                  onUploadMedia={handleUploadMedia}
                  showUploadMedia={!!currentAlbum}
                  albumPath={currentAlbum?.path || ''}
                />
              </div>
            )}

            {/* Theme Toggle Button */}
            <button
              type="button"
              onClick={toggleTheme}
              className="p-2 text-gray-500 hover:text-gray-700 dark:text-slate-400 dark:hover:text-gray-200 transition-colors"
              title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
            >
              {theme === 'dark' ? (
                // Sun icon for switching to light mode
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor" className="w-5 h-5">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 3v2.25m6.364.386-1.591 1.591M21 12h-2.25m-.386 6.364-1.591-1.591M12 18.75V21m-4.773-4.227-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0Z" />
                </svg>
              ) : (
                // Moon icon for switching to dark mode
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor" className="w-5 h-5">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M21.752 15.002A9.72 9.72 0 0 1 18 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 0 0 3 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 0 0 9.002-5.998Z" />
                </svg>
              )}
            </button>
          </div>
        </div>
      </nav>
    </div>
  );
};

export default Navbar;
