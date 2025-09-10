import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Album as AlbumType } from '@shared/types/Album';

export interface AlbumProps {
  album: AlbumType;
  isSelectionMode?: boolean;
  isSelected?: boolean;
  onSelectionToggle?: (albumId: string) => void;
  onSetThumbnail?: (albumId: string) => void;
}

const Album: React.FC<AlbumProps> = ({ 
  album, 
  isSelectionMode = false, 
  isSelected = false, 
  onSelectionToggle,
  onSetThumbnail 
}) => {
  const navigate = useNavigate();

  const handleClick = () => {
    if (isSelectionMode && onSelectionToggle) {
      onSelectionToggle(album.id);
    } else {
      navigate(`/albums/${album.id}`);
    }
  };

  const handleCheckboxClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (onSelectionToggle) {
      onSelectionToggle(album.id);
    }
  };

  const handleSetThumbnail = () => {
    if (onSetThumbnail) {
      onSetThumbnail(album.id);
    }
  };

  return (
    <div
      className={`relative aspect-square bg-white border rounded-lg shadow-sm hover:shadow-lg transition-all duration-300 dark:bg-gray-800 cursor-pointer flex flex-col ${
        isSelectionMode 
          ? isSelected
            ? 'border-blue-500 border-2 ring-2 ring-blue-500 ring-opacity-50' 
            : 'border-gray-200 dark:border-gray-700 hover:border-blue-300 dark:hover:border-blue-600'
          : 'border-gray-200 dark:border-gray-700 dark:hover:border-b-sky-500 dark:hover:border-b-2'
      }`}
      onClick={handleClick}
    >
      <div className="relative flex-shrink-0" style={{ height: '70%' }}>
        <div className="relative h-full bg-gray-100 dark:bg-gray-700 rounded-t-lg overflow-hidden">
          {album.thumbnail ? (
            <div className="relative w-full h-full">
              <img
                src={album.thumbnail}
                alt={album.name}
                className="w-full h-full object-cover hover:scale-105 transition-transform duration-300"
              />

            </div>
          ) : (
            // Stub image placeholder
            <div className="w-full h-full flex flex-col items-center justify-center bg-gray-200 dark:bg-gray-600">
              <svg
                className="w-12 h-12 text-gray-400 dark:text-gray-500"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
                xmlns="http://www.w3.org/2000/svg"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="1.5"
                  d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                />
              </svg>
            </div>
          )}
          
          {/* Sync spinner overlay */}
          {album.syncInProgress && (
            <div className="absolute top-2 left-2">
              <div 
                className="flex items-center justify-center w-6 h-6 bg-blue-500 bg-opacity-90 rounded-full"
                title="Sync job in progress"
              >
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
              </div>
            </div>
          )}

          {/* Selection checkbox overlay */}
          {isSelectionMode && (
            <div className="absolute top-2 right-2">
              <div
                className="flex items-center justify-center w-6 h-6 bg-white dark:bg-gray-700 border-2 border-gray-300 dark:border-gray-500 rounded-md cursor-pointer hover:border-blue-500 dark:hover:border-blue-400 transition-colors"
                onClick={handleCheckboxClick}
              >
                {isSelected && (
                  <svg
                    className="w-4 h-4 text-blue-600 dark:text-blue-400"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                      clipRule="evenodd"
                    />
                  </svg>
                )}
              </div>
            </div>
          )}
        </div>
      </div>

      <div className="p-3 flex-1 flex flex-col justify-between">
        <div className="flex-1">
          <h5 className="mb-1 text-lg font-semibold tracking-tight text-gray-900 dark:text-white hover:text-blue-600 dark:hover:text-blue-400 transition-colors line-clamp-1">
            {album.name}
          </h5>

          {/* Fixed height container for description to ensure consistent card heights */}
          <div className="h-6 mb-2">
            {album.description && (
              <p className="text-base text-gray-700 dark:text-gray-400 line-clamp-1">{album.description}</p>
            )}
          </div>
        </div>

        <div className="flex items-center justify-between gap-3 text-sm text-gray-500 dark:text-gray-400">
          <div className="flex items-center gap-3">
            <div className="flex items-center flex-shrink-0 min-w-0">
              <svg className="w-4 h-4 mr-1 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                />
              </svg>
              <span className="whitespace-nowrap text-xs">{album.mediaCount || 0}</span>
            </div>

            {(album.children?.length || 0) > 0 && (
              <div className="flex items-center flex-shrink-0 min-w-0">
                <svg className="w-4 h-4 mr-1 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2-2V7zm0 0V5a2 2 0 012-2h6l2 2h6a2 2 0 012 2v2M7 13h10"
                  />
                </svg>
                <span className="whitespace-nowrap text-xs">{album.children?.length || 0}</span>
              </div>
            )}
          </div>

          {/* Set/Change Thumbnail button - positioned at bottom right */}
          {((album.mediaCount || 0) > 0 || (album.children?.length || 0) > 0) && !isSelectionMode && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                handleSetThumbnail();
              }}
              className="p-1.5 bg-white dark:bg-gray-800 text-gray-500 dark:text-gray-400 border border-gray-200 dark:border-gray-700 rounded hover:text-blue-600 dark:hover:text-blue-400 hover:border-blue-300 dark:hover:border-blue-600 transition-colors flex-shrink-0"
              title={album.thumbnail ? "Change thumbnail" : (album.mediaCount || 0) > 0 ? "Set thumbnail from photos in this album" : "Set thumbnail by navigating through sub-albums"}
            >
              <svg className="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
                />
              </svg>
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default Album;
