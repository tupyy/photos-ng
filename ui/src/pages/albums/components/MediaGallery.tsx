import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Media } from '@generated/models';
import MediaThumbnail from '@app/shared/components/MediaThumbnail';
import ExifDrawer from '@app/shared/components/ExifDrawer';
import { MediaViewerModal, Alert, LoadingProgressBar } from '@app/shared/components';
import { useAlbumsApi } from '@shared/hooks/useApi';
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
  groupByWeek?: boolean;
  viewMode?: 'grid' | 'masonry';
  onLoadMore?: () => void;
  onMediaDeleted?: () => void;
  isSelectionMode?: boolean;
  selectedIds?: Set<string>;
  onToggleSelection?: (mediaId: string) => void;
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
  groupByWeek = true,
  viewMode = 'grid',
  onLoadMore,
  onMediaDeleted,
  isSelectionMode = false,
  selectedIds = new Set(),
  onToggleSelection,
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

  // Thumbnail setting state
  const [isSettingThumbnail, setIsSettingThumbnail] = useState(false);

  // Scroll to top state
  const [isHeaderSticky, setIsHeaderSticky] = useState(false);
  const [showScrollToTop, setShowScrollToTop] = useState(false);

  // Alert state
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

  // Sorted media for flat grid mode and modal navigation
  const sortedMedia = React.useMemo(() => {
    if (!media || media.length === 0) return [];

    return [...media].sort((a, b) => {
      const capturedAtA = new Date(a.capturedAt).getTime();
      const capturedAtB = new Date(b.capturedAt).getTime();

      if (capturedAtA !== capturedAtB) {
        return capturedAtB - capturedAtA; // Descending order
      }

      // If capturedAt is the same, sort by filename (ascending)
      return a.filename.localeCompare(b.filename);
    });
  }, [media]);

  // Flatten grouped media for modal navigation (use sorted media directly when not grouping)
  const allMedia = React.useMemo(() => {
    if (!groupByWeek) {
      return sortedMedia;
    }
    return groupedMedia.flatMap((group) => group.media);
  }, [groupByWeek, sortedMedia, groupedMedia]);

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

  // Handle click on media item
  const handleMediaItemClick = (mediaItem: Media, rect?: DOMRect) => {
    if (isSelectionMode && !isThumbnailMode && onToggleSelection) {
      onToggleSelection(mediaItem.id);
    } else {
      handleMediaClick(mediaItem, rect);
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

  // Get the appropriate container class based on view mode
  const getContainerClass = () => {
    if (viewMode === 'masonry') {
      return 'media-masonry-container';
    }
    return 'media-grid-container';
  };

  return (
    <div>
      {/* Photo Grid - Grouped by weeks, flat grid, or masonry */}
      {groupByWeek && viewMode !== 'masonry' ? (
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
                    onClick={handleMediaItemClick}
                    isSelectionMode={isSelectionMode}
                    isSelected={selectedIds.has(mediaItem.id)}
                  />
                ))}
              </div>
            </div>
          ))}
        </div>
      ) : (
        /* Flat grid or masonry layout */
        <div className={getContainerClass()}>
          {sortedMedia.map((mediaItem) => (
            <MediaThumbnail
              key={mediaItem.id}
              media={mediaItem}
              onInfoClick={handleInfoClick}
              onClick={handleMediaItemClick}
              isSelectionMode={isSelectionMode}
              isSelected={selectedIds.has(mediaItem.id)}
            />
          ))}
        </div>
      )}

      {/* Infinite scroll sentinel */}
      <div ref={sentinelRef} className="w-full py-6 mt-8" style={{ minHeight: '50px' }} data-testid="infinite-scroll-sentinel">
        {hasMore && (
          <div className="flex flex-col items-center space-y-3">
            <LoadingProgressBar loading={loadingMore} message="Loading more photos..." size="medium" />
            <div className="text-center text-sm text-gray-600 dark:text-gray-400 mb-3">Loading more photos...</div>
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
