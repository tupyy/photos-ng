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
import { useThumbnail } from '@shared/contexts';
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
  
  // Thumbnail context
  const { isThumbnailMode, startThumbnailSelection, exitThumbnailMode } = useThumbnail();

  // Local state for infinite scroll media
  const [accumulatedMedia, setAccumulatedMedia] = useState<any[]>([]);
  const [mediaLoading, setMediaLoading] = useState(false);
  const [mediaLoadingMore, setMediaLoadingMore] = useState(false);
  const [mediaError, setMediaError] = useState<string | null>(null);

  // Local state for inline description editing
  const [isEditingDescription, setIsEditingDescription] = useState(false);
  const [editedDescription, setEditedDescription] = useState('');

  // Media infinite scroll state
  const [currentOffset, setCurrentOffset] = useState(0);
  const pageSize = 100; // Number of media items to load per batch
  const [hasMoreMedia, setHasMoreMedia] = useState(true);

  /**
   * Main effect for page initialization and data fetching
   * Runs when component mounts or when album ID changes
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



  const handleMediaDeleted = () => {
    if (id) {
      // Refresh album data to get updated media list and check if thumbnail was affected
      fetchAlbumByIdApi(id);
      // Reset the media accumulation state - the useEffect will reload media when currentAlbum changes
      setCurrentOffset(0);
      setAccumulatedMedia([]);
      setHasMoreMedia(true);
    }
  };

  // Reset media state when album changes
  useEffect(() => {
    setCurrentOffset(0);
    setAccumulatedMedia([]);
    setHasMoreMedia(true);
  }, [id]);

  // Load initial media when album changes
  useEffect(() => {
    if (currentAlbum?.media && currentAlbum.media.length > 0) {
      loadNextMediaBatch(true);
    } else {
      setAccumulatedMedia([]);
    }
  }, [currentAlbum]);

  const loadNextMediaBatch = async (isInitial = false) => {
    if (!currentAlbum?.media || currentAlbum.media.length === 0) {
      return;
    }

    const startIndex = isInitial ? 0 : currentOffset;
    const endIndex = Math.min(startIndex + pageSize, currentAlbum.media.length);
    const batchHrefs = currentAlbum.media.slice(startIndex, endIndex);

    if (batchHrefs.length === 0) {
      setHasMoreMedia(false);
      return;
    }

    // Set appropriate loading state
    if (isInitial) {
      setMediaLoading(true);
    } else {
      setMediaLoadingMore(true);
    }
    setMediaError(null);

    try {
      // Fetch media objects for current batch hrefs
      const mediaPromises = batchHrefs.map(async (href) => {
        // Extract media ID from href (e.g., "/api/v1/media/123" -> "123")
        const mediaId = href.split('/').pop();
        const response = await fetch(href);
        if (!response.ok) {
          throw new Error(`Failed to fetch media ${mediaId}: ${response.status}`);
        }
        return response.json();
      });

      const mediaObjects = await Promise.all(mediaPromises);
      
      if (isInitial) {
        setAccumulatedMedia(mediaObjects);
      } else {
        setAccumulatedMedia(prev => [...prev, ...mediaObjects]);
      }
      
      // Update offset and check if there's more media
      const newOffset = endIndex;
      setCurrentOffset(newOffset);
      setHasMoreMedia(newOffset < currentAlbum.media.length);
      
    } catch (error) {
      console.error('Error fetching media batch:', error);
      setMediaError(error instanceof Error ? error.message : 'Failed to fetch media');
      if (isInitial) {
        setAccumulatedMedia([]);
      }
    } finally {
      if (isInitial) {
        setMediaLoading(false);
      } else {
        setMediaLoadingMore(false);
      }
    }
  };

  const handleLoadMore = () => {
    loadNextMediaBatch(false);
  };

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
        {/* Thumbnail Selection Mode Banner */}
        {isThumbnailMode && (
          <div className="mb-6 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center">
                <svg className="w-5 h-5 text-blue-600 dark:text-blue-400 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
                <div>
                  <h3 className="text-sm font-medium text-blue-800 dark:text-blue-200">
                    Thumbnail Selection Mode
                  </h3>
                  <p className="text-xs text-blue-600 dark:text-blue-300 mt-1">
                    Navigate through folders and click on any photo to set it as the album thumbnail
                  </p>
                </div>
              </div>
              <button
                onClick={exitThumbnailMode}
                className="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-200 text-sm font-medium"
              >
                Exit
              </button>
            </div>
          </div>
        )}

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
          onSetThumbnail={startThumbnailSelection}
        />

        {/* Media Gallery - Show only when viewing a specific album */}
        {id && currentAlbum && (
          <MediaGallery
            media={accumulatedMedia}
            loading={mediaLoading}
            loadingMore={mediaLoadingMore}
            error={mediaError}
            albumName={currentAlbum.name}
            albumId={id}
            total={currentAlbum.media?.length || 0}
            hasMore={hasMoreMedia}
            onLoadMore={handleLoadMore}
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
