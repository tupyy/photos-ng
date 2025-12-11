import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Album as AlbumType } from '@shared/types/Album';

export interface SubAlbumCardProps {
  album: AlbumType;
}

/**
 * SubAlbumCard - A simplified album card for displaying sub-albums
 *
 * Features:
 * - Square thumbnail with rounded corners
 * - Album name displayed below (not overlaid)
 * - Clean, minimal design without stacked folder effect
 * - Hover effect with subtle scale
 */
const SubAlbumCard: React.FC<SubAlbumCardProps> = ({ album }) => {
  const navigate = useNavigate();

  const handleClick = () => {
    navigate(`/albums/${album.id}`);
  };

  return (
    <div
      className="flex flex-col cursor-pointer group"
      onClick={handleClick}
    >
      {/* Thumbnail Container */}
      <div className="relative w-32 h-32 sm:w-36 sm:h-36 rounded-lg overflow-hidden bg-gray-200 dark:bg-gray-700 transition-transform duration-200 group-hover:scale-105 group-hover:shadow-lg">
        {album.thumbnail ? (
          <img
            src={album.thumbnail}
            alt={album.name}
            className="w-full h-full object-cover"
            loading="lazy"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center">
            <svg
              className="w-10 h-10 text-gray-400 dark:text-gray-500"
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
      </div>

      {/* Album Name */}
      <p className="mt-2 text-sm font-medium text-gray-900 dark:text-white text-center truncate w-32 sm:w-36">
        {album.name}
      </p>
    </div>
  );
};

export default SubAlbumCard;
