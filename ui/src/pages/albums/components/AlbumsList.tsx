import React, { useEffect } from 'react';
import { useAlbumsApi } from '@hooks/useApi';
import { Album } from '@generated/models';

const AlbumsList: React.FC = () => {
  const {
    albums,
    loading,
    error,
    total,
    fetchAlbums,
    syncStatus,
  } = useAlbumsApi();

  useEffect(() => {
    fetchAlbums({ limit: 20, offset: 0 });
  }, [fetchAlbums]);

  if (loading && albums.length === 0) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-md p-4">
        <p className="text-red-600">Error: {error}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {albums.map((album: Album) => (
          <div
            key={album.id}
            className="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow"
          >
            {album.thumbnail && (
              <div className="h-48 bg-gray-200">
                <img
                  src={album.thumbnail}
                  alt={album.name}
                  className="w-full h-full object-cover"
                />
              </div>
            )}

            <div className="p-4">
              <h3 className="text-lg font-semibold text-gray-900 mb-2">
                {album.name}
              </h3>

              {album.description && (
                <p className="text-gray-600 text-sm mb-3">
                  {album.description}
                </p>
              )}

              <div className="flex items-center justify-between text-sm text-gray-500 mb-4">
                <span>{album.media?.length || 0} photos</span>
                <span>{album.children?.length || 0} subalbums</span>
              </div>
            </div>
          </div>
        ))}
      </div>

      {albums.length === 0 && !loading && (
        <div className="text-center py-12">
          <p className="text-gray-500 text-lg">No albums found.</p>
        </div>
      )}
    </div>
  );
};

export default AlbumsList;
