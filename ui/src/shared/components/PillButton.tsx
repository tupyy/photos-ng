import React from 'react';

export interface PillButtonProps {
  onClick: () => void;
  disabled?: boolean;
  variant?: 'default' | 'danger' | 'success';
  children: React.ReactNode;
  className?: string;
}

/**
 * PillButton - A rounded pill-shaped button with consistent styling
 *
 * Variants:
 * - default: gray border, hover to black/white
 * - danger: red border and text
 * - success: green border and text
 */
const PillButton: React.FC<PillButtonProps> = ({
  onClick,
  disabled = false,
  variant = 'default',
  children,
  className = '',
}) => {
  const baseClasses = 'flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-full border-2 bg-transparent focus:outline-none transition-colors';

  const variantClasses = {
    default: 'border-gray-400 text-gray-400 hover:text-black hover:border-black dark:hover:text-white dark:hover:border-white',
    danger: 'border-red-400 text-red-400 hover:text-red-600 hover:border-red-600 dark:hover:text-red-200 dark:hover:border-red-200',
    success: 'border-green-400 text-green-400 hover:text-green-600 hover:border-green-600 dark:hover:text-green-200 dark:hover:border-green-200',
  };

  const disabledClasses = disabled ? 'opacity-50 cursor-not-allowed' : '';

  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={`${baseClasses} ${variantClasses[variant]} ${disabledClasses} ${className}`}
    >
      {children}
    </button>
  );
};

export default PillButton;
