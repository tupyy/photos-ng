import React from 'react';

interface LoadingProgressBarProps {
  /** Whether to show the loading bar */
  loading: boolean;
  /** Optional message to display */
  message?: string;
  /** Size variant */
  size?: 'small' | 'medium' | 'large';
  /** Color variant */
  variant?: 'primary' | 'secondary';
}

/**
 * Loading progress bar component with animated indeterminate progress
 * Perfect for indicating data fetching in infinite scroll scenarios
 */
const LoadingProgressBar: React.FC<LoadingProgressBarProps> = ({
  loading,
  message = 'Loading more photos...',
  size = 'medium',
  variant = 'primary',
}) => {
  if (!loading) return null;

  const sizeClasses = {
    small: 'h-1',
    medium: 'h-2',
    large: 'h-3',
  };

  const colorClasses = {
    primary: 'bg-blue-600',
    secondary: 'bg-gray-600',
  };

  const textSizeClasses = {
    small: 'text-xs',
    medium: 'text-sm',
    large: 'text-base',
  };

  return (
    <div className="flex flex-col items-center justify-center py-4 space-y-3">
      {message && (
        <div className={`text-gray-600 dark:text-gray-400 ${textSizeClasses[size]} font-medium`}>
          {message}
        </div>
      )}
      
      <div className="flex items-center justify-center">
        <div className={`animate-spin rounded-full border-2 border-gray-300 dark:border-gray-600 ${
          variant === 'primary' ? 'border-t-blue-600' : 'border-t-gray-600'
        } ${
          size === 'small' ? 'h-4 w-4' : size === 'medium' ? 'h-6 w-6' : 'h-8 w-8'
        }`} />
      </div>
    </div>
  );
};

export default LoadingProgressBar;
