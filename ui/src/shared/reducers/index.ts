import { combineReducers } from '@reduxjs/toolkit';
import albumsReducer from './albumsSlice';
import timelineReducer from './timelineSlice';
import albumsMediaReducer from './albumsMediaSlice';
import uploadReducer from './uploadSlice';
import statsReducer from './statsSlice';
import mediaReducer from './mediaSlice';
import userReducer from './userSlice';

// Combine all reducers
const rootReducer = combineReducers({
  albums: albumsReducer,
  timeline: timelineReducer,
  albumsMedia: albumsMediaReducer,
  upload: uploadReducer,
  stats: statsReducer,
  media: mediaReducer,
  user: userReducer,
});

export type RootState = ReturnType<typeof rootReducer>;
export default rootReducer;

// Export media actions for convenience
export { deleteMedia } from './mediaSlice';

// Export user actions for convenience
export { fetchCurrentUser, clearUser } from './userSlice';