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
export const selectTimeline = (state: RootState) => state.timeline;
export const selectSync = (state: RootState) => state.sync;

// Specific selectors
export const selectAlbumById = (state: RootState, albumId: string) =>
  state.albums.albums.find(album => album.id === albumId);

export const selectMediaById = (state: RootState, mediaId: string) =>
  state.media.media.find(media => media.id === mediaId);

export const selectMediaByAlbum = (state: RootState, albumId: string) =>
  state.media.media.filter(media => media.albumHref.includes(albumId));

export const selectSelectedMedia = (state: RootState) =>
  state.media.media.filter(media => state.media.selectedMediaIds.includes(media.id));

export const selectTimelineBucketsByYear = (state: RootState, year: number) =>
  state.timeline.buckets.filter(bucket => bucket.year === year);

export const selectTimelineBucketsByYearAndMonth = (state: RootState, year: number, month: number) =>
  state.timeline.buckets.filter(bucket => bucket.year === year && bucket.month === month);

export default store;