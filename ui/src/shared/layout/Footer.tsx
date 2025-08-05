import React from 'react';

const Footer: React.FC = () => {
  const currentYear = new Date().getFullYear();

  return (
    <footer className="w-full backdrop-blur flex-none transition-colors duration-500 supports-backdrop-blur:bg-white/60 dark:bg-slate-900">
      <div className="border-t lg:border-slate-900/10 dark:border-slate-50/[0.06]">
        <div className="flex flex-wrap items-center justify-between mx-auto p-3 lg:px-8">
          {/* Left - Copyright */}
          <div className="flex items-center">
            <p className="text-sm text-gray-600 dark:text-gray-400">
              © {currentYear} Photos NG. All rights reserved.
            </p>
          </div>

          {/* Right - Git Commit */}
          <div className="flex items-center space-x-2">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth="1.5" stroke="currentColor" className="w-4 h-4 text-gray-500 dark:text-gray-400">
              <path strokeLinecap="round" strokeLinejoin="round" d="M17.25 6.75 22.5 12l-5.25 5.25m-10.5 0L1.5 12l5.25-5.25m7.5-3-4.5 16.5" />
            </svg>
            <span className="text-xs text-gray-500 dark:text-gray-400 font-mono">
              a1b2c3d
            </span>
            <span className="text-gray-300 dark:text-gray-600">•</span>
            <span className="text-xs text-gray-500 dark:text-gray-400">
              2 hours ago
            </span>
          </div>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
