import { combineReducers } from '@reduxjs/toolkit';
import albumsReducer from './albumsSlice';
import mediaReducer from './mediaSlice';
import syncReducer from './syncSlice';
import uploadReducer from './uploadSlice';
import statsReducer from './statsSlice';

// Combine all reducers
const rootReducer = combineReducers({
  albums: albumsReducer,
  media: mediaReducer,
  sync: syncReducer,
  upload: uploadReducer,
  stats: statsReducer,
});

export type RootState = ReturnType<typeof rootReducer>;
export default rootReducer;

// Export sync actions for convenience
export { startSyncJob, fetchSyncJobs, fetchSyncJob, clearError, updateJob } from './syncSlice';