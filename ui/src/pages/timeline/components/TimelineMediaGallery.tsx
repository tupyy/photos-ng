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

import React, { useState } from 'react';
import { Media } from '@generated/models';
import MediaThumbnail from '@app/shared/components/MediaThumbnail';
import ExifDrawer from '@app/shared/components/ExifDrawer';
import { MediaViewerModal } from '@app/shared/components';

interface TimelineMediaGalleryProps {
  media: Media[];
  loading?: boolean;
  error?: string | null;
  total?: number;
  currentPage?: number;
  pageSize?: number;
  onPageChange?: (page: number) => void;
  onMediaRefresh?: () => void;
  onVisibleYearChange?: (year: number | null) => void;
}

const TimelineMediaGallery: React.FC<TimelineMediaGalleryProps> = ({
  media,
  loading = false,
  error = null,
  total = 0,
  currentPage = 1,
  pageSize = 100,
  onPageChange,
  onVisibleYearChange,
}) => {
  // EXIF drawer state
  const [selectedMedia, setSelectedMedia] = useState<Media | null>(null);
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);
  
  // Media viewer modal state
  const [isViewerOpen, setIsViewerOpen] = useState(false);
  const [currentMediaIndex, setCurrentMediaIndex] = useState(0);

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
        year: new Date(weekKey).getFullYear()
      }))
      .sort((a, b) => b.weekStart.getTime() - a.weekStart.getTime());
  }, [media]);

  /**
   * Flattened media array for modal navigation
   */
  const allMedia = React.useMemo(() => {
    return groupedMedia.flatMap(group => group.media);
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
            .filter(entry => entry.isIntersecting && entry.intersectionRatio > 0)
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
        threshold: 0 // Single threshold for minimal callbacks
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
        threshold: 0
      }
    );
    
    topObserver.observe(topDetector);

    // Observe year anchor elements
    const yearElements = document.querySelectorAll('[id^="year-"]');
    yearElements.forEach(element => observer.observe(element));

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
   * Handles media click for full-screen viewing
   */
  const handleMediaClick = (media: Media) => {
    const index = allMedia.findIndex(m => m.id === media.id);
    if (index !== -1) {
      setCurrentMediaIndex(index);
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
            inline: 'nearest'
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
        <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Photos</h2>
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
        <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Photos</h2>
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
        <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
          Photos
        </h2>
        <div className="text-center py-12">
          <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
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
    <div className="mt-8">
      {/* Simple Header without selection controls */}
      <div className="sticky top-0 z-30 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-700 pb-4 mb-6 backdrop-blur-sm bg-opacity-95 dark:bg-opacity-95">
        <div className="pt-4">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            Photos ({allMedia.length})
          </h2>
        </div>
      </div>
      
      {/* Grouped by weeks */}
      <div className="space-y-8">
        {groupedMedia.map((group, index) => {
          // Check if this is the first week of a new year
          const isFirstWeekOfYear = index === 0 || groupedMedia[index - 1].year !== group.year;
          
          return (
            <div key={group.weekStart.toISOString()}>
              {/* Year anchor for scrolling - placed at first week of each year */}
              {isFirstWeekOfYear && (
                <div id={`year-${group.year}`} className="scroll-mt-24"></div>
              )}
              
              {/* Week header */}
              <h3 className="text-md font-medium text-gray-700 dark:text-gray-300 mb-4 border-b border-gray-200 dark:border-gray-700 pb-2">
                {group.weekRange}
              </h3>
            
            {/* Media grid for this week */}
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-2">
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
    </div>
  );
};

// Simple Pagination component (reused from MediaGallery)
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

export default TimelineMediaGallery;
