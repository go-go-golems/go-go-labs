import React from 'react';
import type { Meta, StoryObj } from '@storybook/react';
import { WidgetManager } from './WidgetManager';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { widgetsApi } from '../services/widgetsApi';
import featuredWidgetReducer, { featuredWidgetSlice } from '../features/featuredWidget/featuredWidgetSlice';
import { http, HttpResponse, delay } from 'msw';
import { RootState } from '../store';

const rootReducer = {
  featuredWidget: featuredWidgetReducer,
  [widgetsApi.reducerPath]: widgetsApi.reducer,
} as const;

// Create a test store that includes both the widgetsApi and featuredWidget
const createTestStore = (initialState?: RootState) => {
  return configureStore({
    reducer: rootReducer,
    // @ts-ignore - Ignoring type checking for middleware for Storybook
    middleware: (getDefaultMiddleware) => 
      getDefaultMiddleware().concat(widgetsApi.middleware),
    preloadedState: initialState,
  });
};

// Create a decorator to provide the custom Redux store
const withStore = (initialState?: any) => (Story: any) => {
  const store = createTestStore(initialState);
  return (
    <Provider store={store}>
      <div style={{ maxWidth: '800px', margin: '0 auto', padding: '20px' }}>
        <Story />
      </div>
    </Provider>
  );
};

const meta: Meta<typeof WidgetManager> = {
  title: 'Widgets/WidgetManager',
  component: WidgetManager,
  decorators: [],
};
export default meta;

// Default state with mocked widgets
export const Default: StoryObj<typeof WidgetManager> = {
  decorators: [withStore()],
  parameters: {
    msw: {
      handlers: [
        http.get('/api/widgets', () => {
          return HttpResponse.json([
            { id: 1, name: 'Widget One' },
            { id: 2, name: 'Widget Two' },
            { id: 3, name: 'Widget Three' },
          ]);
        }),
      ],
    },
  },
};

// Loading state
export const Loading: StoryObj<typeof WidgetManager> = {
  decorators: [withStore()],
  parameters: {
    msw: {
      handlers: [
        http.get('/api/widgets', async () => {
          await delay(3000);
          return HttpResponse.json([
            { id: 1, name: 'Widget One' },
            { id: 2, name: 'Widget Two' },
            { id: 3, name: 'Widget Three' },
          ]);
        }),
      ],
    },
  },
};

// With a pre-selected featured widget
export const WithPreSelectedWidget: StoryObj<typeof WidgetManager> = {
  decorators: [
    withStore({
      featuredWidget: {
        widget: { id: 2, name: 'Widget Two' },
        isHighlighted: false,
      },
    }),
  ],
  parameters: {
    msw: {
      handlers: [
        http.get('/api/widgets', () => {
          return HttpResponse.json([
            { id: 1, name: 'Widget One' },
            { id: 2, name: 'Widget Two' },
            { id: 3, name: 'Widget Three' },
          ]);
        }),
      ],
    },
  },
};

// With highlighted widget
export const WithHighlightedWidget: StoryObj<typeof WidgetManager> = {
  decorators: [
    withStore({
      featuredWidget: {
        widget: { id: 3, name: 'Widget Three' },
        isHighlighted: true,
      },
    }),
  ],
  parameters: {
    msw: {
      handlers: [
        http.get('/api/widgets', () => {
          return HttpResponse.json([
            { id: 1, name: 'Widget One' },
            { id: 2, name: 'Widget Two' },
            { id: 3, name: 'Widget Three' },
          ]);
        }),
      ],
    },
  },
}; 