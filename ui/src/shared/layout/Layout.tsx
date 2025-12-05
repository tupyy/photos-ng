import React, { useEffect } from 'react';
import Navbar from './Navbar';
import { useAppDispatch } from '@shared/store';
import { fetchCurrentUser } from '@shared/reducers';
import { AUTHZ_ENABLED } from '@shared/config';

export interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const dispatch = useAppDispatch();

  // Fetch current user on app start (only when authz is enabled)
  useEffect(() => {
    if (AUTHZ_ENABLED) {
      dispatch(fetchCurrentUser());
    }
  }, [dispatch]);

  return (
    <div className="min-h-screen flex flex-col bg-white dark:bg-slate-900 transition-colors">
      <Navbar />
      
      <main className="flex-1 bg-gray-50 dark:bg-slate-900 transition-colors">
        {children}
      </main>
    </div>
  );
};

export default Layout;