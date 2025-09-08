import React, { useState, useEffect, useCallback, useRef } from 'react';
import { Media } from '@generated/models';

interface MediaViewerModalProps {
  isOpen: boolean;
  media: Media[];
  currentIndex: number;
  onClose: (currentMedia?: Media) => void;
  onIndexChange: (index: number) => void;
  thumbnailRect?: DOMRect;
}

const MediaViewerModal: React.FC<MediaViewerModalProps> = ({
  isOpen,
  media,
  currentIndex,
  onClose,
  onIndexChange,
  thumbnailRect,
}) => {
  const [isImageLoaded, setIsImageLoaded] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [isAnimating, setIsAnimating] = useState(false);

  // Touch/swipe state
  const [touchStart, setTouchStart] = useState<number | null>(null);
  const [touchEnd, setTouchEnd] = useState<number | null>(null);
  const modalRef = useRef<HTMLDivElement>(null);

  const currentMedia = media[currentIndex];

  // Reset loading state when media changes
  useEffect(() => {
    if (currentMedia) {
      setIsImageLoaded(false);
      setIsLoading(true);
    }
  }, [currentMedia]);


  // Handle modal opening animation
  useEffect(() => {
    if (isOpen && thumbnailRect) {
      setIsAnimating(false);
      // Small delay to ensure modal is rendered, then start animation
      const timer = setTimeout(() => {
        setIsAnimating(true);
      }, 50);
      return () => clearTimeout(timer);
    }
  }, [isOpen, thumbnailRect]);

  // Handle keyboard navigation
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (!isOpen) return;

      switch (event.key) {
        case 'Escape':
          onClose(currentMedia);
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

  // Swipe handlers
  const handleTouchStart = (e: React.TouchEvent) => {
    setTouchEnd(null);
    setTouchStart(e.targetTouches[0].clientX);
  };

  const handleTouchMove = (e: React.TouchEvent) => {
    setTouchEnd(e.targetTouches[0].clientX);
  };

  const handleTouchEnd = () => {
    if (!touchStart || !touchEnd) return;

    const distance = touchStart - touchEnd;
    const isLeftSwipe = distance > 50;
    const isRightSwipe = distance < -50;

    if (isLeftSwipe && currentIndex < media.length - 1) {
      handleNext();
    }
    if (isRightSwipe && currentIndex > 0) {
      handlePrevious();
    }
  };

  const handleImageLoad = () => {
    // Add artificial delay to simulate network latency
    setTimeout(() => {
      setIsImageLoaded(true);
      setIsLoading(false);
    }, 1500); // 1.5 second delay
  };

  const handleImageError = () => {
    setIsLoading(false);
    console.error('Failed to load image:', currentMedia?.content);
  };

  if (!isOpen || !currentMedia) return null;

  return (
    <div
      ref={modalRef}
      className="fixed inset-0 z-50"
      onTouchStart={handleTouchStart}
      onTouchMove={handleTouchMove}
      onTouchEnd={handleTouchEnd}
    >
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black bg-opacity-90 transition-opacity duration-300"
        onClick={() => onClose(currentMedia)}
      />

      {/* Close Button - Fixed at top */}
      <button
        onClick={() => onClose(currentMedia)}
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
          className="hidden md:block fixed left-4 top-1/2 transform -translate-y-1/2 z-50 p-4 rounded-full bg-black bg-opacity-70 text-white hover:bg-opacity-90 transition-colors shadow-lg"
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
          className="hidden md:block fixed right-4 top-1/2 transform -translate-y-1/2 z-50 p-4 rounded-full bg-black bg-opacity-70 text-white hover:bg-opacity-90 transition-colors shadow-lg"
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
          {/* Thumbnail - Always visible, same positioning as full image */}
          <img
            src={currentMedia.thumbnail}
            alt={currentMedia.filename}
            className="absolute inset-0 w-full h-full object-contain"
            onError={(e) => {
              console.error('Failed to load thumbnail:', currentMedia.thumbnail);
            }}
          />

          {/* Full Resolution Image - Overlays thumbnail when loaded */}
          <img
            src={currentMedia.content}
            alt={currentMedia.filename}
            className="absolute inset-0 w-full h-full object-contain"
            style={{ display: isImageLoaded ? 'initial' : 'none' }}
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

        {/* Desktop Only: Image Info - Below the image */}
        <div className="hidden md:block w-full max-w-4xl px-4">
          <div className="bg-black bg-opacity-70 rounded-lg p-3 md:p-4 text-white">
            <div className="flex items-center justify-between">
              <div className="flex-1">
                <h3 className="font-medium text-lg">{currentMedia.filename}</h3>
                <p className="text-sm text-gray-300 mt-1">
                  {currentIndex + 1} of {media.length}
                </p>
              </div>
              <div className="text-right text-sm text-gray-300">
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
