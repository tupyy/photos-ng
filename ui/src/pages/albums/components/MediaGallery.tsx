import React, { useState } from 'react';
import { Media } from '@generated/models';
import MediaThumbnail from '@app/shared/components/MediaThumbnail';
import ExifDrawer from '@app/shared/components/ExifDrawer';
import { MediaViewerModal } from '@app/shared/components';

interface MediaGalleryProps {
  media: Media[];
  loading?: boolean;
  error?: string | null;
  albumName?: string;
  total?: number;
  currentPage?: number;
  pageSize?: number;
  onPageChange?: (page: number) => void;
}

const MediaGallery: React.FC<MediaGalleryProps> = ({
  media,
  loading = false,
  error = null,
  albumName,
  total = 0,
  currentPage = 1,
  pageSize = 100,
  onPageChange,
}) => {
  const [selectedMedia, setSelectedMedia] = useState<Media | null>(null);
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);
  
  // Media viewer modal state
  const [isViewerOpen, setIsViewerOpen] = useState(false);
  const [currentMediaIndex, setCurrentMediaIndex] = useState(0);

  // Sort media locally: maintain API global sorting by capturedAt (desc), but add secondary sort by filename for items with same timestamp within this page
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

  // Media viewer modal handlers
  const handleMediaClick = (mediaItem: Media) => {
    const index = sortedMedia.findIndex(m => m.id === mediaItem.id);
    setCurrentMediaIndex(index);
    setIsViewerOpen(true);
  };

  const handleViewerClose = () => {
    setIsViewerOpen(false);
  };

  const handleIndexChange = (index: number) => {
    setCurrentMediaIndex(index);
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
          <svg className="mx-auto h-12 w-12 text-gray-400" stroke="currentColor" fill="none" viewBox="0 0 48 48">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="2"
              d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6 6h.01M6 20h.01m-.01 4h.01m4.01 0h.01m0 4h.01M10 28h.01m4.01 0h.01m0 4h.01M16 32h.01M8 36a4 4 0 004 4h8a4 4 0 004-4v-8a4 4 0 00-4-4h-8a4 4 0 00-4 4v8z"
            />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-white">No photos</h3>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {albumName ? `${albumName} doesn't contain any photos yet.` : "This album doesn't contain any photos yet."}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="mt-8">
      <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Photos ({sortedMedia.length})</h2>
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-4">
        {sortedMedia.map((mediaItem) => (
          <MediaThumbnail 
            key={mediaItem.id} 
            media={mediaItem} 
            onInfoClick={handleInfoClick}
            onClick={handleMediaClick}
          />
        ))}
      </div>

      {/* Pagination */}
      {total > pageSize && onPageChange && (
        <Pagination currentPage={currentPage} totalItems={total} pageSize={pageSize} onPageChange={onPageChange} />
      )}

      {/* EXIF Drawer */}
      <ExifDrawer isOpen={isDrawerOpen} media={selectedMedia} onClose={handleCloseDrawer} />

      {/* Media Viewer Modal */}
      <MediaViewerModal
        isOpen={isViewerOpen}
        media={sortedMedia}
        currentIndex={currentMediaIndex}
        onClose={handleViewerClose}
        onIndexChange={handleIndexChange}
      />
    </div>
  );
};

// Pagination component
interface PaginationProps {
  currentPage: number;
  totalItems: number;
  pageSize: number;
  onPageChange: (page: number) => void;
}

