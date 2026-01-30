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
// Note: React Compiler handles memoization automatically, no need for useCallback
export const useAlbumsApi = () => {
  const dispatch = useAppDispatch();
  const albumsState = useAppSelector((state) => state.albums);

  return {
    // State
    ...albumsState,

    // Actions - React Compiler auto-memoizes these
    fetchAlbums: (params?: { limit?: number; offset?: number }) => dispatch(fetchAlbums(params || {})),
    fetchAlbumById: (id: string) => dispatch(fetchAlbumById(id)),
    createAlbum: (albumData: CreateAlbumRequest) => dispatch(createAlbum(albumData)),
    updateAlbum: (id: string, albumData: UpdateAlbumRequest) => dispatch(updateAlbum({ id, albumData })),
    deleteAlbum: (id: string) => dispatch(deleteAlbum(id)),
    clearCurrentAlbum: () => dispatch(clearCurrentAlbum()),
    clearError: () => dispatch(clearAlbumsError()),
    setFilters: (filters: { limit?: number; offset?: number }) => dispatch(setAlbumsFilters(filters)),
    toggleAlbumSelection: (albumId: string) => dispatch(toggleAlbumSelection(albumId)),
    selectAllAlbums: () => dispatch(selectAllAlbums()),
    clearAlbumSelection: () => dispatch(clearAlbumSelection()),
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
    fetchMedia: (params?: Partial<TimelineFilters> & { forceRefresh?: boolean }) => dispatch(fetchTimelineMedia(params || {})),
    clearError: () => dispatch(clearTimelineError()),
    setFilters: (filters: Partial<TimelineFilters>) => dispatch(setTimelineFilters(filters)),
    clearFilters: () => dispatch(clearTimelineFilters()),
    toggleMediaSelection: (mediaId: string) => dispatch(toggleTimelineMediaSelection(mediaId)),
    selectAllMedia: () => dispatch(selectAllTimelineMedia()),
    clearSelection: () => dispatch(clearTimelineSelection()),
    invalidateCache: () => dispatch(invalidateTimelineCache()),
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
    fetchMedia: (params: Partial<AlbumsMediaFilters> & { forceRefresh?: boolean }) => {
      if (!params.albumId) {
        throw new Error('albumId is required for albums media');
      }
      return dispatch(fetchAlbumsMedia(params as AlbumsMediaFilters & { forceRefresh?: boolean }));
    },
    setCurrentAlbum: (albumId: string | null) => dispatch(setCurrentAlbum(albumId)),
    clearError: () => dispatch(clearAlbumsMediaError()),
    setFilters: (filters: Partial<AlbumsMediaFilters>) => dispatch(setAlbumsMediaFilters(filters)),
    clearFilters: () => dispatch(clearAlbumsMediaFilters()),
    toggleMediaSelection: (mediaId: string) => dispatch(toggleAlbumsMediaSelection(mediaId)),
    selectAllMedia: () => dispatch(selectAllAlbumsMedia()),
    clearSelection: () => dispatch(clearAlbumsMediaSelection()),
    invalidateCache: () => dispatch(invalidateAlbumsMediaCache()),
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
    fetchMedia: (params?: Partial<MediaFilters> & { forceRefresh?: boolean }) => dispatch(fetchMedia(params || {})),
    fetchMediaById: (params: { id: string; forceRefresh?: boolean }) => dispatch(fetchMediaById(params)),
    updateMedia: (id: string, mediaData: UpdateMediaRequest) => dispatch(updateMedia({ id, mediaData })),
    deleteMedia: (id: string) => dispatch(deleteMedia(id)),
    clearCurrentMedia: () => dispatch(clearCurrentMedia()),
    clearError: () => dispatch(clearMediaError()),
    setFilters: (filters: Partial<MediaFilters>) => dispatch(setMediaFilters(filters)),
    clearFilters: () => dispatch(clearMediaFilters()),
    toggleMediaSelection: (mediaId: string) => dispatch(toggleMediaSelection(mediaId)),
    selectAllMedia: () => dispatch(selectAllMedia()),
    clearSelection: () => dispatch(clearMediaSelection()),
    setViewMode: (viewMode: 'grid' | 'list') => dispatch(setViewMode(viewMode)),
    invalidateCache: () => dispatch(invalidateMediaCache()),
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
    fetchStats: () => dispatch(fetchStats()),
    clearError: () => dispatch(clearStatsError()),
    resetStats: () => dispatch(resetStats()),
  };
};
