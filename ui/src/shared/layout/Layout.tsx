import React, { useEffect } from 'react';
import Navbar from './Navbar';
import { useAppDispatch } from '@shared/store';
import { fetchCurrentUser } from '@shared/reducers';

export interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const dispatch = useAppDispatch();

  // Fetch current user on app start
  useEffect(() => {
    dispatch(fetchCurrentUser());
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