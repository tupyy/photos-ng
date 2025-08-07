import React from 'react';
import Navbar from './Navbar';

export interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
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