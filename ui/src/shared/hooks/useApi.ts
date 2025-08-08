import { useCallback } from 'react';
import { useAppDispatch, useAppSelector } from '@shared/store';
import {
  fetchAlbums,
  fetchAlbumById,
  createAlbum,
  updateAlbum,
  deleteAlbum,
  syncAlbum,
  clearCurrentAlbum,
  clearError as clearAlbumsError,
  setFilters as setAlbumsFilters,
  toggleAlbumSelection,
  selectAllAlbums,
  clearAlbumSelection,
} from '@reducers/albumsSlice';
import {
  fetchMedia,
  fetchMediaById,
  updateMedia,
  deleteMedia,
  clearCurrentMedia,
  clearError as clearMediaError,
  setFilters as setMediaFilters,
  clearFilters as clearMediaFilters,
  toggleMediaSelection,
  selectAllMedia,
  clearSelection,
  setViewMode,
  MediaFilters,
} from '@reducers/mediaSlice';
import {
  fetchStats,
  clearError as clearStatsError,
  resetStats,
} from '@reducers/statsSlice';
import { CreateAlbumRequest, UpdateAlbumRequest, UpdateMediaRequest } from '@generated/models';

// Custom hook for Albums API
export const useAlbumsApi = () => {
  const dispatch = useAppDispatch();
  const albumsState = useAppSelector((state) => state.albums);

  return {
    // State
    ...albumsState,
    
    // Actions
    fetchAlbums: useCallback(
      (params?: { limit?: number; offset?: number }) => dispatch(fetchAlbums(params || {})),
      [dispatch]
    ),
    fetchAlbumById: useCallback(
      (id: string) => dispatch(fetchAlbumById(id)),
      [dispatch]
    ),
    createAlbum: useCallback(
      (albumData: CreateAlbumRequest) => dispatch(createAlbum(albumData)),
      [dispatch]
    ),
    updateAlbum: useCallback(
      (id: string, albumData: UpdateAlbumRequest) => dispatch(updateAlbum({ id, albumData })),
      [dispatch]
    ),
    deleteAlbum: useCallback(
      (id: string) => dispatch(deleteAlbum(id)),
      [dispatch]
    ),
    syncAlbum: useCallback(
      (id: string) => dispatch(syncAlbum(id)),
      [dispatch]
    ),
    clearCurrentAlbum: useCallback(() => dispatch(clearCurrentAlbum()), [dispatch]),
    clearError: useCallback(() => dispatch(clearAlbumsError()), [dispatch]),
    setFilters: useCallback(
      (filters: { limit?: number; offset?: number }) => dispatch(setAlbumsFilters(filters)),
      [dispatch]
    ),
    toggleAlbumSelection: useCallback(
      (albumId: string) => dispatch(toggleAlbumSelection(albumId)),
      [dispatch]
    ),
    selectAllAlbums: useCallback(() => dispatch(selectAllAlbums()), [dispatch]),
    clearAlbumSelection: useCallback(() => dispatch(clearAlbumSelection()), [dispatch]),
  };
};

// Custom hook for Media API
export const useMediaApi = () => {
  const dispatch = useAppDispatch();
  const mediaState = useAppSelector((state) => state.media);

  return {
    // State
    ...mediaState,
    
    // Actions
    fetchMedia: useCallback(
      (params?: Partial<MediaFilters>) => dispatch(fetchMedia(params || {})),
      [dispatch]
    ),
    fetchMediaById: useCallback(
      (id: string) => dispatch(fetchMediaById(id)),
      [dispatch]
    ),
    updateMedia: useCallback(
      (id: string, mediaData: UpdateMediaRequest) => dispatch(updateMedia({ id, mediaData })),
      [dispatch]
    ),
    deleteMedia: useCallback(
      (id: string) => dispatch(deleteMedia(id)),
      [dispatch]
    ),
    clearCurrentMedia: useCallback(() => dispatch(clearCurrentMedia()), [dispatch]),
    clearError: useCallback(() => dispatch(clearMediaError()), [dispatch]),
    setFilters: useCallback(
      (filters: Partial<MediaFilters>) => dispatch(setMediaFilters(filters)),
      [dispatch]
    ),
    clearFilters: useCallback(() => dispatch(clearMediaFilters()), [dispatch]),
    toggleMediaSelection: useCallback(
      (mediaId: string) => dispatch(toggleMediaSelection(mediaId)),
      [dispatch]
    ),
    selectAllMedia: useCallback(() => dispatch(selectAllMedia()), [dispatch]),
    clearSelection: useCallback(() => dispatch(clearSelection()), [dispatch]),
    setViewMode: useCallback(
      (mode: 'grid' | 'list') => dispatch(setViewMode(mode)),
      [dispatch]
    ),
  };
};

// Custom hook for Stats API
export const useStatsApi = () => {
  const dispatch = useAppDispatch();
  const statsState = useAppSelector((state) => state.stats);

  return {
    // State
    ...statsState,
    
    // Actions
    fetchStats: useCallback(() => dispatch(fetchStats()), [dispatch]),
    clearError: useCallback(() => dispatch(clearStatsError()), [dispatch]),
    resetStats: useCallback(() => dispatch(resetStats()), [dispatch]),
  };
};