/**
 * Albums Page Component
 * 
 * Main page for album management in the Photos NG application.
 * Provides functionality for:
 * - Viewing album hierarchies (parent/child albums)
 * - Creating new albums
 * - Editing album descriptions
 * - Displaying media within albums
 * - Managing album thumbnails
 * 
 * The component supports both root-level album listing and individual album views
 * with media galleries, depending on the URL parameter.
 */

import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAppDispatch, useAppSelector, selectAlbumsCreateFormOpen, selectCurrentAlbum } from '@shared/store';
import { setPageActive, setCreateFormOpen, fetchAlbumById, setCurrentAlbum } from '@shared/reducers/albumsSlice';
import { useAlbumsApi, useMediaApi } from '@shared/hooks/useApi';
import { ListMediaSortByEnum, ListMediaSortOrderEnum } from '@generated/api/media-api';
import { Album } from '@shared/types/Album';
import AlbumsList from './components/AlbumsList';
import CreateAlbumForm from './components/CreateAlbumForm';
import MediaGallery from './components/MediaGallery';

const AlbumsPage: React.FC = () => {
  // URL parameters and navigation
  const { id } = useParams<{ id: string }>(); // Album ID from URL (undefined for root albums view)
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  
  // Redux state selectors
  const isCreateFormOpen = useAppSelector(selectAlbumsCreateFormOpen);
  const currentAlbum = useAppSelector(selectCurrentAlbum);
  
  // API hooks for data fetching and operations
  const { albums, loading, error, fetchAlbums, fetchAlbumById: fetchAlbumByIdApi, updateAlbum } = useAlbumsApi();

  // Local state for paginated media
  const [paginatedMedia, setPaginatedMedia] = useState<any[]>([]);
  const [mediaLoading, setMediaLoading] = useState(false);
  const [mediaError, setMediaError] = useState<string | null>(null);

  // Local state for inline description editing
  const [isEditingDescription, setIsEditingDescription] = useState(false);
  const [editedDescription, setEditedDescription] = useState('');

  // Media pagination state
  const [currentPage, setCurrentPage] = useState(1);
  const pageSize = 100; // Number of media items per page

  /**
   * Main effect for page initialization and data fetching
   * Runs when component mounts or when album ID or pagination changes
   */
  useEffect(() => {
    // Set page as active when component mounts
    dispatch(setPageActive(true));

    if (id) {
      // Viewing a specific album - fetch album details (which includes all media)
      fetchAlbumByIdApi(id);
    } else {
      // Viewing root albums list - fetch all albums (no pagination)
      fetchAlbums({ limit: 1000, offset: 0 });
      // Clear current album when navigating to root
      dispatch(setCurrentAlbum(null));
    }

    // Set page as inactive when component unmounts
    return () => {
      dispatch(setPageActive(false));
    };
  }, [dispatch, id, fetchAlbums, fetchAlbumByIdApi]);

  /**
   * Initialize edited description when currentAlbum changes
   * Ensures the input field shows the current album description
   */
  useEffect(() => {
    if (currentAlbum) {
      setEditedDescription(currentAlbum.description || '');
    }
  }, [currentAlbum]);

  /**
   * Handles closing the create album form modal
   */
  const handleCreateFormClose = () => {
    dispatch(setCreateFormOpen(false));
  };

  /**
   * Handles successful album creation
   * Navigates to the newly created album
   * @param albumId - The ID of the newly created album
   */
  const handleCreateAlbumSuccess = (albumId: string) => {
    console.log('Album created successfully:', albumId);
    // Navigate to the created album
    navigate(`/albums/${albumId}`);
  };

  /**
   * Handles navigation back to parent album or root
   * Uses the parentHref from current album to determine navigation target
   */
  const handleBackToParent = () => {
    if (currentAlbum?.parentHref) {
      // Extract parent ID from parentHref and navigate to it
      const parentId = currentAlbum.parentHref.split('/').pop();
      navigate(`/albums/${parentId}`);
    } else {
      // Navigate to root albums list
      navigate('/albums');
    }
  };

  const handleDescriptionEdit = () => {
    setIsEditingDescription(true);
  };

  const handleDescriptionSave = async () => {
    if (!currentAlbum || !id) return;

    try {
      await updateAlbum(id, { description: editedDescription });
      setIsEditingDescription(false);
      // Refresh the album data to get the updated description
      fetchAlbumByIdApi(id);
    } catch (error) {
      console.error('Failed to update album description:', error);
    }
  };

  const handleDescriptionCancel = () => {
    setEditedDescription(currentAlbum?.description || '');
    setIsEditingDescription(false);
  };

  const handleDescriptionKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleDescriptionSave();
    } else if (e.key === 'Escape') {
      handleDescriptionCancel();
    }
  };

  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  const handleMediaDeleted = () => {
    if (id) {
      // Refresh album data to get updated media list and check if thumbnail was affected
      fetchAlbumByIdApi(id);
      // The useEffect will automatically refresh the paginated media when currentAlbum changes
    }
  };

  // Reset page when album changes
  useEffect(() => {
    setCurrentPage(1);
  }, [id]);

  // Fetch media objects for current page when album or page changes
  useEffect(() => {
    const fetchPaginatedMedia = async () => {
      if (!currentAlbum?.media || currentAlbum.media.length === 0) {
        setPaginatedMedia([]);
        return;
      }

      // Calculate pagination for media hrefs
      const startIndex = (currentPage - 1) * pageSize;
      const endIndex = Math.min(startIndex + pageSize, currentAlbum.media.length);
      const currentPageHrefs = currentAlbum.media.slice(startIndex, endIndex);

      if (currentPageHrefs.length === 0) {
        setPaginatedMedia([]);
        return;
      }

      setMediaLoading(true);
      setMediaError(null);

      try {
        // Fetch media objects for current page hrefs
        const mediaPromises = currentPageHrefs.map(async (href) => {
          // Extract media ID from href (e.g., "/api/v1/media/123" -> "123")
          const mediaId = href.split('/').pop();
          const response = await fetch(href);
          if (!response.ok) {
            throw new Error(`Failed to fetch media ${mediaId}: ${response.status}`);
          }
          return response.json();
        });

        const mediaObjects = await Promise.all(mediaPromises);
        setPaginatedMedia(mediaObjects);
      } catch (error) {
        console.error('Error fetching paginated media:', error);
        setMediaError(error instanceof Error ? error.message : 'Failed to fetch media');
        setPaginatedMedia([]);
      } finally {
        setMediaLoading(false);
      }
    };

    fetchPaginatedMedia();
  }, [currentAlbum, currentPage, pageSize]);

  // Determine which albums to show
  const albumsToShow: Album[] =
    id && currentAlbum
      ? currentAlbum.children && currentAlbum.children.length > 0
        ? currentAlbum.children
        : [] // Empty array if album has no children
      : albums;

  return (
    <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
      <div className="px-4 py-6 sm:px-0">
        {/* Album Header - Show when viewing a specific album */}
        {id && currentAlbum && (
          <div className="mb-6">
            <div className="flex items-center justify-between">
              <button
                onClick={handleBackToParent}
                className="flex items-center text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors"
              >
                <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 19l-7-7 7-7" />
                </svg>
                Back
              </button>
            </div>

            {/* Album Name and Description - Under back button */}
            <div className="mt-4">
              <h1 className="text-2xl font-bold text-gray-900 dark:text-white">{currentAlbum.name}</h1>

              {/* Editable Description */}
              <div className="mt-1">
                {isEditingDescription ? (
                  <div className="flex items-center space-x-2">
                    <input
                      type="text"
                      value={editedDescription}
                      onChange={(e) => setEditedDescription(e.target.value)}
                      onKeyDown={handleDescriptionKeyDown}
                      className="flex-1 px-3 py-2 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent dark:bg-gray-700 dark:border-gray-600 dark:text-white dark:focus:ring-blue-400"
                      placeholder="Add a description..."
                      autoFocus
                    />
                    <div className="flex space-x-1">
                      <button
                        onClick={handleDescriptionSave}
                        className="px-3 py-2 text-sm border border-blue-600 text-blue-600 rounded-md hover:bg-blue-800 hover:text-white dark:hover:bg-blue-900 dark:hover:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                      >
                        Save
                      </button>
                      <button
                        onClick={handleDescriptionCancel}
                        className="px-3 py-2 text-sm border border-gray-400 text-gray-600 hover:bg-gray-600 hover:text-white dark:text-gray-400 dark:border-gray-500 rounded-md hover:bg-gray-50 dark:hover:bg-gray-600 dark:hover:text-white focus:outline-none focus:ring-2 focus:ring-gray-500"
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                ) : (
                  <div
                    onClick={handleDescriptionEdit}
                    className="cursor-pointer text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200 transition-colors group"
                  >
                    {currentAlbum.description ? (
                      <p className="group-hover:bg-gray-100 dark:group-hover:bg-gray-700 px-2 py-1 rounded transition-colors">
                        {currentAlbum.description}
                      </p>
                    ) : (
                      <p className="text-gray-500 dark:text-gray-500 italic group-hover:bg-gray-100 dark:group-hover:bg-gray-700 px-2 py-1 rounded transition-colors">
                        Click to add description...
                      </p>
                    )}
                    <span className="text-xs text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity ml-2">
                      Click to edit
                    </span>
                  </div>
                )}
              </div>
            </div>

            {/* Album Thumbnail */}
            {currentAlbum.thumbnail && (
              <div className="mt-4">
                <img
                  src={currentAlbum.thumbnail}
                  alt={currentAlbum.name}
                  className="w-32 h-32 object-cover rounded-lg border border-gray-200 dark:border-gray-600"
                />
              </div>
            )}
          </div>
        )}

        <AlbumsList
          albums={albumsToShow}
          loading={loading}
          error={error}
          emptyStateTitle={id ? 'No sub-albums' : 'No albums yet'}
          emptyStateMessage={
            id
              ? "This album doesn't contain any sub-albums."
              : 'Create your first album to get started organizing your photos.'
          }
        />

        {/* Media Gallery - Show only when viewing a specific album */}
        {id && currentAlbum && (
          <MediaGallery 
            media={paginatedMedia} 
            loading={mediaLoading} 
            error={mediaError}
            albumName={currentAlbum.name}
            albumId={id}
            total={currentAlbum.media?.length || 0}
            currentPage={currentPage}
            pageSize={pageSize}
            onPageChange={handlePageChange}
            onMediaDeleted={handleMediaDeleted}
          />
        )}
      </div>

      {/* Create Album Form Modal */}
      <CreateAlbumForm isOpen={isCreateFormOpen} onClose={handleCreateFormClose} onSuccess={handleCreateAlbumSuccess} />
    </div>
  );
};

export default AlbumsPage;
