/**
 * Mobile Year Dropdown Component
 *
 * A mobile-friendly dropdown for year navigation in the timeline.
 * Features:
 * - Compact dropdown select for space efficiency
 * - Shows "All Years" option
 * - Responsive design (only visible on mobile)
 * - Consistent styling with the rest of the app
 */

import React from 'react';

interface MobileYearDropdownProps {
  availableYears: number[];
  selectedYear: number | null;
  onYearSelect: (year: number | null) => void;
  loading?: boolean;
}

const MobileYearDropdown: React.FC<MobileYearDropdownProps> = ({
  availableYears,
  selectedYear,
  onYearSelect,
  loading = false,
}) => {
  const handleSelectChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    const value = event.target.value;
    if (value === 'all') {
      onYearSelect(null);
    } else {
      onYearSelect(parseInt(value, 10));
    }
  };

  if (loading) {
    return (
      <div className="mb-4 md:hidden">
        <div className="h-10 bg-gray-200 dark:bg-gray-600 rounded animate-pulse"></div>
      </div>
    );
  }

  if (availableYears.length === 0) {
    return null;
  }

  return (
    <div className="mb-4 md:hidden">
      <label htmlFor="year-select" className="sr-only">
        Select year
      </label>
      <select
        id="year-select"
        value={selectedYear || 'all'}
        onChange={handleSelectChange}
        className="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
      >
        <option value="all">All Years</option>
        {availableYears.map((year) => (
          <option key={year} value={year}>
            {year}
          </option>
        ))}
      </select>
    </div>
  );
};

export default MobileYearDropdown;
