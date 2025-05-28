import { configureStore } from '@reduxjs/toolkit';
import { setupListeners } from '@reduxjs/toolkit/query';
import { agentFleetApi } from '@/services/api';
import uiReducer from './slices/uiSlice';

export const store = configureStore({
  reducer: {
    [agentFleetApi.reducerPath]: agentFleetApi.reducer,
    ui: uiReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(agentFleetApi.middleware),
});

setupListeners(store.dispatch);

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
