import React from 'react';
import { performLogout } from './authUtils';

interface LogoutButtonProps {
  className?: string;
  children?: React.ReactNode;
}

/**
 * Sign Out button that triggers the full logout flow:
 * Keycloak logout -> Envoy signout -> App home
 */
export function LogoutButton({ className, children }: LogoutButtonProps) {
  const handleClick = () => {
    performLogout();
  };

  return (
    <button
      type="button"
      onClick={handleClick}
      className={className}
    >
      {children ?? 'Sign Out'}
    </button>
  );
}
