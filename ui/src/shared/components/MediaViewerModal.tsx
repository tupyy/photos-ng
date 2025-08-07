import React, { useState, useEffect, useCallback } from 'react';
import { Media } from '@generated/models';

interface MediaViewerModalProps {
  isOpen: boolean;
  media: Media[];
  currentIndex: number;
  onClose: () => void;
  onIndexChange: (index: number) => void;
}

const MediaViewerModal: React.FC<MediaViewerModalProps> = ({
  isOpen,
  media,
  currentIndex,
  onClose,
  onIndexChange,
}) => {
  const [isImageLoaded, setIsImageLoaded] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  const currentMedia = media[currentIndex];

  // Reset loading state when media changes
  useEffect(() => {
    if (currentMedia) {
      setIsImageLoaded(false);
      setIsLoading(true);
    }
  }, [currentMedia]);

  // Handle keyboard navigation
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (!isOpen) return;

      switch (event.key) {
        case 'Escape':
          onClose();
          break;
        case 'ArrowLeft':
          event.preventDefault();
          handlePrevious();
          break;
        case 'ArrowRight':
          event.preventDefault();
          handleNext();
          break;
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleKeyDown);
      // Prevent body scroll when modal is open
      document.body.style.overflow = 'hidden';
    }

    return () => {
      document.removeEventListener('keydown', handleKeyDown);
      document.body.style.overflow = 'unset';
    };
  }, [isOpen, onClose]);

  const handlePrevious = useCallback(() => {
    if (currentIndex > 0) {
      onIndexChange(currentIndex - 1);
    }
  }, [currentIndex, onIndexChange]);

  const handleNext = useCallback(() => {
    if (currentIndex < media.length - 1) {
      onIndexChange(currentIndex + 1);
    }
  }, [currentIndex, media.length, onIndexChange]);

  const handleImageLoad = () => {
    setIsImageLoaded(true);
    setIsLoading(false);
  };

  const handleImageError = () => {
    setIsLoading(false);
    console.error('Failed to load image:', currentMedia?.content);
  };

  if (!isOpen || !currentMedia) return null;

  return (
    <div className="fixed inset-0 z-50">
      {/* Backdrop */}
      <div 
        className="absolute inset-0 bg-black bg-opacity-90 transition-opacity duration-300"
        onClick={onClose}
      />

      {/* Close Button - Fixed at top */}
      <button
        onClick={onClose}
        className="fixed top-4 right-4 z-50 p-3 rounded-full bg-black bg-opacity-70 text-white hover:bg-opacity-90 transition-colors shadow-lg"
      >
        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>

      {/* Previous Button - Fixed on left side (desktop only) */}
      {currentIndex > 0 && (
        <button
          onClick={handlePrevious}
          className="hidden md:fixed left-4 top-1/2 transform -translate-y-1/2 z-50 p-4 rounded-full bg-black bg-opacity-70 text-white hover:bg-opacity-90 transition-colors shadow-lg"
        >
          <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 19l-7-7 7-7" />
          </svg>
        </button>
      )}

      {/* Next Button - Fixed on right side (desktop only) */}
      {currentIndex < media.length - 1 && (
        <button
          onClick={handleNext}
          className="hidden md:fixed right-4 top-1/2 transform -translate-y-1/2 z-50 p-4 rounded-full bg-black bg-opacity-70 text-white hover:bg-opacity-90 transition-colors shadow-lg"
        >
          <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5l7 7-7 7" />
          </svg>
        </button>
      )}

      {/* Modal Content */}
      <div className="relative w-full h-full flex flex-col items-center justify-center p-4 pt-16">

        {/* Image Container */}
        <div className="relative flex-1 w-full flex items-center justify-center min-h-0">
          {/* Loading State - Show thumbnail while loading */}
          {isLoading && (
            <div className="absolute inset-0 flex items-center justify-center">
              <img
                src={currentMedia.thumbnail}
                alt={`Loading ${currentMedia.filename}`}
                className="max-w-full max-h-full object-contain blur-sm opacity-50"
                onError={(e) => {
                  // Hide loading state if thumbnail also fails
                  console.error('Failed to load thumbnail:', currentMedia.thumbnail);
                }}
              />
              <div className="absolute inset-0 flex items-center justify-center">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-white"></div>
              </div>
            </div>
          )}

          {/* Full Resolution Image */}
          <img
            src={currentMedia.content}
            alt={currentMedia.filename}
            className={`max-w-full max-h-full object-contain transition-opacity duration-300 ${
              isImageLoaded ? 'opacity-100' : 'opacity-0'
            }`}
            onLoad={handleImageLoad}
            onError={handleImageError}
          />

          {/* Fallback content if image fails to load */}
          {!isLoading && !isImageLoaded && (
            <div className="flex flex-col items-center justify-center text-white p-8">
              <svg className="w-16 h-16 mb-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6 6h.01M6 20h.01m-.01 4h.01m4.01 0h.01m0 4h.01M10 28h.01m4.01 0h.01m0 4h.01M16 32h.01M8 36a4 4 0 004 4h8a4 4 0 004-4v-8a4 4 0 00-4-4h-8a4 4 0 00-4 4v8z" />
              </svg>
              <p className="text-lg font-medium">Failed to load image</p>
              <p className="text-sm text-gray-300 mt-2">{currentMedia.filename}</p>
              <p className="text-xs text-gray-400 mt-1">URL: {currentMedia.content}</p>
            </div>
          )}
        </div>

        {/* Mobile Navigation Buttons - Below image on mobile */}
        <div className="md:hidden flex justify-center gap-4 mt-4 mb-2">
          <button
            onClick={handlePrevious}
            disabled={currentIndex === 0}
            className="p-2 rounded-full bg-black bg-opacity-70 text-white hover:bg-opacity-90 transition-colors shadow-lg disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
          <button
            onClick={handleNext}
            disabled={currentIndex === media.length - 1}
            className="p-2 rounded-full bg-black bg-opacity-70 text-white hover:bg-opacity-90 transition-colors shadow-lg disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5l7 7-7 7" />
            </svg>
          </button>
        </div>

        {/* Image Info - Below the image */}
        <div className="w-full max-w-4xl px-4">
          <div className="bg-black bg-opacity-70 rounded-lg p-3 md:p-4 text-white">
            <div className="flex items-center justify-between">
              <div className="flex-1">
                <h3 className="font-medium text-xs lg:text-lg">{currentMedia.filename}</h3>
                <p className="text-xs md:text-sm text-gray-300 mt-1">
                  {currentIndex + 1} of {media.length}
                </p>
              </div>
              <div className="text-right text-xs md:text-sm text-gray-300">
                <p>{currentMedia.type}</p>
                <p>{new Date(currentMedia.capturedAt).toLocaleDateString()}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default MediaViewerModal;
