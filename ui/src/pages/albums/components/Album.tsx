import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Album as AlbumType } from '@shared/types/Album';

export interface AlbumProps {
  album: AlbumType;
}

const Album: React.FC<AlbumProps> = ({ album }) => {
  const navigate = useNavigate();

  const handleClick = () => {
    navigate(`/albums/${album.id}`);
  };

  return (
    <div
      className="relative aspect-square bg-white border border-gray-200 rounded-lg shadow-sm hover:shadow-lg transition-shadow duration-300 dark:bg-gray-800 dark:border-gray-700 cursor-pointer flex flex-col dark:hover:border-b-sky-500 dark:hover:border-b-2"
      onClick={handleClick}
    >
      <div className="relative flex-shrink-0" style={{ height: '70%' }}>
        <div className="relative h-full bg-gray-100 dark:bg-gray-700 rounded-t-lg overflow-hidden">
          {album.thumbnail ? (
            <img
              src={album.thumbnail}
              alt={album.name}
              className="w-full h-full object-cover hover:scale-105 transition-transform duration-300"
            />
          ) : (
            // Stub image placeholder
            <div className="w-full h-full flex items-center justify-center bg-gray-200 dark:bg-gray-600">
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

        <div className="flex items-center justify-between text-sm text-gray-500 dark:text-gray-400">
          <div className="flex items-center">
            <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
              />
            </svg>
            <span>{album.media?.length || 0} photos</span>
          </div>

          {(album.children?.length || 0) > 0 && (
            <div className="flex items-center">
              <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2-2V7zm0 0V5a2 2 0 012-2h6l2 2h6a2 2 0 012 2v2M7 13h10"
                />
              </svg>
              <span>{album.children?.length || 0} albums</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Album;
