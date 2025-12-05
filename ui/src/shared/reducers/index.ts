import { combineReducers } from '@reduxjs/toolkit';
import albumsReducer from './albumsSlice';
import timelineReducer from './timelineSlice';
import albumsMediaReducer from './albumsMediaSlice';
import syncReducer from './syncSlice';
import uploadReducer from './uploadSlice';
import statsReducer from './statsSlice';
import mediaReducer from './mediaSlice';
import userReducer from './userSlice';

// Combine all reducers
const rootReducer = combineReducers({
  albums: albumsReducer,
  timeline: timelineReducer,
  albumsMedia: albumsMediaReducer,
  sync: syncReducer,
  upload: uploadReducer,
  stats: statsReducer,
  media: mediaReducer,
  user: userReducer,
});

export type RootState = ReturnType<typeof rootReducer>;
export default rootReducer;

// Export sync actions for convenience
export { startSyncJob, fetchSyncJobs, fetchSyncJob, clearError, updateJob } from './syncSlice';

// Export media actions for convenience
export { deleteMedia } from './mediaSlice';

// Export user actions for convenience
export { fetchCurrentUser, clearUser } from './userSlice';