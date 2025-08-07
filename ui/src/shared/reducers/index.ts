import { combineReducers } from '@reduxjs/toolkit';
import albumsReducer from './albumsSlice';
import mediaReducer from './mediaSlice';
import timelineReducer from './timelineSlice';
import syncReducer from './syncSlice';
import uploadReducer from './uploadSlice';
import statsReducer from './statsSlice';

// Combine all reducers
const rootReducer = combineReducers({
  albums: albumsReducer,
  media: mediaReducer,
  timeline: timelineReducer,
  sync: syncReducer,
  upload: uploadReducer,
  stats: statsReducer,
});

export type RootState = ReturnType<typeof rootReducer>;
export default rootReducer;

// Export sync actions for convenience
export { startSync, cancelSync, updateProgress, clearError } from './syncSlice';