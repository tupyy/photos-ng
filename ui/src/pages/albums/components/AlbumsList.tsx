import React, { useMemo } from 'react';
import { Album as AlbumType } from '@shared/types/Album';
import Album from './Album';

export interface AlbumsListProps {
  albums: AlbumType[];
  loading?: boolean;
  error?: string | null;
  emptyStateTitle?: string;
  emptyStateMessage?: string;
  onSetThumbnail?: (albumId: string) => void;
  isSelectionMode?: boolean;
  selectedIds?: Set<string>;
  onToggleSelection?: (albumId: string) => void;
}

/**
 * AlbumsList - Displays albums in a grid layout
 *
 * Features:
 * - Alphabetically sorted album cards
 * - Selection mode controlled by parent
 * - Supports thumbnail selection
 */
const AlbumsList: React.FC<AlbumsListProps> = ({
  albums,
  loading = false,
  error = null,
  onSetThumbnail,
  isSelectionMode = false,
  selectedIds = new Set(),
  onToggleSelection,
}) => {
  const sortedAlbums = useMemo(() => {
    const albumsArray = albums || [];
    return [...albumsArray].sort((a, b) => {
      const nameA = a.name || '';
      const nameB = b.name || '';
      return nameA.localeCompare(nameB, undefined, {
        numeric: true,
        sensitivity: 'base',
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
    <div className="album-grid-container">
      {sortedAlbums.map((album: AlbumType) => (
        <Album
          key={album.id}
          album={album}
          isSelectionMode={isSelectionMode}
          isSelected={selectedIds.has(album.id)}
          onSelectionToggle={onToggleSelection}
          onSetThumbnail={onSetThumbnail}
        />
      ))}
    </div>
  );
};

export default AlbumsList;
