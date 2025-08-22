/**
 * Timeline Page Component
 *
 * Displays all media in chronological order with year-based navigation.
 * Features:
 * - All media displayed in captured_at descending order
 * - Year navigation sidebar for quick temporal navigation
 * - MediaGallery without selection capabilities (read-only view)
 * - Responsive layout with sidebar on the right
 *
 * Uses the /api/v1/media endpoint to fetch all media sorted by captured_at
 * and the /api/v1/stats endpoint to get total media count and available years.
 */

import React, { useEffect, useState, useMemo, useCallback } from 'react';
import { useAppDispatch, useAppSelector } from '@shared/store';
import { useMediaApi, useStatsApi } from '@shared/hooks/useApi';
import { ListMediaSortByEnum, ListMediaSortOrderEnum, ListMediaDirectionEnum } from '@generated/api/media-api';
import TimelineMediaGallery from './components/TimelineMediaGallery';
import YearNavigation from './components/YearNavigation';
import MobileYearDropdown from './components/MobileYearDropdown';
import { TIMELINE_PAGE_SIZE } from '@app/shared/config';

const TimelinePage: React.FC = () => {
  const dispatch = useAppDispatch();
  const { media, loading, loadingMore, error, hasMore, nextCursor, fetchMedia } = useMediaApi();
  const { data: statsData, loading: statsLoading, fetchStats } = useStatsApi();

  // Debug logging (dev only)
  if (process.env.NODE_ENV === 'development') {
    console.log('📊 Timeline state:', {
      mediaCount: media?.length || 0,
      hasMore,
      loading,
      loadingMore,
      error,
      pageSize,
      statsTotal: statsData?.countMedia,
    });
  }

  // State for year scrolling
  const [currentYear, setCurrentYear] = useState<number>(new Date().getFullYear());
  const [visibleYear, setVisibleYear] = useState<number | null>(null); // Year currently in view
  const [currentDirection, setCurrentDirection] = useState<ListMediaDirectionEnum>(ListMediaDirectionEnum.Forward); // Track current navigation direction
  const [scrollToYear, setScrollToYear] = useState<number | null>(null); // Year to scroll to after media loads
  const pageSize = TIMELINE_PAGE_SIZE; // Load more items to enable scrolling through years

  /**
   * Effect for page initialization and data fetching
   * Fetches all media sorted by capture date (most recent first) and stats
   */
  useEffect(() => {
    // Fetch stats to get total media count and available years
    fetchStats();

    // Fetch initial media sorted by capture date descending
    fetchMedia({
      limit: pageSize,
      sortBy: ListMediaSortByEnum.CapturedAt,
      sortOrder: ListMediaSortOrderEnum.Desc,
      direction: ListMediaDirectionEnum.Forward,
    });
  }, [fetchMedia, fetchStats]);

  /**
   * Get available years from stats API
   * Years are already sorted in descending order from the API
   */
  const availableYears = useMemo(() => {
    return statsData?.years || [];
  }, [statsData]);

  /**
   * Handles loading more media for infinite scroll
   * Supports both forward and backward directions
   */
  const handleLoadMore = useCallback((direction: 'forward' | 'backward') => {
    if (!loadingMore && hasMore) {
      const scrollDirection = direction === 'backward'
        ? ListMediaDirectionEnum.Backward
        : ListMediaDirectionEnum.Forward;

      if (process.env.NODE_ENV === 'development') {
        console.log('🔄 Loading more with cursor:', nextCursor, 'direction:', scrollDirection);
      }

      fetchMedia({
        limit: pageSize,
        cursor: nextCursor,
        sortBy: ListMediaSortByEnum.CapturedAt,
        sortOrder: ListMediaSortOrderEnum.Desc,
        direction: scrollDirection,
      });

      // Update current direction for future operations
      setCurrentDirection(scrollDirection);
    }
  }, [loadingMore, hasMore, nextCursor, fetchMedia, pageSize]);

  /**
   * Creates a cursor for jumping to a specific year
   * Uses December 31st of the previous year as the cursor position
   */
  const createYearCursor = (year: number): string => {
    // Create cursor for December 31st of the previous year at midnight
    const yearStart = `${year}-12-31T00:00:00Z`;
    // Use a dummy ID that will be lexicographically first
    const cursor = {
      captured_at: yearStart,
      id: '00000000000000000000000000000000', // 32 chars of zeros for earliest possible ID
    };
    return btoa(JSON.stringify(cursor));
  };

  /**
   * Handles year selection from navigation
   * Uses cursor-based jumping to navigate to specific years
   */
  const handleYearSelect = async (year: number | null) => {
    if (!year) {
      return;
    }

    // Check if we have media for this year loaded
    const hasYearLoaded = media?.some((m) => new Date(m.capturedAt).getFullYear() === year);

    if (hasYearLoaded) {
      // Year is loaded, tell TimelineMediaGallery to scroll to it
      setScrollToYear(year);
    } else {
      // Year not loaded, use cursor-based jumping
      const yearCursor = createYearCursor(year);

      if (process.env.NODE_ENV === 'development') {
        console.log(`🗓️ Jumping to year ${year} using cursor-based navigation`);
        console.log('📍 Year cursor:', yearCursor);
      }

      let direction = ListMediaDirectionEnum.Forward;
      if (year - currentYear > 0) {
        setCurrentDirection(ListMediaDirectionEnum.Backward);
      }

      try {
        await fetchMedia({
          limit: pageSize,
          cursor: yearCursor,
          direction: direction,
          sortBy: ListMediaSortByEnum.CapturedAt,
          sortOrder: ListMediaSortOrderEnum.Desc,
          forceRefresh: true, // Force refresh to replace existing media instead of appending
        });
        setCurrentYear(year);
        // Only set scroll year after fetch completes successfully
        setScrollToYear(year);
      } catch (error) {
        if (process.env.NODE_ENV === 'development') {
          console.error('Failed to fetch media for year:', year, error);
        }
      }
    }
  };

  /**
   * Handles visible year changes from scroll events
   * Updates the year navigation to reflect current scroll position
   */
  const handleVisibleYearChange = (year: number | null) => {
    setVisibleYear(year);
  };

  /**
   * Handles media refresh after operations (if needed in the future)
   * Also refreshes stats to keep data in sync
   */
  const handleMediaRefresh = () => {
    // Refresh stats to get updated counts and years
    fetchStats();

    // Refresh media from the beginning
    setCurrentDirection(ListMediaDirectionEnum.Forward);
    fetchMedia({
      limit: pageSize,
      sortBy: ListMediaSortByEnum.CapturedAt,
      sortOrder: ListMediaSortOrderEnum.Desc,
      direction: ListMediaDirectionEnum.Forward,
    });
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="md:flex md:gap-8">
          {/* Media Gallery */}
          <div className="flex-1 min-w-0">
            {/* Mobile Year Dropdown - Only visible on mobile */}
            <MobileYearDropdown
              availableYears={availableYears}
              selectedYear={visibleYear}
              onYearSelect={handleYearSelect}
              loading={loading || statsLoading}
            />

            <TimelineMediaGallery
              media={media || []}
              loading={loading || statsLoading}
              loadingMore={loadingMore}
              error={error}
              total={statsData?.countMedia || 0}
              hasMore={hasMore}
              scrollToYear={scrollToYear}
              onLoadMore={handleLoadMore}
              onMediaRefresh={handleMediaRefresh}
              onVisibleYearChange={handleVisibleYearChange}
              onScrollComplete={() => setScrollToYear(null)}
            />
          </div>

          {/* Year Navigation Sidebar - Hidden on mobile */}
          <div className="hidden md:block w-30 flex-shrink-0">
            <YearNavigation
              availableYears={availableYears}
              selectedYear={visibleYear} // Use visibleYear instead of selectedYear
              onYearSelect={handleYearSelect}
              loading={loading || statsLoading}
            />
          </div>
        </div>
      </div>
    </div>
  );
};

export default TimelinePage;
