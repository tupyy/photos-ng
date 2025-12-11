import React, { useMemo } from 'react';
import { Album as AlbumType } from '@shared/types/Album';
import Album from './Album';

export interface SubAlbumsListProps {
  albums: AlbumType[];
  loading?: boolean;
  isSelectionMode?: boolean;
  selectedIds?: Set<string>;
  onToggleSelection?: (albumId: string) => void;
}

/**
 * SubAlbumsList - Displays sub-albums using the same card style as the main albums page
 *
 * Features:
 * - Uses the same Album card component as the main page
 * - Grid layout matching the main albums page
 * - Supports selection mode with checkboxes
 */
const SubAlbumsList: React.FC<SubAlbumsListProps> = ({
  albums,
  loading = false,
  isSelectionMode = false,
  selectedIds = new Set(),
  onToggleSelection,
}) => {
  // Sort albums alphabetically
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

  // Don't render if no albums
  if (!sortedAlbums || sortedAlbums.length === 0) {
    return null;
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-32">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
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
        />
      ))}
    </div>
  );
};

export default SubAlbumsList;
