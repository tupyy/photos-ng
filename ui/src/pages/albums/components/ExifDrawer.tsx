import React from 'react';
import { Media } from '@generated/models';

interface ExifDrawerProps {
  isOpen: boolean;
  media: Media | null;
  onClose: () => void;
}

const ExifDrawer: React.FC<ExifDrawerProps> = ({ isOpen, media, onClose }) => {
  if (!isOpen || !media) return null;

  // Format EXIF data for display
  const formatExifData = () => {
    if (!media.exif || media.exif.length === 0) {
      return [];
    }

    return media.exif.map(exifItem => ({
      key: exifItem.key || 'Unknown',
      value: exifItem.value || 'N/A'
    }));
  };

  const exifData = formatExifData();

  return (
    <>
      {/* Backdrop */}
      <div 
        className="fixed inset-0 bg-black bg-opacity-50 z-40 transition-opacity"
        onClick={onClose}
      />
      
      {/* Drawer */}
      <div className="fixed inset-0 md:right-0 md:left-auto md:top-0 md:bottom-auto h-full w-full md:w-96 bg-white dark:bg-gray-800 shadow-xl z-50 transform transition-transform duration-300 ease-in-out">
        <div className="flex flex-col h-full">
          {/* Header */}
          <div className="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">
              Photo Information
            </h3>
            <button
              onClick={onClose}
              className="p-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
            >
              <svg
                className="w-6 h-6 text-gray-500 dark:text-gray-400"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth="2"
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>

          {/* Content */}
          <div className="flex-1 overflow-y-auto p-4">
            {/* Thumbnail */}
            <div className="mb-6">
              <h4 className="text-sm font-medium text-gray-900 dark:text-white mb-3">Preview</h4>
              <div className="w-full aspect-square bg-gray-200 dark:bg-gray-700 rounded-lg overflow-hidden">
                <img
                  src={media.thumbnail}
                  alt={`Preview of ${media.filename}`}
                  className="w-full h-full object-cover"
                  onError={(e) => {
                    // Fallback to placeholder if thumbnail fails to load
                    e.currentTarget.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjNmNGY2Ii8+CiAgPHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSIxNCIgZmlsbD0iIzk5YTNhZiIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPk5vIEltYWdlPC90ZXh0Pgo8L3N2Zz4K';
                  }}
                />
              </div>
            </div>

            {/* Basic Info */}
            <div className="mb-6">
              <h4 className="text-sm font-medium text-gray-900 dark:text-white mb-3">Basic Information</h4>
              <table className="w-full text-sm">
                <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                  <tr>
                    <td className="py-2 text-gray-500 dark:text-gray-400">Filename</td>
                    <td className="py-2 text-gray-900 dark:text-white font-medium break-words">{media.filename}</td>
                  </tr>
                  <tr>
                    <td className="py-2 text-gray-500 dark:text-gray-400">Type</td>
                    <td className="py-2 text-gray-900 dark:text-white font-medium">{media.type}</td>
                  </tr>
                  <tr>
                    <td className="py-2 text-gray-500 dark:text-gray-400">Captured At</td>
                    <td className="py-2 text-gray-900 dark:text-white font-medium">{media.capturedAt}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            {/* EXIF Data */}
            {exifData.length > 0 ? (
              <div>
                <h4 className="text-sm font-medium text-gray-900 dark:text-white mb-3">EXIF Data</h4>
                <table className="w-full text-sm">
                  <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                    {exifData.map((item, index) => (
                      <tr key={index}>
                        <td className="py-2 text-gray-500 dark:text-gray-400 align-top">{item.key}</td>
                        <td className="py-2 text-gray-900 dark:text-white font-medium break-words">{item.value}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="text-center py-8">
                <svg
                  className="mx-auto h-12 w-12 text-gray-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                  />
                </svg>
                <h3 className="mt-2 text-sm font-medium text-gray-900 dark:text-white">No EXIF data</h3>
                <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  This image doesn't contain any EXIF metadata.
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </>
  );
};

export default ExifDrawer;