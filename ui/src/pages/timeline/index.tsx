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
import { ListMediaSortByEnum, ListMediaSortOrderEnum } from '@generated/api/media-api';
import TimelineMediaGallery from './components/TimelineMediaGallery';
import YearNavigation from './components/YearNavigation';
import MobileYearDropdown from './components/MobileYearDropdown';
import { TIMELINE_PAGE_SIZE } from '@app/shared/config';

const TimelinePage: React.FC = () => {
  const dispatch = useAppDispatch();
  const { media, loading, loadingMore, error, total, hasMore, fetchMedia, loadNextPage } = useMediaApi();
  const { data: statsData, loading: statsLoading, fetchStats } = useStatsApi();

  // Debug logging (dev only)
  if (process.env.NODE_ENV === 'development') {
    console.log('📊 Timeline state:', {
      mediaCount: media?.length || 0,
      total,
      hasMore,
      loading,
      loadingMore,
      error,
      pageSize,
      statsTotal: statsData?.countMedia,
    });
  }

  // State for year scrolling
  const [selectedYear, setSelectedYear] = useState<number | null>(null);
  const [visibleYear, setVisibleYear] = useState<number | null>(null); // Year currently in view
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
      offset: 0,
      sortBy: ListMediaSortByEnum.CapturedAt,
      sortOrder: ListMediaSortOrderEnum.Desc,
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
   */
  const handleLoadMore = useCallback(() => {
    if (!loadingMore && media) {
      if (process.env.NODE_ENV === 'development') {
        console.log('🔄 Loading more - offset:', media.length);
      }
      
      fetchMedia({
        limit: pageSize,
        offset: media.length,
        sortBy: ListMediaSortByEnum.CapturedAt,
        sortOrder: ListMediaSortOrderEnum.Desc,
      });
    }
  }, [loadingMore, media, pageSize, fetchMedia]);

  /**
   * Handles year selection from navigation
   * Fetches media for that year if not loaded, otherwise scrolls to it
   */
  const handleYearSelect = (year: number | null) => {
    setSelectedYear(year);
    
    if (!year) {
      // "All Years" selected - reload from the beginning
      fetchMedia({
        limit: pageSize,
        offset: 0,
        sortBy: ListMediaSortByEnum.CapturedAt,
        sortOrder: ListMediaSortOrderEnum.Desc,
      });
      return;
    }

    // Check if we have media for this year loaded
    const hasYearLoaded = media?.some(m => new Date(m.capturedAt).getFullYear() === year);
    
    if (hasYearLoaded) {
      // Year is loaded, scroll to it
      const yearElement = document.getElementById(`year-${year}`);
      if (yearElement) {
        yearElement.scrollIntoView({
          behavior: 'smooth',
          block: 'start',
          inline: 'nearest',
        });
      }
    } else {
      // Year not loaded, fetch media starting from that year
      const startDate = `${year}-01-01`;
      const endDate = `${year}-12-31`;
      
      if (process.env.NODE_ENV === 'development') {
        console.log(`🗓️ Fetching media for year ${year}`, { startDate, endDate });
        console.log('🧪 Testing: First try without date filters to see if we get any data...');
      }
      
      // Test: Try fetching without date filters first to see if the issue is with date filtering
      const queryParams = {
        limit: pageSize,
        offset: 0,
        // startDate,  // Temporarily disabled
        // endDate,    // Temporarily disabled
        sortBy: ListMediaSortByEnum.CapturedAt,
        sortOrder: ListMediaSortOrderEnum.Desc,
      };
      
      if (process.env.NODE_ENV === 'development') {
        console.log('📤 API Query Parameters (without date filters for testing):', queryParams);
      }
      
      fetchMedia(queryParams);
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
    fetchMedia({
      limit: pageSize,
      offset: 0,
      sortBy: ListMediaSortByEnum.CapturedAt,
      sortOrder: ListMediaSortOrderEnum.Desc,
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
              total={statsData?.countMedia || total}
              hasMore={hasMore}
              onLoadMore={handleLoadMore}
              onMediaRefresh={handleMediaRefresh}
              onVisibleYearChange={handleVisibleYearChange}
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
