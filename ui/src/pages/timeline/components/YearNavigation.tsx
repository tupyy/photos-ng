/**
 * Year Navigation Component
 *
 * Provides a sidebar for navigating through different years in the timeline.
 * Features:
 * - List of available years extracted from media
 * - Highlighting of currently selected year
 * - "All Years" option to show all media
 * - Sticky positioning for easy access while scrolling
 * - Loading states and empty states
 */

import React from 'react';

interface YearNavigationProps {
  availableYears: number[];
  selectedYear: number | null;
  onYearSelect: (year: number | null) => void;
  loading?: boolean;
}

const YearNavigation: React.FC<YearNavigationProps> = ({
  availableYears,
  selectedYear,
  onYearSelect,
  loading = false,
}) => {
  return (
    <div className="sticky top-24">
      <div className="bg-gray-50 dark:bg-gray-900">
        {/* Content */}
        <div>
          {loading ? (
            /* Loading state */
            <div className="space-y-2">
              {[...Array(5)].map((_, index) => (
                <div
                  key={index}
                  className="h-8 bg-gray-200 dark:bg-gray-600 rounded animate-pulse"
                />
              ))}
            </div>
          ) : availableYears.length === 0 ? (
            /* Empty state */
            <div className="text-center py-6">
              <svg className="mx-auto h-8 w-8 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
                No photos found
              </p>
            </div>
          ) : (
            /* Year navigation list */
            <div className="flex flex-col items-start space-y-0.5">
              {/* Individual years */}
              {availableYears.map((year) => (
                <button
                  key={year}
                  onClick={() => onYearSelect(year)}
                  className={`text-left px-2 py-1.5 text-sm transition-colors border-b-2 ${
                    selectedYear === year
                      ? 'font-bold text-gray-900 dark:text-white border-blue-500'
                      : 'font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 border-transparent'
                  }`}
                >
                  {year}
                </button>
              ))}
            </div>
          )}
        </div>

      </div>
    </div>
  );
};

export default YearNavigation;
