import React from 'react';
import { XMarkIcon, ShareIcon, ArrowDownTrayIcon, TrashIcon, PhotoIcon } from '@heroicons/react/24/outline';

export interface SelectionBarProps {
  selectedCount: number;
  isVisible: boolean;
  onClose: () => void;
  onDelete?: () => void;
  onShare?: () => void;
  onDownload?: () => void;
  onSetThumbnail?: () => void;
  isDeleting?: boolean;
  isSettingThumbnail?: boolean;
}

/**
 * SelectionBar - Floating bottom bar for bulk selection actions
 *
 * Features:
 * - Shows selected item count
 * - Action buttons for share, download, delete
 * - Smooth slide-up animation when visible
 * - Fixed at bottom center of screen
 */
const SelectionBar: React.FC<SelectionBarProps> = ({
  selectedCount,
  isVisible,
  onClose,
  onDelete,
  onShare,
  onDownload,
  onSetThumbnail,
  isDeleting = false,
  isSettingThumbnail = false,
}) => {
  return (
    <div
      className={`fixed bottom-6 left-1/2 transform -translate-x-1/2 z-50 transition-all duration-300 ${
        isVisible
          ? 'translate-y-0 opacity-100'
          : 'translate-y-24 opacity-0 pointer-events-none'
      }`}
    >
      <div className="bg-gray-900 dark:bg-gray-800 text-white shadow-2xl rounded-2xl px-6 py-3 flex items-center gap-6 border border-gray-700/50 backdrop-blur-md">
        {/* Selected count and close button */}
        <div className="flex items-center gap-3 border-r border-gray-700 pr-6">
          <div className="bg-blue-600 text-white text-xs font-bold rounded-full w-6 h-6 flex items-center justify-center">
            {selectedCount}
          </div>
          <span className="text-sm font-medium">Selected</span>
          <button
            onClick={onClose}
            className="ml-2 text-gray-400 hover:text-white transition-colors"
            title="Exit selection mode"
          >
            <XMarkIcon className="w-5 h-5" />
          </button>
        </div>

        {/* Action buttons */}
        <div className="flex items-center gap-2">
          {onShare && (
            <button
              onClick={onShare}
              disabled={selectedCount === 0}
              className="p-2 hover:bg-gray-700 rounded-lg transition-colors text-gray-300 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed"
              title="Share"
            >
              <ShareIcon className="w-5 h-5" />
            </button>
          )}
          {onDownload && (
            <button
              onClick={onDownload}
              disabled={selectedCount === 0}
              className="p-2 hover:bg-gray-700 rounded-lg transition-colors text-gray-300 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed"
              title="Download"
            >
              <ArrowDownTrayIcon className="w-5 h-5" />
            </button>
          )}
          {onSetThumbnail && (
            <button
              onClick={onSetThumbnail}
              disabled={isDeleting || isSettingThumbnail}
              className="p-2 hover:bg-blue-900/50 hover:text-blue-400 rounded-lg transition-colors text-gray-300 disabled:opacity-50 disabled:cursor-not-allowed"
              title="Set as album thumbnail"
            >
              <PhotoIcon className="w-5 h-5" />
            </button>
          )}
          {onDelete && (
            <button
              onClick={onDelete}
              disabled={selectedCount === 0 || isDeleting}
              className="p-2 hover:bg-red-900/50 hover:text-red-400 rounded-lg transition-colors text-gray-300 disabled:opacity-50 disabled:cursor-not-allowed"
              title="Delete"
            >
              <TrashIcon className="w-5 h-5" />
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default SelectionBar;
