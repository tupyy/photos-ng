import { combineReducers } from '@reduxjs/toolkit';
import albumsReducer from './albumsSlice';
import mediaReducer from './mediaSlice';
import timelineReducer from './timelineSlice';
import syncReducer from './syncSlice';

// Combine all reducers
const rootReducer = combineReducers({
  albums: albumsReducer,
  media: mediaReducer,
  timeline: timelineReducer,
  sync: syncReducer,
});

export type RootState = ReturnType<typeof rootReducer>;
export default rootReducer;

// Export sync actions for convenience
export { startSync, cancelSync, updateProgress, clearError } from './syncSlice';