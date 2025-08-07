import React from 'react';

interface BuildInfoProps {
  variant?: 'plain' | 'link';
}

const BuildInfo: React.FC<BuildInfoProps> = ({ variant = 'plain' }) => {
  const gitCommit = process.env.GIT_COMMIT || 'unknown';
  
  return (
    <div 
      className="flex items-center space-x-1 text-xs text-gray-500 dark:text-gray-400"
      title={`Build commit: ${gitCommit}`}
    >
      <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg">
        <path fillRule="evenodd" d="M12.316 3.051a1 1 0 01.633 1.265l-4 12a1 1 0 11-1.898-.632l4-12a1 1 0 011.265-.633zM5.707 6.293a1 1 0 010 1.414L3.414 10l2.293 2.293a1 1 0 11-1.414 1.414l-3-3a1 1 0 010-1.414l3-3a1 1 0 011.414 0zm8.586 0a1 1 0 011.414 0l3 3a1 1 0 010 1.414l-3 3a1 1 0 11-1.414-1.414L16.586 10l-2.293-2.293a1 1 0 010-1.414z" clipRule="evenodd" />
      </svg>
      <span>{gitCommit}</span>
    </div>
  );
};

export default BuildInfo;
