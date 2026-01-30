import { Suspense, lazy } from 'react';
import { Routes, Route } from 'react-router-dom';
import { Layout } from '@shared/layout';
import { ThumbnailProvider } from '@shared/contexts';

// Lazy load pages for code splitting
const AlbumsPage = lazy(() => import('@app/pages/albums'));
const TimelinePage = lazy(() => import('@app/pages/timeline'));
const UploadMediaPage = lazy(() => import('@app/pages/upload'));

// Loading fallback component
const PageLoader = () => (
  <div className="flex items-center justify-center min-h-[50vh]">
    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
  </div>
);

function App() {
  return (
    <Layout>
      <Suspense fallback={<PageLoader />}>
        <Routes>
          <Route path="/" element={<TimelinePage />} />
          <Route path="/albums/:id?" element={
            <ThumbnailProvider>
              <AlbumsPage />
            </ThumbnailProvider>
          } />
          <Route path="/upload/:albumId" element={<UploadMediaPage />} />
          {/* Fallback route for any unmatched paths */}
          <Route path="*" element={<TimelinePage />} />
        </Routes>
      </Suspense>
    </Layout>
  );
}

export default App;
