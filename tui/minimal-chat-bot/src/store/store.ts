import { configureStore } from '@reduxjs/toolkit';
import chatReducer from './chatSlice.js';
import scrollReducer from './scrollSlice.js';

export const store = configureStore({
  reducer: {
    chat: chatReducer,
    scroll: scrollReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch; 