import React from 'react';
import { Routes, Route } from 'react-router-dom';
import { Layout } from '@shared/layout';
import { AlbumsPage, TimelinePage } from '@app/pages';

function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<TimelinePage />} />
        <Route path="/albums/:id?" element={<AlbumsPage />} />
        {/* Fallback route for any unmatched paths */}
        <Route path="*" element={<TimelinePage />} />
      </Routes>
    </Layout>
  );
}

export default App;
