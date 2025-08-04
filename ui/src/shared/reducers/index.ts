import { combineReducers } from '@reduxjs/toolkit';
import albumsReducer from './albumsSlice';
import mediaReducer from './mediaSlice';
import timelineReducer from './timelineSlice';

// Combine all reducers
const rootReducer = combineReducers({
  albums: albumsReducer,
  media: mediaReducer,
  timeline: timelineReducer,
});

export type RootState = ReturnType<typeof rootReducer>;
export default rootReducer;