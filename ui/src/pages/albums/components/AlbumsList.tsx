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
                  className="px-3 py-1 text-sm border border-gray-300 rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
                >
                  {selectedAlbumIds.length === sortedAlbums.length ? 'Deselect All' : 'Select All'}
                </button>
                <button
                  onClick={() => clearAlbumSelection()}
                  disabled={isDeleting}
                  className="px-3 py-1 text-sm border border-gray-300 rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
                >
                  Clear
                </button>

                {selectedAlbumIds.length > 0 && (
                  <button
                    onClick={handleDeleteSelected}
                    disabled={isDeleting}
                    className="px-3 py-1 text-sm border border-red-300 rounded-md text-red-700 bg-red-50 hover:bg-red-100 focus:outline-none focus:ring-2 focus:ring-red-500 disabled:opacity-50 dark:bg-red-900/20 dark:border-red-600 dark:text-red-300 dark:hover:bg-red-900/30"
                  >
                    {isDeleting ? 'Deleting...' : `Delete (${selectedAlbumIds.length})`}
                  </button>
                )}
                <button
                  onClick={toggleSelectionMode}
                  disabled={isDeleting}
                  className="px-3 py-1 text-sm border border-gray-300 rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
                >
                  Cancel
                </button>
              </>
            ) : (
              <button
                onClick={toggleSelectionMode}
                className="px-3 py-1 text-sm border border-gray-300 rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-700"
              >
                Select Albums
              </button>
            )}
          </div>
        </div>
      )}

      {/* Albums Gallery */}
      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-2">
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
