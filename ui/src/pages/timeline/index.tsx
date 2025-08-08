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

import React, { useEffect, useState, useMemo } from 'react';
import { useAppDispatch, useAppSelector } from '@shared/store';
import { useMediaApi, useStatsApi } from '@shared/hooks/useApi';
import { ListMediaSortByEnum, ListMediaSortOrderEnum } from '@generated/api/media-api';
import TimelineMediaGallery from './components/TimelineMediaGallery';
import YearNavigation from './components/YearNavigation';

const TimelinePage: React.FC = () => {
  const dispatch = useAppDispatch();
  const { media, loading, error, total, fetchMedia } = useMediaApi();
  const { data: statsData, loading: statsLoading, fetchStats } = useStatsApi();

  // State for pagination and year scrolling
  const [currentPage, setCurrentPage] = useState(1);
  const [selectedYear, setSelectedYear] = useState<number | null>(null);
  const [visibleYear, setVisibleYear] = useState<number | null>(null); // Year currently in view
  const pageSize = 1000; // Load more items to enable scrolling through years

  /**
   * Effect for page initialization and data fetching
   * Fetches all media sorted by capture date (most recent first) and stats
   */
  useEffect(() => {
    // Fetch stats to get total media count and available years
    fetchStats();

    // Fetch all media sorted by capture date descending
    fetchMedia({
      limit: pageSize,
      offset: (currentPage - 1) * pageSize,
      sortBy: ListMediaSortByEnum.CapturedAt,
      sortOrder: ListMediaSortOrderEnum.Desc,
    });
  }, [fetchMedia, fetchStats, currentPage]);

  /**
   * Get available years from stats API
   * Years are already sorted in descending order from the API
   */
  const availableYears = useMemo(() => {
    return statsData?.years || [];
  }, [statsData]);

  /**
   * Handles pagination changes
   * Resets to page 1 when year filter changes
   */
  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  /**
   * Handles year selection from navigation
   * Scrolls to the selected year instead of filtering
   */
  const handleYearSelect = (year: number | null) => {
    setSelectedYear(year);
    if (year) {
      // Check if this is the most recent year (first in our sorted list)
      const mostRecentYear = availableYears[0];
      
      if (year === mostRecentYear) {
        // For the most recent year, scroll to the very top
        window.scrollTo({ top: 0, behavior: 'smooth' });
      } else {
        // For other years, find the year anchor and scroll to it
        const yearElement = document.getElementById(`year-${year}`);
        if (yearElement) {
          yearElement.scrollIntoView({
            behavior: 'smooth',
            block: 'start',
            inline: 'nearest',
          });
        }
      }
    } else {
      // Scroll to top for "All Years"
      window.scrollTo({ top: 0, behavior: 'smooth' });
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
    
    // Refresh media
    fetchMedia({
      limit: pageSize,
      offset: (currentPage - 1) * pageSize,
      sortBy: ListMediaSortByEnum.CapturedAt,
      sortOrder: ListMediaSortOrderEnum.Desc,
    });
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex gap-8">
          {/* Media Gallery */}
          <div className="flex-1 min-w-0">
            <TimelineMediaGallery
              media={media || []}
              loading={loading || statsLoading}
              error={error}
              total={statsData?.countMedia || total}
              currentPage={currentPage}
              pageSize={pageSize}
              onPageChange={handlePageChange}
              onMediaRefresh={handleMediaRefresh}
              onVisibleYearChange={handleVisibleYearChange}
            />
          </div>

          {/* Year Navigation Sidebar */}
          <div className="w-30 flex-shrink-0">
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
