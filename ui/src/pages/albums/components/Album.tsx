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

  const handleSetThumbnail = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (onSetThumbnail) {
      onSetThumbnail(album.id);
    }
  };

  const subAlbumsCount = album.children?.length || 0;

  return (
    <div
      className={`album-card-container ${
        isSelectionMode && isSelected ? 'selected' : ''
      }`}
      onClick={handleClick}
    >
      {/* Stack effect layers */}
      <div className="stack-effect-layer layer-1"></div>
      <div className="stack-effect-layer layer-2"></div>

      {/* Main card content */}
      <div className={`album-card-content ${
        isSelectionMode
          ? isSelected
            ? 'ring-2 ring-blue-500 ring-opacity-75'
            : ''
          : ''
      }`}>
        {/* Image wrapper */}
        <div className="image-wrapper">
          {album.thumbnail ? (
            <img
              src={album.thumbnail}
              alt={album.name}
              loading="lazy"
            />
          ) : (
            <div className="placeholder-image">
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

        {/* Selection checkbox overlay */}
        {isSelectionMode && (
          <div className="absolute top-2 right-2 z-10">
            <div
              className={`flex items-center justify-center w-6 h-6 bg-white/90 dark:bg-gray-700/90 border-2 rounded-md cursor-pointer transition-colors ${
                isSelected
                  ? 'border-blue-500 bg-blue-500'
                  : 'border-gray-300 dark:border-gray-500 hover:border-blue-500 dark:hover:border-blue-400'
              }`}
              onClick={handleCheckboxClick}
            >
              {isSelected && (
                <svg
                  className="w-4 h-4 text-white"
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

        {/* Set thumbnail button - top right when not in selection mode */}
        {!isSelectionMode && ((album.mediaCount || 0) > 0 || subAlbumsCount > 0) && (
          <button
            onClick={handleSetThumbnail}
            className="absolute top-2 right-2 z-10 p-1.5 bg-black/50 hover:bg-black/70 text-white rounded-md transition-all opacity-0 group-hover:opacity-100"
            title={album.thumbnail ? "Change thumbnail" : (album.mediaCount || 0) > 0 ? "Set thumbnail from photos in this album" : "Set thumbnail by navigating through sub-albums"}
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
              />
            </svg>
          </button>
        )}

        {/* Text overlay with gradient */}
        <div className="card-overlay">
          <div className="card-info">
            <h3 className="album-title">{album.name}</h3>
            <div className="album-meta">
              <span className="count">{album.mediaCount || 0} items</span>
              {subAlbumsCount > 0 && (
                <span className="sub-badge">
                  +{subAlbumsCount} folders
                </span>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Album;
