import { configureStore } from '@reduxjs/toolkit';
import chatReducer from './chatSlice.js';

export const store = configureStore({
  reducer: {
    chat: chatReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch; 