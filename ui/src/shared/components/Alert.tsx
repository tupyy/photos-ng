import React, { useEffect, useState } from 'react';

type AlertType = 'success' | 'error' | 'warning' | 'info';

interface AlertProps {
  type: AlertType;
  title?: string;
  message: string;
  isVisible: boolean;
  onDismiss?: () => void;
  autoDismiss?: boolean;
  autoDismissTimeout?: number;
  dismissible?: boolean;
}

const Alert: React.FC<AlertProps> = ({
  type,
  title,
  message,
  isVisible,
  onDismiss,
  autoDismiss = true,
  autoDismissTimeout = 5000,
  dismissible = true,
}) => {
  const [show, setShow] = useState(isVisible);

  useEffect(() => {
    setShow(isVisible);
  }, [isVisible]);

  useEffect(() => {
    if (show && autoDismiss) {
      const timer = setTimeout(() => {
        handleDismiss();
      }, autoDismissTimeout);

      return () => clearTimeout(timer);
    }
  }, [show, autoDismiss, autoDismissTimeout]);

  const handleDismiss = () => {
    setShow(false);
    if (onDismiss) {
      onDismiss();
    }
  };

  if (!show) return null;

  const getAlertStyles = () => {
    switch (type) {
      case 'success':
        return {
          container: 'text-green-800 bg-green-50 dark:bg-gray-800 dark:text-green-400',
          border: 'border-green-300 dark:border-green-600',
          icon: (
            <svg className="shrink-0 inline w-4 h-4 me-3" aria-hidden="true" fill="currentColor" viewBox="0 0 20 20">
              <path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5ZM9.5 4a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3ZM12 15H8a1 1 0 0 1 0-2h1v-3H8a1 1 0 0 1 0-2h2a1 1 0 0 1 1 1v4h1a1 1 0 0 1 0 2Z"/>
            </svg>
          ),
        };
      case 'error':
        return {
          container: 'text-red-800 bg-red-50 dark:bg-gray-800 dark:text-red-400',
          border: 'border-red-300 dark:border-red-600',
          icon: (
            <svg className="shrink-0 inline w-4 h-4 me-3" aria-hidden="true" fill="currentColor" viewBox="0 0 20 20">
              <path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5ZM9.5 4a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3ZM12 15H8a1 1 0 0 1 0-2h1v-3H8a1 1 0 0 1 0-2h2a1 1 0 0 1 1 1v4h1a1 1 0 0 1 0 2Z"/>
            </svg>
          ),
        };
      case 'warning':
        return {
          container: 'text-yellow-800 bg-yellow-50 dark:bg-gray-800 dark:text-yellow-300',
          border: 'border-yellow-300 dark:border-yellow-600',
          icon: (
            <svg className="shrink-0 inline w-4 h-4 me-3" aria-hidden="true" fill="currentColor" viewBox="0 0 20 20">
              <path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5ZM9.5 4a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3ZM12 15H8a1 1 0 0 1 0-2h1v-3H8a1 1 0 0 1 0-2h2a1 1 0 0 1 1 1v4h1a1 1 0 0 1 0 2Z"/>
            </svg>
          ),
        };
      case 'info':
      default:
        return {
          container: 'text-blue-800 bg-blue-50 dark:bg-gray-800 dark:text-blue-400',
          border: 'border-blue-300 dark:border-blue-600',
          icon: (
            <svg className="shrink-0 inline w-4 h-4 me-3" aria-hidden="true" fill="currentColor" viewBox="0 0 20 20">
              <path d="M10 .5a9.5 9.5 0 1 0 9.5 9.5A9.51 9.51 0 0 0 10 .5ZM9.5 4a1.5 1.5 0 1 1 0 3 1.5 1.5 0 0 1 0-3ZM12 15H8a1 1 0 0 1 0-2h1v-3H8a1 1 0 0 1 0-2h2a1 1 0 0 1 1 1v4h1a1 1 0 0 1 0 2Z"/>
            </svg>
          ),
        };
    }
  };

  const styles = getAlertStyles();

  return (
    <div className="fixed top-4 right-4 z-50 max-w-sm w-full">
      <div
        className={`flex items-center p-4 mb-4 text-sm rounded-lg border ${styles.container} ${styles.border}`}
        role="alert"
      >
        {styles.icon}
        <span className="sr-only">{type.charAt(0).toUpperCase() + type.slice(1)}</span>
        <div className="flex-1">
          {title && <span className="font-medium">{title}</span>}
          {title && message && <span> </span>}
          <span>{message}</span>
        </div>
        {dismissible && (
          <button
            type="button"
            onClick={handleDismiss}
            className={`ms-auto -mx-1.5 -my-1.5 p-1.5 rounded-lg focus:ring-2 inline-flex items-center justify-center h-8 w-8 ${
              type === 'success'
                ? 'bg-green-50 text-green-500 focus:ring-green-400 hover:bg-green-200 dark:bg-gray-800 dark:text-green-400 dark:hover:bg-gray-700'
                : type === 'error'
                ? 'bg-red-50 text-red-500 focus:ring-red-400 hover:bg-red-200 dark:bg-gray-800 dark:text-red-400 dark:hover:bg-gray-700'
                : type === 'warning'
                ? 'bg-yellow-50 text-yellow-500 focus:ring-yellow-400 hover:bg-yellow-200 dark:bg-gray-800 dark:text-yellow-300 dark:hover:bg-gray-700'
                : 'bg-blue-50 text-blue-500 focus:ring-blue-400 hover:bg-blue-200 dark:bg-gray-800 dark:text-blue-400 dark:hover:bg-gray-700'
            }`}
            aria-label="Close"
          >
            <span className="sr-only">Close</span>
            <svg className="w-3 h-3" aria-hidden="true" fill="none" viewBox="0 0 14 14">
              <path
                stroke="currentColor"
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="m1 1 6 6m0 0 6 6M7 7l6-6M7 7l-6 6"
              />
            </svg>
          </button>
        )}
      </div>
    </div>
  );
};

export default Alert;
