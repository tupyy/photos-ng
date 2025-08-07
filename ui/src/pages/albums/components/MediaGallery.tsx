import React, { useState } from 'react';
import { Media } from '@generated/models';
import MediaThumbnail from '@app/shared/components/MediaThumbnail';
import ExifDrawer from '@app/shared/components/ExifDrawer';
import { MediaViewerModal, ConfirmDeleteModal, Alert } from '@app/shared/components';
import { useMediaApi, useAlbumsApi } from '@shared/hooks/useApi';

interface MediaGalleryProps {
  media: Media[];
  loading?: boolean;
  error?: string | null;
  albumName?: string;
  albumId?: string;
  total?: number;
  currentPage?: number;
  pageSize?: number;
  onPageChange?: (page: number) => void;
  onMediaDeleted?: () => void;
}

const MediaGallery: React.FC<MediaGalleryProps> = ({
  media,
  loading = false,
  error = null,
  albumName,
  albumId,
  total = 0,
  currentPage = 1,
  pageSize = 100,
  onPageChange,
  onMediaDeleted,
}) => {
  const [selectedMedia, setSelectedMedia] = useState<Media | null>(null);
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);

  // Media viewer modal state
  const [isViewerOpen, setIsViewerOpen] = useState(false);
  const [currentMediaIndex, setCurrentMediaIndex] = useState(0);

  // Multi-select state
  const [isSelectionMode, setIsSelectionMode] = useState(false);
  const [selectedMediaIds, setSelectedMediaIds] = useState<Set<string>>(new Set());
  const [isDeleting, setIsDeleting] = useState(false);
  const [isSettingThumbnail, setIsSettingThumbnail] = useState(false);

  // Modal and alert state
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [alert, setAlert] = useState<{
    type: 'success' | 'error' | 'warning' | 'info';
    title?: string;
    message: string;
    visible: boolean;
  }>({
    type: 'info',
    message: '',
    visible: false,
  });

  // API hooks
  const { deleteMedia } = useMediaApi();
  const { updateAlbum } = useAlbumsApi();

  // Alert helper functions
  const showAlert = (type: 'success' | 'error' | 'warning' | 'info', message: string, title?: string) => {
    setAlert({
      type,
      title,
      message,
      visible: true,
    });
  };

  const hideAlert = () => {
    setAlert(prev => ({ ...prev, visible: false }));
  };

  // Helper function to get the start of the week (Monday) for a given date
  const getWeekStart = (date: Date) => {
    const d = new Date(date);
    const day = d.getDay();
    const diff = d.getDate() - day + (day === 0 ? -6 : 1); // Adjust for Monday start
    d.setDate(diff);
    d.setHours(0, 0, 0, 0);
    return d;
  };

  // Format week range string
  const formatWeekRange = (weekStart: Date) => {
    const weekEnd = new Date(weekStart);
    weekEnd.setDate(weekStart.getDate() + 6);

    const startDay = weekStart.getDate();
    const endDay = weekEnd.getDate();
    const startMonth = weekStart.toLocaleDateString('en-US', { month: 'long' });
    const endMonth = weekEnd.toLocaleDateString('en-US', { month: 'long' });

    if (startMonth === endMonth) {
      return `${startDay} - ${endDay} ${startMonth}`;
    } else {
      return `${startDay} ${startMonth} - ${endDay} ${endMonth}`;
    }
  };

  // Group media by week intervals
  const groupedMedia = React.useMemo(() => {
    if (!media || media.length === 0) return [];

    // First sort all media by capturedAt (descending)
    const sortedMedia = [...media].sort((a, b) => {
      const capturedAtA = new Date(a.capturedAt).getTime();
      const capturedAtB = new Date(b.capturedAt).getTime();

      if (capturedAtA !== capturedAtB) {
        return capturedAtB - capturedAtA; // Descending order
      }

      // If capturedAt is the same, sort by filename (ascending)
      return a.filename.localeCompare(b.filename);
    });

    // Group by week
    const groups = new Map<string, Media[]>();

    sortedMedia.forEach((mediaItem) => {
      const capturedAt = new Date(mediaItem.capturedAt);
      const weekStart = getWeekStart(capturedAt);
      const weekKey = weekStart.toISOString();

      if (!groups.has(weekKey)) {
        groups.set(weekKey, []);
      }
      groups.get(weekKey)!.push(mediaItem);
    });

    // Convert to array and sort weeks by date (most recent first)
    return Array.from(groups.entries())
      .map(([weekKey, mediaItems]) => ({
        weekStart: new Date(weekKey),
        weekRange: formatWeekRange(new Date(weekKey)),
        media: mediaItems
      }))
      .sort((a, b) => b.weekStart.getTime() - a.weekStart.getTime());
  }, [media]);

  // Flatten grouped media for modal navigation
  const allMedia = React.useMemo(() => {
    return groupedMedia.flatMap(group => group.media);
  }, [groupedMedia]);

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
    const index = allMedia.findIndex(m => m.id === mediaItem.id);
    setCurrentMediaIndex(index);
    setIsViewerOpen(true);
  };

  const handleViewerClose = (currentMedia?: Media) => {
    setIsViewerOpen(false);
    
    // Scroll to the media that was being viewed when modal was closed
    if (currentMedia) {
      // Use setTimeout to ensure modal is fully closed before scrolling
      setTimeout(() => {
        const mediaElement = document.querySelector(`[data-media-id="${currentMedia.id}"]`);
        if (mediaElement) {
          mediaElement.scrollIntoView({
            behavior: 'smooth',
            block: 'center',
            inline: 'nearest'
          });
        }
      }, 100);
    }
  };

  const handleIndexChange = (index: number) => {
    setCurrentMediaIndex(index);
  };

  // Multi-select handlers
  const toggleSelectionMode = () => {
    setIsSelectionMode(!isSelectionMode);
    setSelectedMediaIds(new Set());
  };

  const toggleMediaSelection = (mediaId: string) => {
    const newSelected = new Set(selectedMediaIds);
    if (newSelected.has(mediaId)) {
      newSelected.delete(mediaId);
    } else {
      newSelected.add(mediaId);
    }
    setSelectedMediaIds(newSelected);
  };

  const selectAllMedia = () => {
    const allIds = new Set(allMedia.map(m => m.id));
    setSelectedMediaIds(allIds);
  };

  const clearSelection = () => {
    setSelectedMediaIds(new Set());
  };

  const handleDeleteSelected = () => {
    if (selectedMediaIds.size === 0) return;
    setShowDeleteModal(true);
  };

  const confirmDelete = async () => {
    setIsDeleting(true);
    setShowDeleteModal(false);

    try {
      // Delete media one by one
      const deletePromises = Array.from(selectedMediaIds).map(id => deleteMedia(id));
      await Promise.all(deletePromises);

      const deletedCount = selectedMediaIds.size;

      // Clear selection and exit selection mode
      setSelectedMediaIds(new Set());
      setIsSelectionMode(false);

      // Notify parent component to refresh album data
      // This is important in case any of the deleted media was used as the album thumbnail
      if (onMediaDeleted) {
        onMediaDeleted();
      }

      // Show success alert
      showAlert(
        'success',
        `Successfully deleted ${deletedCount} ${deletedCount === 1 ? 'photo' : 'photos'}.`,
        'Deletion Complete!'
      );

    } catch (error) {
      console.error('Failed to delete media:', error);
      showAlert(
        'error',
        'Failed to delete some photos. Please try again.',
        'Deletion Failed!'
      );
    } finally {
      setIsDeleting(false);
    }
  };

  const cancelDelete = () => {
    setShowDeleteModal(false);
  };

  const handleSetThumbnail = async () => {
    if (selectedMediaIds.size !== 1 || !albumId) return;

    const selectedMediaId = Array.from(selectedMediaIds)[0];

    setIsSettingThumbnail(true);

    try {
      await updateAlbum(albumId, { thumbnail: selectedMediaId });

      // Clear selection and exit selection mode
      setSelectedMediaIds(new Set());
      setIsSelectionMode(false);

      // Show success alert
      showAlert(
        'success',
        'The album thumbnail has been updated successfully.',
        'Thumbnail Updated!'
      );

    } catch (error) {
      console.error('Failed to set album thumbnail:', error);
      showAlert(
        'error',
        'Failed to set album thumbnail. Please try again.',
        'Update Failed!'
      );
    } finally {
      setIsSettingThumbnail(false);
    }
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

  if (!allMedia || allMedia.length === 0) {
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
      {/* Sticky Header with selection controls */}
      <div className="sticky top-0 z-30 bg-gray-50 dark:bg-slate-900 border-b border-gray-200 dark:border-gray-700 pb-4 mb-6 backdrop-blur-sm bg-opacity-95 dark:bg-opacity-95">
        <div className="flex items-center justify-between pt-4">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            Photos ({allMedia.length})
            {isSelectionMode && selectedMediaIds.size > 0 && (
              <span className="ml-2 text-sm text-blue-600 dark:text-blue-400">
                ({selectedMediaIds.size} selected)
              </span>
            )}
          </h2>

          <div className="flex items-center space-x-2">
            {isSelectionMode ? (
              <>
                <button
                  onClick={selectAllMedia}
                  disabled={isDeleting || isSettingThumbnail}
                  className="px-3 py-1 text-sm border border-gray-300 rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
                >
                  Select All
                </button>
                <button
                  onClick={clearSelection}
                  disabled={isDeleting || isSettingThumbnail}
                  className="px-3 py-1 text-sm border border-gray-300 rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
                >
                  Clear
                </button>
                {selectedMediaIds.size === 1 && albumId && (
                  <button
                    onClick={handleSetThumbnail}
                    disabled={isDeleting || isSettingThumbnail}
                    className="px-3 py-1 text-sm border border-green-300 rounded-md text-green-700 bg-green-50 hover:bg-green-100 focus:outline-none focus:ring-2 focus:ring-green-500 disabled:opacity-50 dark:bg-green-900/20 dark:border-green-600 dark:text-green-300 dark:hover:bg-green-900/30"
                  >
                    {isSettingThumbnail ? 'Setting...' : 'Set Album Thumbnail'}
                  </button>
                )}
                {selectedMediaIds.size > 0 && (
                  <button
                    onClick={handleDeleteSelected}
                    disabled={isDeleting || isSettingThumbnail}
                    className="px-3 py-1 text-sm border border-red-300 rounded-md text-red-700 bg-red-50 hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-red-500 disabled:opacity-50 dark:bg-red-900/20 dark:border-red-600 dark:text-red-300 dark:hover:bg-red-900/30"
                  >
                    {isDeleting ? 'Deleting...' : `Delete (${selectedMediaIds.size})`}
                  </button>
                )}
                <button
                  onClick={toggleSelectionMode}
                  disabled={isDeleting || isSettingThumbnail}
                  className="px-3 py-1 text-sm border border-gray-300 rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
                >
                  Cancel
                </button>
              </>
            ) : (
              <button
                onClick={toggleSelectionMode}
                className="px-3 py-1 text-sm border border-gray-300 rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
              >
                Select Photos
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Grouped by weeks */}
      <div className="space-y-8">
        {groupedMedia.map((group) => (
          <div key={group.weekStart.toISOString()}>
            {/* Week header */}
            <h3 className="text-md font-medium text-gray-700 dark:text-gray-300 mb-4 border-b border-gray-200 dark:border-gray-700 pb-2">
              {group.weekRange} ({group.media.length} photos)
            </h3>

            {/* Media grid for this week */}
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-4">
              {group.media.map((mediaItem) => (
                <MediaThumbnail
                  key={mediaItem.id}
                  media={mediaItem}
                  onInfoClick={handleInfoClick}
                  onClick={isSelectionMode ? () => toggleMediaSelection(mediaItem.id) : handleMediaClick}
                  isSelectionMode={isSelectionMode}
                  isSelected={selectedMediaIds.has(mediaItem.id)}
                />
              ))}
            </div>
          </div>
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
        media={allMedia}
        currentIndex={currentMediaIndex}
        onClose={handleViewerClose}
        onIndexChange={handleIndexChange}
      />

      {/* Delete Confirmation Modal */}
      <ConfirmDeleteModal
        isOpen={showDeleteModal}
        itemCount={selectedMediaIds.size}
        isDeleting={isDeleting}
        onConfirm={confirmDelete}
        onCancel={cancelDelete}
      />

      {/* Alert */}
      <Alert
        type={alert.type}
        title={alert.title}
        message={alert.message}
        isVisible={alert.visible}
        onDismiss={hideAlert}
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
