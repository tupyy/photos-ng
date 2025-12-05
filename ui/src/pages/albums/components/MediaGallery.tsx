import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Media } from '@generated/models';
import MediaThumbnail from '@app/shared/components/MediaThumbnail';
import ExifDrawer from '@app/shared/components/ExifDrawer';
import { MediaViewerModal, ConfirmDeleteModal, Alert, LoadingProgressBar } from '@app/shared/components';
import { useMediaApi, useAlbumsApi } from '@shared/hooks/useApi';
import { useThumbnail } from '@shared/contexts';
import { useInView } from 'react-intersection-observer';

interface MediaGalleryProps {
  media: Media[];
  loading?: boolean;
  loadingMore?: boolean;
  error?: string | null;
  albumName?: string;
  albumId?: string;
  total?: number;
  hasMore?: boolean;
  onLoadMore?: () => void;
  onMediaDeleted?: () => void;
}

const MediaGallery: React.FC<MediaGalleryProps> = ({
  media,
  loading = false,
  loadingMore = false,
  error = null,
  albumName,
  albumId,
  total = 0,
  hasMore = true,
  onLoadMore,
  onMediaDeleted,
}) => {
  const navigate = useNavigate();
  
  // Thumbnail context
  const { isThumbnailMode, selectThumbnail } = useThumbnail();
  const [selectedMedia, setSelectedMedia] = useState<Media | null>(null);
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);

  // Media viewer modal state
  const [isViewerOpen, setIsViewerOpen] = useState(false);
  const [currentMediaIndex, setCurrentMediaIndex] = useState(0);
  const [thumbnailRect, setThumbnailRect] = useState<DOMRect | undefined>(undefined);

  // Multi-select state
  const [isSelectionMode, setIsSelectionMode] = useState(false);
  const [selectedMediaIds, setSelectedMediaIds] = useState<Set<string>>(new Set());
  const [isDeleting, setIsDeleting] = useState(false);
  const [isSettingThumbnail, setIsSettingThumbnail] = useState(false);

  // Scroll to top state
  const [isHeaderSticky, setIsHeaderSticky] = useState(false);
  const [showScrollToTop, setShowScrollToTop] = useState(false);

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

  // Infinite scroll setup
  const { ref: sentinelRef, inView } = useInView({
    threshold: 0,
    rootMargin: '200px',
  });

  // Scroll detection effect
  useEffect(() => {
    const handleScroll = () => {
      // Check if user has scrolled past the initial header position
      // We'll consider the header sticky when scrolled more than 100px
      setIsHeaderSticky(window.scrollY > 100);
      
      // Show scroll to top button when scrolled more than 300px on mobile
      setShowScrollToTop(window.scrollY > 300);
    };

    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  // Infinite scroll trigger effect
  useEffect(() => {
    if (inView && !loadingMore && hasMore && onLoadMore) {
      onLoadMore();
    }
  }, [inView, loadingMore, hasMore, onLoadMore]);

  // Scroll to top function
  const scrollToTop = () => {
    window.scrollTo({
      top: 0,
      behavior: 'smooth',
    });
  };

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
    setAlert((prev) => ({ ...prev, visible: false }));
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
        media: mediaItems,
      }))
      .sort((a, b) => b.weekStart.getTime() - a.weekStart.getTime());
  }, [media]);

  // Flatten grouped media for modal navigation
  const allMedia = React.useMemo(() => {
    return groupedMedia.flatMap((group) => group.media);
  }, [groupedMedia]);

  const handleInfoClick = (mediaItem: Media) => {
    setSelectedMedia(mediaItem);
    setIsDrawerOpen(true);
  };

  const handleCloseDrawer = () => {
    setIsDrawerOpen(false);
    setSelectedMedia(null);
  };

  const handleNavigateToAlbum = (targetAlbumId: string) => {
    setIsDrawerOpen(false);
    setSelectedMedia(null);
    navigate(`/albums/${targetAlbumId}`);
  };

  // Media viewer modal handlers
  const handleMediaClick = (mediaItem: Media, rect?: DOMRect) => {
    if (isThumbnailMode) {
      // In thumbnail selection mode, select this media as thumbnail
      selectThumbnail(mediaItem.id);
    } else {
      // Normal mode, open media viewer
      const index = allMedia.findIndex((m) => m.id === mediaItem.id);
      setCurrentMediaIndex(index);
      setThumbnailRect(rect);
      setIsViewerOpen(true);
    }
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
            inline: 'nearest',
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
    const allIds = new Set(allMedia.map((m) => m.id));
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
      const deletePromises = Array.from(selectedMediaIds).map((id) => deleteMedia(id));
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
      showAlert('error', 'Failed to delete some photos. Please try again.', 'Deletion Failed!');
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
      showAlert('success', 'The album thumbnail has been updated successfully.', 'Thumbnail Updated!');
    } catch (error) {
      console.error('Failed to set album thumbnail:', error);
      showAlert('error', 'Failed to set album thumbnail. Please try again.', 'Update Failed!');
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
      <div className="sticky top-0 z-30 bg-gray-50 dark:bg-slate-900 pb-4 mb-6 bg-opacity-95 dark:bg-opacity-95">
        <div className="flex items-center justify-between pt-4">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            {isThumbnailMode ? (
              <>
                <span className="text-blue-600 dark:text-blue-400">📸 Select Thumbnail</span>
                <span className="ml-2 text-sm text-gray-600 dark:text-gray-400">({total} photos)</span>
              </>
            ) : (
              <>
                Photos ({total})
                {isSelectionMode && selectedMediaIds.size > 0 && (
                  <span className="ml-2 text-sm text-blue-600 dark:text-blue-400">({selectedMediaIds.size} selected)</span>
                )}
              </>
            )}
          </h2>

          <div className="flex items-center space-x-2">


            {isSelectionMode && !isThumbnailMode ? (
              <>
                <button
                  onClick={selectAllMedia}
                  disabled={isDeleting || isSettingThumbnail}
                  className="px-4 py-2 text-sm font-medium border-2 border-gray-400 rounded-full text-gray-400 bg-transparent hover:text-black hover:border-black dark:hover:text-white dark:hover:border-white focus:outline-none transition-colors disabled:opacity-50"
                >
                  Select All
                </button>
                <button
                  onClick={clearSelection}
                  disabled={isDeleting || isSettingThumbnail}
                  className="px-4 py-2 text-sm font-medium border-2 border-gray-400 rounded-full text-gray-400 bg-transparent hover:text-black hover:border-black dark:hover:text-white dark:hover:border-white focus:outline-none transition-colors disabled:opacity-50"
                >
                  Clear
                </button>
                {selectedMediaIds.size === 1 && albumId && (
                  <button
                    onClick={handleSetThumbnail}
                    disabled={isDeleting || isSettingThumbnail}
                    className="px-4 py-2 text-sm font-medium border-2 border-green-400 rounded-full text-green-400 bg-transparent hover:text-green-600 hover:border-green-600 dark:hover:text-green-200 dark:hover:border-green-200 focus:outline-none transition-colors disabled:opacity-50"
                  >
                    {isSettingThumbnail ? 'Setting...' : 'Set Album Thumbnail'}
                  </button>
                )}
                {selectedMediaIds.size > 0 && (
                  <button
                    onClick={handleDeleteSelected}
                    disabled={isDeleting || isSettingThumbnail}
                    className="px-4 py-2 text-sm font-medium border-2 border-red-400 rounded-full text-red-400 bg-transparent hover:text-red-600 hover:border-red-600 dark:hover:text-red-200 dark:hover:border-red-200 focus:outline-none transition-colors disabled:opacity-50"
                  >
                    {isDeleting ? 'Deleting...' : `Delete (${selectedMediaIds.size})`}
                  </button>
                )}
                <button
                  onClick={toggleSelectionMode}
                  disabled={isDeleting || isSettingThumbnail}
                  className="px-4 py-2 text-sm font-medium border-2 border-gray-400 rounded-full text-gray-400 bg-transparent hover:text-black hover:border-black dark:hover:text-white dark:hover:border-white focus:outline-none transition-colors disabled:opacity-50"
                >
                  Cancel
                </button>
              </>
            ) : (
              <button
                onClick={toggleSelectionMode}
                className="flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-full border-2 border-gray-400 text-gray-400 bg-transparent hover:text-black hover:border-black dark:hover:text-white dark:hover:border-white focus:outline-none transition-colors"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                Select
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
            <div className="media-grid-container">
              {group.media.map((mediaItem) => (
                <MediaThumbnail
                  key={mediaItem.id}
                  media={mediaItem}
                  onInfoClick={handleInfoClick}
                  onClick={isSelectionMode && !isThumbnailMode ? () => toggleMediaSelection(mediaItem.id) : handleMediaClick}
                  isSelectionMode={isSelectionMode}
                  isSelected={selectedMediaIds.has(mediaItem.id)}
                />
              ))}
            </div>
          </div>
        ))}
      </div>

      {/* Infinite scroll sentinel */}
      <div ref={sentinelRef} className="w-full py-6 mt-8" style={{ minHeight: '50px' }} data-testid="infinite-scroll-sentinel">
        {hasMore ? (
          <div className="flex flex-col items-center space-y-3">
            <LoadingProgressBar loading={loadingMore} message="Loading more photos..." size="medium" />
            <div className="text-center text-sm text-gray-600 dark:text-gray-400 mb-3">Loading more photos...</div>
          </div>
        ) : (
          <div className="text-center">
            <div className="text-green-600 dark:text-green-400 font-medium">✅ All photos loaded</div>
            <div className="text-xs text-gray-500 mt-1">
              {media.length} of {total} photos
            </div>
          </div>
        )}
      </div>

      {/* EXIF Drawer */}
      <ExifDrawer 
        isOpen={isDrawerOpen} 
        media={selectedMedia} 
        onClose={handleCloseDrawer}
      />

      {/* Media Viewer Modal */}
      <MediaViewerModal
        isOpen={isViewerOpen}
        media={allMedia}
        currentIndex={currentMediaIndex}
        onClose={handleViewerClose}
        onIndexChange={handleIndexChange}
        thumbnailRect={thumbnailRect}
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

      {/* Floating Scroll to Top Button - Mobile only, hidden when modal is open */}
      {showScrollToTop && !isViewerOpen && (
        <button
          onClick={scrollToTop}
          className="fixed bottom-6 right-6 md:hidden z-50 p-3 bg-blue-600 hover:bg-blue-700 text-white rounded-full shadow-lg transition-all duration-300 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
          title="Scroll to top"
          aria-label="Scroll to top"
        >
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 15l7-7 7 7" />
          </svg>
        </button>
      )}
    </div>
  );
};



export default MediaGallery;