const Pagination: React.FC<PaginationProps> = ({ currentPage, totalItems, pageSize, onPageChange }) => {
  const totalPages = Math.ceil(totalItems / pageSize);
  const startItem = (currentPage - 1) * pageSize + 1;
  const endItem = Math.min(currentPage * pageSize, totalItems);

  if (totalPages <= 1) return null;

  const generatePageNumbers = () => {
    const pages: (number | string)[] = [];
    const showPages = 5; // Show 5 pages at most

    if (totalPages <= showPages) {
      // Show all pages if total is small
      for (let i = 1; i <= totalPages; i++) {
        pages.push(i);
      }
    } else {
      // Always show first page
      pages.push(1);

      if (currentPage > 3) {
        pages.push('...');
      }

      // Show pages around current page
      const start = Math.max(2, currentPage - 1);
      const end = Math.min(totalPages - 1, currentPage + 1);

      for (let i = start; i <= end; i++) {
        if (i !== 1 && i !== totalPages) {
          pages.push(i);
        }
      }

      if (currentPage < totalPages - 2) {
        pages.push('...');
      }

      // Always show last page
      if (totalPages > 1) {
        pages.push(totalPages);
      }
    }

    return pages;
  };

  return (
    <div className="flex items-center justify-between border-t border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-slate-900 px-4 py-3 sm:px-6 mt-6">
      <div className="flex flex-1 justify-between sm:hidden">
        {/* Mobile pagination */}
        <button
          onClick={() => onPageChange(currentPage - 1)}
          disabled={currentPage === 1}
          className="relative inline-flex items-center rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Previous
        </button>
        <button
          onClick={() => onPageChange(currentPage + 1)}
          disabled={currentPage === totalPages}
          className="relative ml-3 inline-flex items-center rounded-md border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-600 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Next
        </button>
      </div>
      <div className="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
        <div>
          <p className="text-sm text-gray-700 dark:text-gray-300">
            Showing <span className="font-medium">{startItem}</span> to <span className="font-medium">{endItem}</span>{' '}
            of <span className="font-medium">{totalItems}</span> results
          </p>
        </div>
        <div>
          <nav className="isolate inline-flex -space-x-px rounded-md shadow-sm" aria-label="Pagination">
            {/* Previous button */}
            <button
              onClick={() => onPageChange(currentPage - 1)}
              disabled={currentPage === 1}
              className="relative inline-flex items-center rounded-l-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 dark:ring-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700 focus:z-20 focus:outline-offset-0 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <svg className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                <path
                  fillRule="evenodd"
                  d="M12.79 5.23a.75.75 0 01-.02 1.06L8.832 10l3.938 3.71a.75.75 0 11-1.04 1.08l-4.5-4.25a.75.75 0 010-1.08l4.5-4.25a.75.75 0 011.06.02z"
                  clipRule="evenodd"
                />
              </svg>
            </button>

            {/* Page numbers */}
            {generatePageNumbers().map((page, index) => (
              <React.Fragment key={index}>
                {page === '...' ? (
                  <span className="relative inline-flex items-center px-4 py-2 text-sm font-semibold text-gray-700 dark:text-gray-300 ring-1 ring-inset ring-gray-300 dark:ring-gray-600 focus:outline-offset-0">
                    ...
                  </span>
                ) : (
                  <button
                    onClick={() => onPageChange(page as number)}
                    className={`relative inline-flex items-center px-4 py-2 text-sm font-semibold ring-1 ring-inset ring-gray-300 dark:ring-gray-600 focus:z-20 focus:outline-offset-0 ${
                      currentPage === page
                        ? 'z-10 bg-blue-800 text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600'
                        : 'text-gray-900 dark:text-gray-100 hover:bg-gray-50 dark:hover:bg-gray-700'
                    }`}
                  >
                    {page}
                  </button>
                )}
              </React.Fragment>
            ))}

            {/* Next button */}
            <button
              onClick={() => onPageChange(currentPage + 1)}
              disabled={currentPage === totalPages}
              className="relative inline-flex items-center rounded-r-md px-2 py-2 text-gray-400 ring-1 ring-inset ring-gray-300 dark:ring-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700 focus:z-20 focus:outline-offset-0 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <svg className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                <path
                  fillRule="evenodd"
                  d="M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z"
                  clipRule="evenodd"
                />
              </svg>
            </button>
          </nav>
        </div>
      </div>
    </div>
  );
};

export default MediaGallery;
