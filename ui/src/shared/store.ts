import { configureStore } from '@reduxjs/toolkit';
import { useDispatch, useSelector, TypedUseSelectorHook } from 'react-redux';
import rootReducer, { RootState } from '@reducers/index';

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
export const selectMedia = (state: RootState) => state.media;
export const selectSync = (state: RootState) => state.sync;
export const selectStats = (state: RootState) => state.stats;

// Sync job selectors
export const selectSyncJobs = (state: RootState) => state.sync.jobs;
export const selectActiveSyncJobs = (state: RootState) => 
  state.sync.jobs.filter(job => job.status === 'running');
export const selectHasActiveSyncJobs = (state: RootState) => 
  state.sync.jobs.some(job => job.status === 'running');
export const selectSyncLoading = (state: RootState) => state.sync.loading;
export const selectSyncStarting = (state: RootState) => state.sync.startingSync;
export const selectSyncError = (state: RootState) => state.sync.error;

// Specific selectors
export const selectAlbumsPageActive = (state: RootState) => state.albums.isPageActive;
export const selectAlbumsCreateFormOpen = (state: RootState) => state.albums.isCreateFormOpen;
export const selectCurrentAlbum = (state: RootState) => state.albums.currentAlbum;

export const selectAlbumById = (state: RootState, albumId: string) =>
  state.albums.albums.find(album => album.id === albumId);

export const selectMediaById = (state: RootState, mediaId: string) =>
  state.media.media.find(media => media.id === mediaId);

export const selectMediaByAlbum = (state: RootState, albumId: string) =>
  state.media.media.filter(media => media.albumHref.includes(albumId));

export const selectSelectedMedia = (state: RootState) =>
  state.media.media.filter(media => state.media.selectedMediaIds.includes(media.id));

// Upload selectors
export const selectUpload = (state: RootState) => state.upload;
export const selectUploadFiles = (state: RootState) => state.upload.files;
export const selectUploadIsUploading = (state: RootState) => state.upload.isUploading;

export default store;