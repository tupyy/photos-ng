import React, { useState } from 'react';
import { Media } from '@generated/models';
import ExifDrawer from './ExifDrawer';

interface MediaGalleryProps {
  media: Media[];
  loading?: boolean;
  error?: string | null;
  albumName?: string;
}

const MediaGallery: React.FC<MediaGalleryProps> = ({ media, loading = false, error = null, albumName }) => {
  const [selectedMedia, setSelectedMedia] = useState<Media | null>(null);
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);

  // Sort media locally: maintain API sorting by capturedAt, but add secondary sort by filename for same timestamps
  const sortedMedia = React.useMemo(() => {
    if (!media || media.length === 0) return media;
    
    return [...media].sort((a, b) => {
      // First, sort by capturedAt (descending, as API does)
      const capturedAtA = new Date(a.capturedAt).getTime();
      const capturedAtB = new Date(b.capturedAt).getTime();
      
      if (capturedAtA !== capturedAtB) {
        return capturedAtB - capturedAtA; // Descending order
      }
      
      // If capturedAt is the same, sort by filename (ascending)
      return a.filename.localeCompare(b.filename);
    });
  }, [media]);

  const handleInfoClick = (mediaItem: Media) => {
    setSelectedMedia(mediaItem);
    setIsDrawerOpen(true);
  };

  const handleCloseDrawer = () => {
    setIsDrawerOpen(false);
    setSelectedMedia(null);
  };
  if (loading) {
    return (
      <div className="mt-8">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Photos</h2>
        <div className="flex justify-center items-center py-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="mt-8">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Photos</h2>
        <div className="text-center py-8">
          <p className="text-red-600 dark:text-red-400">{error}</p>
        </div>
      </div>
    );
  }

  if (!sortedMedia || sortedMedia.length === 0) {
    return (
      <div className="mt-8">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Photos</h2>
        <div className="text-center py-8">
          <svg
            className="mx-auto h-12 w-12 text-gray-400"
            stroke="currentColor"
            fill="none"
            viewBox="0 0 48 48"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="2"
              d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6 6h.01M6 20h.01m-.01 4h.01m4.01 0h.01m0 4h.01M10 28h.01m4.01 0h.01m0 4h.01M16 32h.01M8 36a4 4 0 004 4h8a4 4 0 004-4v-8a4 4 0 00-4-4h-8a4 4 0 00-4 4v8z"
            />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-white">No photos</h3>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {albumName ? `${albumName} doesn't contain any photos yet.` : 'This album doesn\'t contain any photos yet.'}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="mt-8">
      <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
        Photos ({sortedMedia.length})
      </h2>
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-4">
        {sortedMedia.map((mediaItem) => (
          <MediaThumbnail key={mediaItem.id} media={mediaItem} onInfoClick={handleInfoClick} />
        ))}
      </div>

      {/* EXIF Drawer */}
      <ExifDrawer 
        isOpen={isDrawerOpen} 
        media={selectedMedia} 
        onClose={handleCloseDrawer} 
      />
    </div>
  );
};

interface MediaThumbnailProps {
  media: Media;
  onInfoClick: (media: Media) => void;
}

const MediaThumbnail: React.FC<MediaThumbnailProps> = ({ media, onInfoClick }) => {
  const handleClick = () => {
    // TODO: Implement modal or lightbox for viewing full image
    console.log('Media clicked:', media);
  };

  const handleInfoClick = (e: React.MouseEvent) => {
    e.stopPropagation(); // Prevent triggering the main click
    onInfoClick(media);
  };

  const handleError = (e: React.SyntheticEvent<HTMLImageElement>) => {
    // Fallback to a placeholder when thumbnail fails to load
    e.currentTarget.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjNmNGY2Ii8+CiAgPHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSIxNCIgZmlsbD0iIzk5YTNhZiIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPk5vIEltYWdlPC90ZXh0Pgo8L3N2Zz4K';
  };

  return (
    <div
      className="relative aspect-square bg-gray-200 dark:bg-gray-700 rounded-lg overflow-hidden cursor-pointer hover:opacity-80 transition-opacity group"
      onClick={handleClick}
    >
      <img
        src={media.thumbnail}
        alt={`Media ${media.filename}`}
        className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-200"
        loading="lazy"
        onError={handleError}
      />
      
      {/* Info button */}
      <button
        onClick={handleInfoClick}
        className="absolute top-2 right-2 w-6 h-6 bg-black bg-opacity-50 hover:bg-opacity-70 rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity duration-200"
        title="View EXIF data"
      >
        <svg
          className="w-3 h-3 text-white"
          fill="currentColor"
          viewBox="0 0 20 20"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            fillRule="evenodd"
            d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
            clipRule="evenodd"
          />
        </svg>
      </button>
    </div>
  );
};

export default MediaGallery;