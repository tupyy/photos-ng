import React from 'react';
import { Media } from '@generated/models';

interface MediaThumbnailProps {
  media: Media;
  onInfoClick: (media: Media) => void;
  onClick?: (media: Media) => void;
  isSelectionMode?: boolean;
  isSelected?: boolean;
}

const MediaThumbnail: React.FC<MediaThumbnailProps> = ({ media, onInfoClick, onClick, isSelectionMode = false, isSelected = false }) => {
  const handleClick = () => {
    if (onClick) {
      onClick(media);
    }
  };

  const handleInfoClick = (e: React.MouseEvent) => {
    e.stopPropagation(); // Prevent triggering the main click
    onInfoClick(media);
  };

  const handleError = (e: React.SyntheticEvent<HTMLImageElement>) => {
    // Fallback to a placeholder when thumbnail fails to load
    e.currentTarget.src =
      'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjNmNGY2Ii8+CiAgPHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSIxNCIgZmlsbD0iIzk5YTNhZiIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPk5vIEltYWdlPC90ZXh0Pgo8L3N2Zz4K';
  };

  return (
    <div
      className={`relative aspect-square bg-gray-200 dark:bg-gray-700 rounded-lg overflow-hidden cursor-pointer transition-all duration-200 group ${
        isSelectionMode 
          ? isSelected 
            ? 'ring-4 ring-blue-500 ring-opacity-75' 
            : 'hover:ring-2 hover:ring-blue-300 hover:ring-opacity-50'
          : 'hover:opacity-80'
      }`}
      onClick={handleClick}
    >
      <img
        src={media.thumbnail}
        alt={`Media ${media.filename}`}
        className={`w-full h-full object-cover transition-transform duration-200 ${
          isSelectionMode ? '' : 'group-hover:scale-105'
        }`}
        loading="lazy"
        onError={handleError}
      />

      {/* Selection checkbox */}
      {isSelectionMode && (
        <div className="absolute top-2 left-2 w-6 h-6 bg-white rounded-full flex items-center justify-center shadow-lg">
          {isSelected ? (
            <svg className="w-4 h-4 text-blue-600" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
            </svg>
          ) : (
            <div className="w-4 h-4 border-2 border-gray-300 rounded-sm"></div>
          )}
        </div>
      )}

      {/* Selection overlay */}
      {isSelectionMode && isSelected && (
        <div className="absolute inset-0 bg-blue-500 bg-opacity-20"></div>
      )}

      {/* Info button */}
      {!isSelectionMode && (
        <button
          onClick={handleInfoClick}
          className="absolute top-2 right-2 w-6 h-6 bg-black bg-opacity-50 hover:bg-opacity-70 rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity duration-200"
          title="View EXIF data"
        >
          <svg className="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
            <path
              fillRule="evenodd"
              d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
              clipRule="evenodd"
            />
          </svg>
        </button>
      )}
    </div>
  );
};

export default MediaThumbnail;
