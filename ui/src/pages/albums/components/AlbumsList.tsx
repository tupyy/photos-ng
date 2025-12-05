import React, { useMemo, useState } from 'react';
import { Album as AlbumType } from '@shared/types/Album';
import { useAlbumsApi } from '@shared/hooks/useApi';
import { ConfirmDeleteModal } from '@app/shared/components';
import Album from './Album';

export interface AlbumsListProps {
  albums: AlbumType[];
  loading?: boolean;
  error?: string | null;
  emptyStateTitle?: string;
  emptyStateMessage?: string;
  onSetThumbnail?: (albumId: string) => void;
}

const AlbumsList: React.FC<AlbumsListProps> = ({ albums, loading = false, error = null, onSetThumbnail }) => {
  const [isSelectionMode, setIsSelectionMode] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const { selectedAlbumIds, toggleAlbumSelection, selectAllAlbums, clearAlbumSelection, deleteAlbum } = useAlbumsApi();
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

  // Selection handlers
  const toggleSelectionMode = () => {
    setIsSelectionMode(!isSelectionMode);
    clearAlbumSelection();
  };

  const handleSelectAll = () => {
    if (selectedAlbumIds.length === sortedAlbums.length) {
      clearAlbumSelection();
    } else {
      selectAllAlbums();
    }
  };

  const handleDeleteSelected = () => {
    if (selectedAlbumIds.length === 0) return;
    setShowDeleteModal(true);
  };

  const confirmDelete = async () => {
    setIsDeleting(true);
    setShowDeleteModal(false);

    try {
      // Delete albums one by one
      const deletePromises = selectedAlbumIds.map((id) => deleteAlbum(id));
      await Promise.all(deletePromises);

      // Clear selection and exit selection mode
      clearAlbumSelection();
      setIsSelectionMode(false);
    } catch (error) {
      console.error('Failed to delete albums:', error);
    } finally {
      setIsDeleting(false);
    }
  };

  const cancelDelete = () => {
    setShowDeleteModal(false);
  };

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
      {/* Selection Toolbar - Only show if there are albums */}
      {sortedAlbums.length > 0 && (
        <div className="flex items-center justify-end">
          <div className="flex items-center space-x-2">
            {isSelectionMode ? (
              <>
                <button
                  onClick={handleSelectAll}
                  disabled={isDeleting}
                  className="px-4 py-2 text-sm font-medium border-2 border-gray-400 rounded-full text-gray-400 bg-transparent hover:text-black hover:border-black dark:hover:text-white dark:hover:border-white focus:outline-none transition-colors disabled:opacity-50"
                >
                  {selectedAlbumIds.length === sortedAlbums.length ? 'Deselect All' : 'Select All'}
                </button>
                <button
                  onClick={() => clearAlbumSelection()}
                  disabled={isDeleting}
                  className="px-4 py-2 text-sm font-medium border-2 border-gray-400 rounded-full text-gray-400 bg-transparent hover:text-black hover:border-black dark:hover:text-white dark:hover:border-white focus:outline-none transition-colors disabled:opacity-50"
                >
                  Clear
                </button>

                {selectedAlbumIds.length > 0 && (
                  <button
                    onClick={handleDeleteSelected}
                    disabled={isDeleting}
                    className="px-4 py-2 text-sm font-medium border-2 border-red-400 rounded-full text-red-400 bg-transparent hover:text-red-600 hover:border-red-600 dark:hover:text-red-200 dark:hover:border-red-200 focus:outline-none transition-colors disabled:opacity-50"
                  >
                    {isDeleting ? 'Deleting...' : `Delete (${selectedAlbumIds.length})`}
                  </button>
                )}
                <button
                  onClick={toggleSelectionMode}
                  disabled={isDeleting}
                  className="px-4 py-2 text-sm font-medium border-2 border-gray-400 rounded-full text-gray-400 bg-transparent hover:text-black hover:border-black dark:hover:text-white dark:hover:border-white focus:outline-none transition-colors disabled:opacity-50"
                >
                  Cancel
                </button>
              </>
            ) : (
              <button
                onClick={toggleSelectionMode}
                className="flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-full border-2 border-gray-400 text-gray-400 bg-transparent hover:text-black hover:border-black dark:hover:text-white dark:hover:border-white focus:outline-none transition-colors"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                Select
              </button>
            )}
          </div>
        </div>
      )}

      {/* Albums Gallery */}
      <div className="album-grid-container">
        {sortedAlbums.map((album: AlbumType) => (
          <Album
            key={album.id}
            album={album}
            isSelectionMode={isSelectionMode}
            isSelected={selectedAlbumIds.includes(album.id)}
            onSelectionToggle={toggleAlbumSelection}
            onSetThumbnail={onSetThumbnail}
          />
        ))}
      </div>

      {/* Delete Confirmation Modal */}
      <ConfirmDeleteModal
        isOpen={showDeleteModal}
        itemCount={selectedAlbumIds.length}
        itemType="Album"
        isDeleting={isDeleting}
        onConfirm={confirmDelete}
        onCancel={cancelDelete}
      />
    </div>
  );
};

export default AlbumsList;
