import { configureStore } from '@reduxjs/toolkit';
import itineraryReducer from './itinerarySlice';
import mapReducer from './mapSlice';

export const store = configureStore({
  reducer: {
    itineraries: itineraryReducer,
    map: mapReducer
  }
});

// Helper types for useSelector and useDispatch hooks in TypeScript:
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;