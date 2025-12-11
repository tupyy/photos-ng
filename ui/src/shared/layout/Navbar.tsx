import React, { useState, useEffect, useRef } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { useTheme } from '@shared/contexts';
import { useAppSelector, useAppDispatch, selectAlbumsPageActive, selectCurrentAlbum, selectUser, selectCanCreateAlbums } from '@shared/store';
import { setCreateFormOpen } from '@shared/reducers/albumsSlice';
import { HomeIcon, FolderIcon, SunIcon, MoonIcon, PlusIcon, ArrowUpTrayIcon, UserIcon } from '@heroicons/react/24/outline';

export interface NavbarProps {}

const Navbar: React.FC<NavbarProps> = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const { theme, toggleTheme } = useTheme();
  const isAlbumsPageActive = useAppSelector(selectAlbumsPageActive);
  const currentAlbum = useAppSelector(selectCurrentAlbum);
  const canCreateAlbums = useAppSelector(selectCanCreateAlbums);
  const user = useAppSelector(selectUser);
  const [profileMenuOpen, setProfileMenuOpen] = useState(false);
  const profileMenuRef = useRef<HTMLDivElement>(null);

  // Close dropdowns when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (profileMenuRef.current && !profileMenuRef.current.contains(event.target as Node)) {
        setProfileMenuOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

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
      icon: <FolderIcon className="w-6 h-6" />,
    },
  ];

  return (
    <div className="sticky top-0 z-30 w-full backdrop-blur flex-none transition-colors duration-500 lg:z-30 supports-backdrop-blur:bg-white/60 dark:bg-slate-900">
      <nav>
        <div className="flex flex-wrap items-center justify-between mx-auto p-4 lg:border-b lg:border-slate-900/10">
          {/* Left - Brand and Navigation Icons */}
          <div className="flex items-center space-x-6">
            <div className="flex items-center">
              <Link to="/" className="p-2 text-gray-500 hover:text-gray-700 dark:text-slate-400 dark:hover:text-gray-200 transition-colors" title="Photos NG">
                <HomeIcon className="w-6 h-6" />
              </Link>
            </div>
            
            <div className="flex dark:border-slate-50/[0.06] items-center space-x-3">
              {navItems.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  className={`p-2 transition-colors ${
                    location.pathname.startsWith(item.path)
                      ? 'text-gray-700 dark:text-gray-200'
                      : 'text-gray-500 hover:text-gray-700 dark:text-slate-400 dark:hover:text-gray-200'
                  }`}
                  title={item.label}
                >
                  {item.icon}
                </Link>
              ))}
            </div>
          </div>

          {/* Right side - Create Album, Build Info, Theme Toggle, User Profile */}
          <div className="flex items-center md:order-2 space-x-3 md:space-x-2">
            {/* Create Album Button - Only show when albums page is active and user can create albums */}
            {isAlbumsPageActive && canCreateAlbums && (
              <button
                type="button"
                onClick={handleCreateAlbumFormOpen}
                className="p-2 text-gray-500 hover:text-gray-700 dark:text-slate-400 dark:hover:text-gray-200 transition-colors"
                title="Create Album"
              >
                <PlusIcon className="w-6 h-6" />
              </button>
            )}

            {/* Upload Media Button - Only show when inside an album */}
            {isAlbumsPageActive && currentAlbum && (
              <button
                type="button"
                onClick={handleUploadMedia}
                className="p-2 text-gray-500 hover:text-gray-700 dark:text-slate-400 dark:hover:text-gray-200 transition-colors"
                title="Upload Media"
              >
                <ArrowUpTrayIcon className="w-6 h-6" />
              </button>
            )}

            {/* Theme Toggle Button */}
            <button
              type="button"
              onClick={toggleTheme}
              className="p-2 text-gray-500 hover:text-gray-700 dark:text-slate-400 dark:hover:text-gray-200 transition-colors"
              title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
            >
              {theme === 'dark' ? (
                <SunIcon className="w-6 h-6" />
              ) : (
                <MoonIcon className="w-6 h-6" />
              )}
            </button>

            {/* User Profile Dropdown */}
            {user && (
              <div className="relative" ref={profileMenuRef}>
                <button
                  type="button"
                  onClick={() => setProfileMenuOpen(!profileMenuOpen)}
                  className="flex items-center p-2 text-gray-500 hover:text-gray-700 dark:text-slate-400 dark:hover:text-gray-200 transition-colors"
                  title="User profile"
                >
                  <UserIcon className="w-6 h-6" />
                </button>

                {profileMenuOpen && (
                  <div className="absolute right-0 mt-2 w-60 bg-white dark:bg-slate-800 rounded-lg shadow-lg z-50">
                    <div className="p-4 border-b border-gray-200 dark:border-slate-700">
                      <p className="text-sm font-medium text-gray-900 dark:text-white">{user.name}</p>
                      <p className="text-xs text-gray-500 dark:text-slate-400">{user.user}</p>
                    </div>
                    <div className="p-3">
                      <p className="text-xs font-medium text-gray-500 dark:text-slate-400 uppercase mb-2">Permissions</p>
                      <div className="space-y-1.5">
                        <div className="flex items-center justify-between text-sm">
                          <span className="text-gray-700 dark:text-slate-300">Create Albums</span>
                          <span className={`px-2 py-0.5 rounded text-xs font-medium ${
                            canCreateAlbums
                              ? 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400'
                              : 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400'
                          }`}>
                            {canCreateAlbums ? 'allowed' : 'denied'}
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </nav>
    </div>
  );
};

export default Navbar;
