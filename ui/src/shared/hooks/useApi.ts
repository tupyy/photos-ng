import { useCallback } from 'react';
import { useAppDispatch, useAppSelector } from '@shared/store';
import {
  fetchAlbums,
  fetchAlbumById,
  createAlbum,
  updateAlbum,
  deleteAlbum,
  clearCurrentAlbum,
  clearError as clearAlbumsError,
  setFilters as setAlbumsFilters,
  toggleAlbumSelection,
  selectAllAlbums,
  clearAlbumSelection,
} from '@reducers/albumsSlice';
import {
  fetchTimelineMedia,
  clearError as clearTimelineError,
  setFilters as setTimelineFilters,
  clearFilters as clearTimelineFilters,
  toggleMediaSelection as toggleTimelineMediaSelection,
  selectAllMedia as selectAllTimelineMedia,
  clearSelection as clearTimelineSelection,
  invalidateCache as invalidateTimelineCache,
  TimelineFilters,
} from '@reducers/timelineSlice';
import {
  fetchAlbumsMedia,
  clearError as clearAlbumsMediaError,
  setCurrentAlbum,
  setFilters as setAlbumsMediaFilters,
  clearFilters as clearAlbumsMediaFilters,
  toggleMediaSelection as toggleAlbumsMediaSelection,
  selectAllMedia as selectAllAlbumsMedia,
  clearSelection as clearAlbumsMediaSelection,
  invalidateCache as invalidateAlbumsMediaCache,
  AlbumsMediaFilters,
} from '@reducers/albumsMediaSlice';
import {
  fetchStats,
  clearError as clearStatsError,
  resetStats,
} from '@reducers/statsSlice';
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
  clearSelection as clearMediaSelection,
  setViewMode,
  invalidateCache as invalidateMediaCache,
  MediaFilters,
} from '@reducers/mediaSlice';
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

// Custom hook for Timeline Media API
export const useTimelineApi = () => {
  const dispatch = useAppDispatch();
  const timelineState = useAppSelector((state) => state.timeline);

  return {
    // State
    ...timelineState,
    
    // Actions
    fetchMedia: useCallback(
      (params?: Partial<TimelineFilters> & { forceRefresh?: boolean }) => dispatch(fetchTimelineMedia(params || {})),
      [dispatch]
    ),
    clearError: useCallback(() => dispatch(clearTimelineError()), [dispatch]),
    setFilters: useCallback(
      (filters: Partial<TimelineFilters>) => dispatch(setTimelineFilters(filters)),
      [dispatch]
    ),
    clearFilters: useCallback(() => dispatch(clearTimelineFilters()), [dispatch]),
    toggleMediaSelection: useCallback(
      (mediaId: string) => dispatch(toggleTimelineMediaSelection(mediaId)),
      [dispatch]
    ),
    selectAllMedia: useCallback(() => dispatch(selectAllTimelineMedia()), [dispatch]),
    clearSelection: useCallback(() => dispatch(clearTimelineSelection()), [dispatch]),
    invalidateCache: useCallback(() => dispatch(invalidateTimelineCache()), [dispatch]),
  };
};

// Custom hook for Albums Media API
export const useAlbumsMediaApi = () => {
  const dispatch = useAppDispatch();
  const albumsMediaState = useAppSelector((state) => state.albumsMedia);

  return {
    // State
    ...albumsMediaState,
    
    // Actions
    fetchMedia: useCallback(
      (params: Partial<AlbumsMediaFilters> & { forceRefresh?: boolean }) => {
        if (!params.albumId) {
          throw new Error('albumId is required for albums media');
        }
        return dispatch(fetchAlbumsMedia(params as AlbumsMediaFilters & { forceRefresh?: boolean }));
      },
      [dispatch]
    ),
    setCurrentAlbum: useCallback(
      (albumId: string | null) => dispatch(setCurrentAlbum(albumId)),
      [dispatch]
    ),
    clearError: useCallback(() => dispatch(clearAlbumsMediaError()), [dispatch]),
    setFilters: useCallback(
      (filters: Partial<AlbumsMediaFilters>) => dispatch(setAlbumsMediaFilters(filters)),
      [dispatch]
    ),
    clearFilters: useCallback(() => dispatch(clearAlbumsMediaFilters()), [dispatch]),
    toggleMediaSelection: useCallback(
      (mediaId: string) => dispatch(toggleAlbumsMediaSelection(mediaId)),
      [dispatch]
    ),
    selectAllMedia: useCallback(() => dispatch(selectAllAlbumsMedia()), [dispatch]),
    clearSelection: useCallback(() => dispatch(clearAlbumsMediaSelection()), [dispatch]),
    invalidateCache: useCallback(() => dispatch(invalidateAlbumsMediaCache()), [dispatch]),
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
      (params?: Partial<MediaFilters> & { forceRefresh?: boolean }) => dispatch(fetchMedia(params || {})),
      [dispatch]
    ),
    fetchMediaById: useCallback(
      (params: { id: string; forceRefresh?: boolean }) => dispatch(fetchMediaById(params)),
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
    clearSelection: useCallback(() => dispatch(clearMediaSelection()), [dispatch]),
    setViewMode: useCallback(
      (viewMode: 'grid' | 'list') => dispatch(setViewMode(viewMode)),
      [dispatch]
    ),
    invalidateCache: useCallback(() => dispatch(invalidateMediaCache()), [dispatch]),
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