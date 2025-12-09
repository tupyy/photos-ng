import { configureStore } from '@reduxjs/toolkit';
import { useDispatch, useSelector, TypedUseSelectorHook } from 'react-redux';
import rootReducer, { RootState } from '@reducers/index';
import { AUTHZ_ENABLED } from '@shared/config';

// Configure store
export const store = configureStore({
  reducer: rootReducer,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        // Ignore these action types
        ignoredActions: ['persist/PERSIST', 'persist/REHYDRATE'],
      },
    }),
  devTools: process.env.NODE_ENV !== 'production',
});

export type AppDispatch = typeof store.dispatch;

// Export typed hooks
export const useAppDispatch = () => useDispatch<AppDispatch>();
export const useAppSelector: TypedUseSelectorHook<RootState> = useSelector;

// Export store selectors
export const selectAlbums = (state: RootState) => state.albums;
export const selectTimeline = (state: RootState) => state.timeline;
export const selectAlbumsMedia = (state: RootState) => state.albumsMedia;
export const selectStats = (state: RootState) => state.stats;

// Specific selectors
export const selectAlbumsPageActive = (state: RootState) => state.albums.isPageActive;
export const selectAlbumsCreateFormOpen = (state: RootState) => state.albums.isCreateFormOpen;
export const selectCurrentAlbum = (state: RootState) => state.albums.currentAlbum;
export const selectSelectedAlbumIds = (state: RootState) => state.albums.selectedAlbumIds;

export const selectAlbumById = (state: RootState, albumId: string) =>
  state.albums.albums.find(album => album.id === albumId);

// Timeline selectors
export const selectTimelineMediaById = (state: RootState, mediaId: string) =>
  state.timeline.media.find(media => media.id === mediaId);

export const selectTimelineSelectedMedia = (state: RootState) =>
  state.timeline.media.filter(media => state.timeline.selectedMediaIds.includes(media.id));

// Albums media selectors
export const selectAlbumsMediaById = (state: RootState, mediaId: string) =>
  state.albumsMedia.media.find(media => media.id === mediaId);

export const selectAlbumsMediaByAlbum = (state: RootState, albumId: string) =>
  state.albumsMedia.currentAlbumId === albumId ? state.albumsMedia.media : [];

export const selectAlbumsSelectedMedia = (state: RootState) =>
  state.albumsMedia.media.filter(media => state.albumsMedia.selectedMediaIds.includes(media.id));

// Upload selectors
export const selectUpload = (state: RootState) => state.upload;
export const selectUploadFiles = (state: RootState) => state.upload.files;
export const selectUploadIsUploading = (state: RootState) => state.upload.isUploading;

// User selectors
export const selectUser = (state: RootState) => state.user.user;
export const selectUserLoading = (state: RootState) => state.user.loading;
export const selectUserInitialized = (state: RootState) => state.user.initialized;
export const selectUserError = (state: RootState) => state.user.error;

// Permission selectors - when AUTHZ_ENABLED is false, all permissions are granted
export const selectCanCreateAlbums = (state: RootState) => {
  if (!AUTHZ_ENABLED) return true;
  return state.user.user?.permissions?.can_create_albums === 'allowed';
};

export default store;