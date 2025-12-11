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
 * - Unified selection mode for albums and media
 *
 * The component supports both root-level album listing and individual album views
 * with media galleries, depending on the URL parameter.
 */

import React, { useEffect, useState, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAppDispatch, useAppSelector, selectAlbumsCreateFormOpen, selectCurrentAlbum } from '@shared/store';
import { setPageActive, setCreateFormOpen, setCurrentAlbum } from '@shared/reducers/albumsSlice';
import { useAlbumsApi, useAlbumsMediaApi, useMediaApi } from '@shared/hooks/useApi';
import { useThumbnail } from '@shared/contexts';
import { PillButton, SelectionBar, ConfirmDeleteModal } from '@shared/components';
import { ListMediaSortByEnum, ListMediaSortOrderEnum, ListMediaDirectionEnum } from '@generated/api/media-api';
import {
  ArrowLeftIcon,
  PencilSquareIcon,
  CheckIcon,
  XMarkIcon,
  PhotoIcon,
  Squares2X2Icon,
} from '@heroicons/react/24/outline';
import AlbumsList from './components/AlbumsList';
import SubAlbumsList from './components/SubAlbumsList';
import CreateAlbumForm from './components/CreateAlbumForm';
import MediaGallery from './components/MediaGallery';

const AlbumsPage: React.FC = () => {
  // URL parameters and navigation
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();

  // Redux state selectors
  const isCreateFormOpen = useAppSelector(selectAlbumsCreateFormOpen);
  const currentAlbum = useAppSelector(selectCurrentAlbum);

  // API hooks for data fetching and operations
  const {
    albums,
    loading,
    error,
    fetchAlbums,
    fetchAlbumById: fetchAlbumByIdApi,
    updateAlbum,
    deleteAlbum,
  } = useAlbumsApi();
  const {
    media,
    loading: mediaLoading,
    loadingMore: mediaLoadingMore,
    error: mediaError,
    hasMore,
    nextCursor,
    fetchMedia,
    setCurrentAlbum: setCurrentAlbumMedia,
  } = useAlbumsMediaApi();
  const { deleteMedia } = useMediaApi();

  // Thumbnail context
  const { isThumbnailMode, startThumbnailSelection, exitThumbnailMode } = useThumbnail();

  // Page size for media pagination
  const pageSize = 100;

  // Local state for inline description editing
  const [isEditingDescription, setIsEditingDescription] = useState(false);
  const [editedDescription, setEditedDescription] = useState('');
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // Local state for managing media display
  const [hasMoreMedia, setHasMoreMedia] = useState(true);

  // Track previous album ID to detect actual changes
  const [prevAlbumId, setPrevAlbumId] = useState<string | undefined>(undefined);

  // Unified selection state
  const [isSelectionMode, setIsSelectionMode] = useState(false);
  const [selectedAlbumIds, setSelectedAlbumIds] = useState<Set<string>>(new Set());
  const [selectedMediaIds, setSelectedMediaIds] = useState<Set<string>>(new Set());
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [isSettingThumbnail, setIsSettingThumbnail] = useState(false);

  // Effects
  useEffect(() => {
    dispatch(setPageActive(true));

    if (prevAlbumId !== id) {
      setCurrentAlbumMedia(id || null);
      setPrevAlbumId(id);
      // Reset selection when changing albums
      exitSelectionMode();
    }

    if (id) {
      fetchAlbumByIdApi(id);
    } else {
      fetchAlbums({ limit: 1000, offset: 0 });
      dispatch(setCurrentAlbum(null));
    }

    return () => {
      dispatch(setPageActive(false));
    };
  }, [dispatch, id, fetchAlbums, fetchAlbumByIdApi, prevAlbumId]);

  useEffect(() => {
    if (currentAlbum) {
      setEditedDescription(currentAlbum.description || '');
    }
  }, [currentAlbum]);

  useEffect(() => {
    if (isEditingDescription && textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      textareaRef.current.style.height = textareaRef.current.scrollHeight + 'px';
    }
  }, [editedDescription, isEditingDescription]);

  useEffect(() => {
    if (id && currentAlbum) {
      fetchMedia({
        limit: pageSize,
        albumId: id,
        sortBy: ListMediaSortByEnum.CapturedAt,
        sortOrder: ListMediaSortOrderEnum.Desc,
        direction: ListMediaDirectionEnum.Forward,
        forceRefresh: true,
      });
    }
  }, [id, currentAlbum]);

  useEffect(() => {
    setHasMoreMedia(hasMore);
  }, [hasMore]);

  // Navigation handlers
  const handleCreateFormClose = () => dispatch(setCreateFormOpen(false));

  const handleCreateAlbumSuccess = (albumId: string) => {
    navigate(`/albums/${albumId}`);
  };

  const handleBackToParent = () => {
    if (currentAlbum?.parentHref) {
      const parentId = currentAlbum.parentHref.split('/').pop();
      navigate(`/albums/${parentId}`);
    } else {
      navigate('/albums');
    }
  };

  // Description editing handlers
  const handleDescriptionEdit = () => {
    setIsEditingDescription(true);
  };

  const handleDescriptionSave = async () => {
    if (!currentAlbum || !id) return;

    try {
      await updateAlbum(id, { description: editedDescription });
      setIsEditingDescription(false);
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
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleDescriptionSave();
    } else if (e.key === 'Escape') {
      handleDescriptionCancel();
    }
  };

  // Media handlers
  const handleMediaDeleted = () => {
    if (id) {
      fetchAlbumByIdApi(id);
      fetchMedia({
        limit: pageSize,
        albumId: id,
        sortBy: ListMediaSortByEnum.CapturedAt,
        sortOrder: ListMediaSortOrderEnum.Desc,
        direction: ListMediaDirectionEnum.Forward,
        forceRefresh: true,
      });
    }
  };

  const handleLoadMore = () => {
    if (!mediaLoadingMore && hasMore && id) {
      fetchMedia({
        limit: pageSize,
        cursor: nextCursor,
        albumId: id,
        sortBy: ListMediaSortByEnum.CapturedAt,
        sortOrder: ListMediaSortOrderEnum.Desc,
        direction: ListMediaDirectionEnum.Forward,
      });
    }
  };

  // Selection handlers
  const enterSelectionMode = () => setIsSelectionMode(true);

  const exitSelectionMode = () => {
    setIsSelectionMode(false);
    setSelectedAlbumIds(new Set());
    setSelectedMediaIds(new Set());
  };

  const toggleAlbumSelection = (albumId: string) => {
    const newSet = new Set(selectedAlbumIds);
    if (newSet.has(albumId)) {
      newSet.delete(albumId);
    } else {
      newSet.add(albumId);
    }
    setSelectedAlbumIds(newSet);

    // Auto-enter selection mode if not active
    if (!isSelectionMode) setIsSelectionMode(true);
  };

  const toggleMediaSelection = (mediaId: string) => {
    const newSet = new Set(selectedMediaIds);
    if (newSet.has(mediaId)) {
      newSet.delete(mediaId);
    } else {
      newSet.add(mediaId);
    }
    setSelectedMediaIds(newSet);

    // Auto-enter selection mode if not active
    if (!isSelectionMode) setIsSelectionMode(true);
  };

  const handleBulkDelete = () => {
    const totalSelected = selectedAlbumIds.size + selectedMediaIds.size;
    if (totalSelected > 0) {
      setShowDeleteModal(true);
    }
  };

  const confirmBulkDelete = async () => {
    setIsDeleting(true);
    setShowDeleteModal(false);

    try {
      // Delete albums
      const albumDeletePromises = Array.from(selectedAlbumIds).map((albumId) => deleteAlbum(albumId));

      // Delete media
      const mediaDeletePromises = Array.from(selectedMediaIds).map((mediaId) => deleteMedia(mediaId));

      await Promise.all([...albumDeletePromises, ...mediaDeletePromises]);

      // Refresh data
      if (id) {
        fetchAlbumByIdApi(id);
        handleMediaDeleted();
      } else {
        fetchAlbums({ limit: 1000, offset: 0 });
      }

      exitSelectionMode();
    } catch (error) {
      console.error('Failed to delete items:', error);
    } finally {
      setIsDeleting(false);
    }
  };

  const handleSetThumbnail = async () => {
    if (selectedMediaIds.size !== 1 || !id) return;

    const selectedMediaId = Array.from(selectedMediaIds)[0];
    setIsSettingThumbnail(true);

    try {
      await updateAlbum(id, { thumbnail: selectedMediaId });
      fetchAlbumByIdApi(id);
      exitSelectionMode();
    } catch (error) {
      console.error('Failed to set album thumbnail:', error);
    } finally {
      setIsSettingThumbnail(false);
    }
  };

  const totalSelectedCount = selectedAlbumIds.size + selectedMediaIds.size;
  const canSetThumbnail = id && selectedMediaIds.size === 1 && selectedAlbumIds.size === 0;

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-slate-900 pb-32">
      <div className="max-w-[1400px] mx-auto px-6 sm:px-8 lg:px-12 pt-10">
        {/* Thumbnail Selection Mode Banner */}
        {isThumbnailMode && (
          <div className="mb-8 rounded-xl bg-indigo-50 dark:bg-indigo-900/20 border border-indigo-100 dark:border-indigo-800 p-4 flex items-center justify-between shadow-sm animate-fadeIn">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-indigo-100 dark:bg-indigo-800 rounded-lg text-indigo-600 dark:text-indigo-300">
                <PhotoIcon className="w-5 h-5" />
              </div>
              <div>
                <h3 className="font-semibold text-indigo-900 dark:text-indigo-200 text-sm">Select Album Cover</h3>
                <p className="text-xs text-indigo-600 dark:text-indigo-400">
                  Click any photo below to set it as the thumbnail.
                </p>
              </div>
            </div>
            <button
              onClick={exitThumbnailMode}
              className="text-sm font-medium text-indigo-600 hover:text-indigo-800 dark:text-indigo-400 dark:hover:text-indigo-200 px-3 py-1.5 rounded-md hover:bg-indigo-100 dark:hover:bg-indigo-800/50 transition-colors"
            >
              Exit
            </button>
          </div>
        )}

        {/* Album Header - Show when viewing a specific album */}
        {id && currentAlbum && (
          <header className="mb-12 animate-fadeIn">
            {/* Top Controls Row */}
            <div className="flex items-center justify-between mb-8">
              <PillButton onClick={handleBackToParent}>
                <ArrowLeftIcon className="w-4 h-4" />
                Back
              </PillButton>

              {/* Selection Mode Toggle */}
              {!isSelectionMode ? (
                <PillButton onClick={enterSelectionMode}>
                  <Squares2X2Icon className="w-4 h-4" />
                  Select
                </PillButton>
              ) : (
                <PillButton onClick={exitSelectionMode}>Done</PillButton>
              )}
            </div>

            {/* Title & Description */}
            <div className="space-y-4 max-w-4xl">
              <h1 className="text-4xl sm:text-5xl font-extrabold tracking-tight text-gray-900 dark:text-white leading-tight">
                {currentAlbum.name}
              </h1>

              {/* Editable Description */}
              <div className="relative group">
                {isEditingDescription ? (
                  <div className="relative">
                    <textarea
                      ref={textareaRef}
                      value={editedDescription}
                      onChange={(e) => setEditedDescription(e.target.value)}
                      onKeyDown={handleDescriptionKeyDown}
                      className="w-full bg-transparent text-lg sm:text-xl text-gray-500 dark:text-gray-400 resize-none outline-none border-b-2 border-blue-500 py-2 placeholder-gray-300 dark:placeholder-gray-600"
                      placeholder="Add a description..."
                      autoFocus
                    />
                    <div className="absolute right-0 top-2 flex gap-2">
                      <button
                        onMouseDown={(e) => e.preventDefault()}
                        onClick={handleDescriptionSave}
                        className="p-1.5 bg-blue-100 dark:bg-blue-900 text-blue-600 dark:text-blue-400 rounded hover:bg-blue-200 dark:hover:bg-blue-800 transition-colors"
                      >
                        <CheckIcon className="w-4 h-4" />
                      </button>
                      <button
                        onMouseDown={(e) => e.preventDefault()}
                        onClick={handleDescriptionCancel}
                        className="p-1.5 bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 rounded hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
                      >
                        <XMarkIcon className="w-4 h-4" />
                      </button>
                    </div>
                  </div>
                ) : (
                  <div
                    onClick={handleDescriptionEdit}
                    className="flex items-baseline gap-3 cursor-pointer py-2 -ml-2 px-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800/50 transition-colors"
                  >
                    <p className="text-lg sm:text-xl text-gray-500 dark:text-gray-400 font-normal leading-relaxed">
                      {currentAlbum.description || (
                        <span className="italic opacity-50">No description... click to add one.</span>
                      )}
                    </p>
                    <PencilSquareIcon className="w-5 h-5 text-gray-300 group-hover:text-gray-500 dark:group-hover:text-gray-400 opacity-0 group-hover:opacity-100 transition-all flex-shrink-0" />
                  </div>
                )}
              </div>
            </div>
          </header>
        )}

        {/* Sub-albums Section - Show when viewing a specific album with children */}
        {id && currentAlbum && currentAlbum.children && currentAlbum.children.length > 0 && (
          <section className="mb-12">
            <SubAlbumsList
              albums={currentAlbum.children}
              loading={loading}
              isSelectionMode={isSelectionMode}
              selectedIds={selectedAlbumIds}
              onToggleSelection={toggleAlbumSelection}
            />
          </section>
        )}

        {/* Root-level Albums List - Show when at /albums (no id) */}
        {!id && (
          <section>
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white">Albums</h2>
              {!isSelectionMode ? (
                <PillButton onClick={enterSelectionMode}>
                  <Squares2X2Icon className="w-4 h-4" />
                  Select
                </PillButton>
              ) : (
                <PillButton onClick={exitSelectionMode}>Done</PillButton>
              )}
            </div>
            <AlbumsList
              albums={albums}
              loading={loading}
              error={error}
              emptyStateTitle="No albums yet"
              emptyStateMessage="Create your first album to get started organizing your photos."
              onSetThumbnail={startThumbnailSelection}
              isSelectionMode={isSelectionMode}
              selectedIds={selectedAlbumIds}
              onToggleSelection={toggleAlbumSelection}
            />
          </section>
        )}

        {/* Media Gallery - Show only when viewing a specific album */}
        {id && currentAlbum && (
          <section>
            <h2 className="text-xl font-bold text-gray-900 dark:text-white mb-6">Photos</h2>
            <MediaGallery
              media={media || []}
              loading={mediaLoading}
              loadingMore={mediaLoadingMore}
              error={mediaError}
              albumName={currentAlbum.name}
              albumId={id}
              total={currentAlbum.media?.length || 0}
              hasMore={hasMoreMedia}
              groupByWeek={false}
              viewMode="masonry"
              onLoadMore={handleLoadMore}
              onMediaDeleted={handleMediaDeleted}
              isSelectionMode={isSelectionMode}
              selectedIds={selectedMediaIds}
              onToggleSelection={toggleMediaSelection}
            />
          </section>
        )}
      </div>

      {/* Floating Selection Bar */}
      <SelectionBar
        selectedCount={totalSelectedCount}
        isVisible={isSelectionMode}
        onClose={exitSelectionMode}
        onDelete={handleBulkDelete}
        onSetThumbnail={canSetThumbnail ? handleSetThumbnail : undefined}
        isDeleting={isDeleting}
        isSettingThumbnail={isSettingThumbnail}
      />

      {/* Delete Confirmation Modal */}
      <ConfirmDeleteModal
        isOpen={showDeleteModal}
        itemCount={totalSelectedCount}
        itemType={
          selectedAlbumIds.size > 0 && selectedMediaIds.size > 0
            ? 'Item'
            : selectedAlbumIds.size > 0
              ? 'Album'
              : 'Photo'
        }
        isDeleting={isDeleting}
        onConfirm={confirmBulkDelete}
        onCancel={() => setShowDeleteModal(false)}
      />

      {/* Create Album Form Modal */}
      <CreateAlbumForm isOpen={isCreateFormOpen} onClose={handleCreateFormClose} onSuccess={handleCreateAlbumSuccess} />
    </div>
  );
};

export default AlbumsPage;
