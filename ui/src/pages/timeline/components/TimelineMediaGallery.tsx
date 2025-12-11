/**
 * Timeline Media Gallery Component
 *
 * A specialized version of MediaGallery for the timeline page.
 * Key differences from regular MediaGallery:
 * - No selection capabilities (read-only view)
 * - No album-specific operations (delete, set thumbnail)
 * - Simplified header without action buttons
 * - Grouped by weeks like the regular MediaGallery
 * - Supports media viewer modal for full-screen viewing
 */

import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Media } from '@generated/models';
import MediaThumbnail from '@app/shared/components/MediaThumbnail';
import ExifDrawer from '@app/shared/components/ExifDrawer';
import { MediaViewerModal, LoadingProgressBar } from '@app/shared/components';
import { useInView } from 'react-intersection-observer';

interface TimelineMediaGalleryProps {
  media: Media[];
  loading?: boolean;
  loadingMore?: boolean;
  error?: string | null;
  total?: number;
  hasMore?: boolean;
  scrollToYear?: number | null;
  onLoadMore?: (direction: 'forward' | 'backward') => void;
  onMediaRefresh?: () => void;
  onVisibleYearChange?: (year: number | null) => void;
  onScrollComplete?: () => void;
}

const TimelineMediaGallery: React.FC<TimelineMediaGalleryProps> = ({
  media,
  loading = false,
  loadingMore = false,
  error = null,
  total = 0,
  hasMore = true,
  scrollToYear = null,
  onLoadMore,
  onVisibleYearChange,
  onScrollComplete,
}) => {
  const navigate = useNavigate();

  // EXIF drawer state
  const [selectedMedia, setSelectedMedia] = useState<Media | null>(null);
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);

  // Media viewer modal state
  const [isViewerOpen, setIsViewerOpen] = useState(false);
  const [currentMediaIndex, setCurrentMediaIndex] = useState(0);
  const [thumbnailRect, setThumbnailRect] = useState<DOMRect | undefined>(undefined);

  // Scroll to top state (mobile only)
  const [showScrollToTop, setShowScrollToTop] = useState(false);

  // Debug state values (dev only)
  if (process.env.NODE_ENV === 'development') {
    console.log('🔍 TimelineMediaGallery state:', {
      hasMore,
      loadingMore,
      total,
      mediaLength: media?.length || 0,
      onLoadMoreProvided: !!onLoadMore,
    });
  }

  // Use react-intersection-observer for bottom sentinel (forward direction)
  const { ref: bottomRef, inView: bottomInView } = useInView({
    threshold: 0,
    rootMargin: '200px',
  });

  // Use react-intersection-observer for top sentinel (backward direction)
  const { ref: topRef, inView: topInView } = useInView({
    threshold: 0,
    rootMargin: '200px',
  });

  // Scroll detection effect for mobile scroll-to-top button
  useEffect(() => {
    const handleScroll = () => {
      // Show scroll to top button when scrolled more than 300px on mobile
      setShowScrollToTop(window.scrollY > 300);
    };

    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  // Scroll to top function
  const scrollToTop = () => {
    window.scrollTo({
      top: 0,
      behavior: 'smooth',
    });
  };

  // Trigger load more when bottom sentinel comes into view (forward direction)
  useEffect(() => {
    if (bottomInView && !loadingMore && onLoadMore) {
      if (process.env.NODE_ENV === 'development') {
        console.log('🚀 Bottom sentinel triggered - loading more forward');
      }
      onLoadMore('forward');
    }
  }, [bottomInView, loadingMore, onLoadMore]);

  // Handle scrolling to year when scrollToYear prop changes
  useEffect(() => {
    if (scrollToYear && media && media.length > 0) {
      // Check if the year is available in the current media
      const hasYearLoaded = media.some((m) => new Date(m.capturedAt).getFullYear() === scrollToYear);

      if (hasYearLoaded) {
        // Small delay to ensure DOM has updated after media rendering
        setTimeout(() => {
          const yearElement = document.getElementById(`year-${scrollToYear}`);
          if (yearElement) {
            yearElement.scrollIntoView({
              behavior: 'smooth',
              block: 'start',
              inline: 'nearest',
            });
            // Notify parent that scroll is complete
            onScrollComplete?.();
          }
        }, 100);
      }
    }
  }, [scrollToYear, media, onScrollComplete]);

  /**
   * Helper function to get the start of the week (Monday) for a given date
   */
  const getWeekStart = (date: Date) => {
    const d = new Date(date);
    const day = d.getDay();
    const diff = d.getDate() - day + (day === 0 ? -6 : 1); // Adjust for Monday start
    d.setHours(0, 0, 0, 0);
    return d;
  };

  /**
   * Helper function to format week range for display
   */
  const formatWeekRange = (weekStart: Date) => {
    const weekEnd = new Date(weekStart);
    weekEnd.setDate(weekStart.getDate() + 6);

    const startDay = weekStart.getDate();
    const endDay = weekEnd.getDate();
    const startMonth = weekStart.toLocaleDateString('en-US', { month: 'long' });
    const endMonth = weekEnd.toLocaleDateString('en-US', { month: 'long' });
    const year = weekStart.getFullYear();

    if (startMonth === endMonth) {
      return `${startDay} - ${endDay} ${startMonth} ${year}`;
    } else {
      return `${startDay} ${startMonth} - ${endDay} ${endMonth} ${year}`;
    }
  };

  /**
   * Group media by weeks for better organization
   */
  const groupedMedia = React.useMemo(() => {
    if (!media || media.length === 0) return [];

    // Sort media by captured date (descending) and then by filename
    const sortedMedia = [...media].sort((a, b) => {
      const capturedAtA = new Date(a.capturedAt).getTime();
      const capturedAtB = new Date(b.capturedAt).getTime();

      if (capturedAtA !== capturedAtB) {
        return capturedAtB - capturedAtA; // Descending order
      }
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

    return Array.from(groups.entries())
      .map(([weekKey, mediaItems]) => ({
        weekStart: new Date(weekKey),
        weekRange: formatWeekRange(new Date(weekKey)),
        media: mediaItems,
        year: new Date(weekKey).getFullYear(),
      }))
      .sort((a, b) => b.weekStart.getTime() - a.weekStart.getTime());
  }, [media]);

  /**
   * Flattened media array for modal navigation
   */
  const allMedia = React.useMemo(() => {
    return groupedMedia.flatMap((group) => group.media);
  }, [groupedMedia]);

  /**
   * Ref to track the last detected year to avoid unnecessary updates
   */
  const lastDetectedYear = React.useRef<number | null>(null);

  /**
   * Set up intersection observer to detect which year is currently visible
   */
  React.useEffect(() => {
    if (!onVisibleYearChange || groupedMedia.length === 0) return;

    // Use only intersection observer for better performance
    // Remove scroll listener completely to eliminate scroll jank
    const observer = new IntersectionObserver(
      (entries) => {
        // Use requestAnimationFrame to ensure smooth scrolling
        requestAnimationFrame(() => {
          // Filter intersecting entries and find the topmost one
          const intersectingEntries = entries
            .filter((entry) => entry.isIntersecting && entry.intersectionRatio > 0)
            .sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top);

          if (intersectingEntries.length > 0) {
            const topEntry = intersectingEntries[0];
            const yearMatch = topEntry.target.id.match(/year-(\d+)/);
            if (yearMatch) {
              const year = parseInt(yearMatch[1], 10);
              // Only update if the year actually changed
              if (lastDetectedYear.current !== year) {
                lastDetectedYear.current = year;
                onVisibleYearChange(year);
              }
            }
          }
        });
      },
      {
        root: null,
        rootMargin: '-100px 0px -70% 0px', // More conservative margins
        threshold: 0, // Single threshold for minimal callbacks
      }
    );

    // Add a small invisible element at the very top to detect when we're at the beginning
    const topDetector = document.createElement('div');
    topDetector.id = 'timeline-top-detector';
    topDetector.style.position = 'absolute';
    topDetector.style.top = '0';
    topDetector.style.height = '1px';
    topDetector.style.width = '100%';
    topDetector.style.pointerEvents = 'none';

    const timelineContainer = document.querySelector('.timeline-container') || document.body;
    timelineContainer.appendChild(topDetector);

    // Observer for top detector
    const topObserver = new IntersectionObserver(
      (entries) => {
        const entry = entries[0];
        if (entry.isIntersecting && groupedMedia.length > 0) {
          const firstYear = groupedMedia[0]?.year;
          if (firstYear && lastDetectedYear.current !== firstYear) {
            lastDetectedYear.current = firstYear;
            onVisibleYearChange(firstYear);
          }
        }
      },
      {
        root: null,
        rootMargin: '0px',
        threshold: 0,
      }
    );

    topObserver.observe(topDetector);

    // Observe year anchor elements
    const yearElements = document.querySelectorAll('[id^="year-"]');
    yearElements.forEach((element) => observer.observe(element));

    return () => {
      observer.disconnect();
      topObserver.disconnect();
      if (topDetector.parentNode) {
        topDetector.parentNode.removeChild(topDetector);
      }
    };
  }, [groupedMedia, onVisibleYearChange]);

  /**
   * Handles media thumbnail click for EXIF info
   */
  const handleInfoClick = (media: Media) => {
    setSelectedMedia(media);
    setIsDrawerOpen(true);
  };

  /**
   * Handles closing the EXIF drawer
   */
  const handleCloseDrawer = () => {
    setIsDrawerOpen(false);
  };

  /**
   * Handles navigation to album
   */
  const handleNavigateToAlbum = (albumId: string) => {
    setIsDrawerOpen(false);
    navigate(`/albums/${albumId}`);
  };

  /**
   * Handles media click for full-screen viewing
   */
  const handleMediaClick = (media: Media, rect?: DOMRect) => {
    const index = allMedia.findIndex((m) => m.id === media.id);
    if (index !== -1) {
      setCurrentMediaIndex(index);
      setThumbnailRect(rect);
      setIsViewerOpen(true);
    }
  };

  /**
   * Handles closing the media viewer modal and scrolls to the last viewed media
   */
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

  /**
   * Handles index change in media viewer modal
   */
  const handleIndexChange = (index: number) => {
    setCurrentMediaIndex(index);
  };

  // Loading state
  if (loading) {
    return (
      <div className="mt-8">
        <div className="flex justify-center items-center py-8">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="mt-8">
        <div className="text-center py-8">
          <div className="text-red-600 dark:text-red-400 mb-2">Failed to load photos</div>
          <div className="text-sm text-gray-500 dark:text-gray-400">{error}</div>
        </div>
      </div>
    );
  }

  // Empty state
  if (!media || media.length === 0) {
    return (
      <div className="mt-8">
        <div className="text-center py-12">
          <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"
            />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-white">No photos found</h3>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            Start uploading photos to see them in your timeline.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div>
      {/* Top sentinel for backward infinite scroll */}
      <div ref={topRef} className="w-full py-2" style={{ minHeight: '10px' }} data-testid="backward-scroll-sentinel">
        {/* This sentinel triggers loading older content when it comes into view */}
      </div>

      {/* Grouped by weeks */}
      <div className="space-y-8">
        {groupedMedia.map((group, index) => {
          // Check if this is the first week of a new year (handles both directions)
          const isFirstWeekOfYear = index === 0 || groupedMedia[index - 1].year !== group.year;
          // Check if this is the last week of a year (for backward scroll scenarios)
          const isLastWeekOfYear = index === groupedMedia.length - 1 || groupedMedia[index + 1].year !== group.year;

          return (
            <div key={group.weekStart.toISOString()}>
              {/* Year anchor for scrolling - placed at first week of each year */}
              {isFirstWeekOfYear && <div id={`year-${group.year}`} className="scroll-mt-24"></div>}

              {/* Additional year anchor for last week to handle backward navigation */}
              {isLastWeekOfYear && !isFirstWeekOfYear && (
                <div id={`year-${group.year}-end`} className="scroll-mt-24"></div>
              )}

              {/* Week header */}
              <h3 className="text-md font-medium text-gray-700 dark:text-gray-300 mb-4">{group.weekRange}</h3>

              {/* Media grid for this week */}
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-1">
                {group.media.map((mediaItem) => (
                  <MediaThumbnail
                    key={mediaItem.id}
                    media={mediaItem}
                    onInfoClick={handleInfoClick}
                    onClick={handleMediaClick}
                    isSelectionMode={false} // Always disabled for timeline
                    isSelected={false} // Never selected in timeline
                  />
                ))}
              </div>
            </div>
          );
        })}
      </div>

      {/* Bottom sentinel for forward infinite scroll - always render at the bottom */}
      <div
        ref={bottomRef}
        className="w-full py-6 mt-8"
        style={{ minHeight: '50px' }}
        data-testid="forward-scroll-sentinel"
      >
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
        onNavigateToAlbum={handleNavigateToAlbum}
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

export default TimelineMediaGallery;
