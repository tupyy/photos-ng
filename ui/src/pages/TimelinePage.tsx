import React from 'react';

const TimelinePage: React.FC = () => {
  return (
    <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
      <div className="px-4 py-6 sm:px-0">
        <div className="text-center py-12">
          <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100 mb-4">
            <svg 
              xmlns="http://www.w3.org/2000/svg" 
              fill="none" 
              viewBox="0 0 24 24" 
              strokeWidth="1.5" 
              stroke="currentColor" 
              className="w-6 h-6 text-green-600"
            >
              <path 
                strokeLinecap="round" 
                strokeLinejoin="round" 
                d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" 
              />
            </svg>
          </div>
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Timeline View</h2>
          <p className="text-gray-600 mb-8">
            Explore your photos organized chronologically. Navigate through your memories by date.
          </p>
          <div className="space-y-4">
            <div className="bg-gray-50 rounded-lg p-6 text-left max-w-2xl mx-auto">
              <h3 className="font-semibold text-gray-900 mb-2">Coming Soon:</h3>
              <ul className="text-sm text-gray-600 space-y-1">
                <li>• Interactive timeline with year/month navigation</li>
                <li>• Photo clusters by date taken</li>
                <li>• Quick jump to specific dates</li>
                <li>• Timeline statistics and insights</li>
                <li>• Memories and highlights</li>
                <li>• Date range filtering</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default TimelinePage;