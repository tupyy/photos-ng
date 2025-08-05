import React, { useMemo } from 'react';
import { Album as AlbumType } from '@shared/types/Album';
import Album from './Album';

export interface AlbumsListProps {
  albums: AlbumType[];
  loading?: boolean;
  error?: string | null;
  emptyStateTitle?: string;
  emptyStateMessage?: string;
}

const AlbumsList: React.FC<AlbumsListProps> = ({
  albums,
  loading = false,
  error = null,
}) => {
  const sortedAlbums = useMemo(() => {
    // Ensure albums is an array to prevent undefined errors
    const albumsArray = albums || [];
    return [...albumsArray].sort((a, b) => {
      // Handle null/undefined names
      const nameA = a.name || '';
      const nameB = b.name || '';

      // Use localeCompare for proper alphabetical sorting with locale support
      return nameA.localeCompare(nameB, undefined, {
        numeric: true,
        sensitivity: 'base', // Case-insensitive
      });
    });
  }, [albums]);

  if (loading && (!albums || albums.length === 0)) {
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
      {/* Albums Gallery */}
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-4">
        {sortedAlbums.map((album: AlbumType) => (
          <Album key={album.id} album={album} />
        ))}
      </div>
    </div>
  );
};

export default AlbumsList;
