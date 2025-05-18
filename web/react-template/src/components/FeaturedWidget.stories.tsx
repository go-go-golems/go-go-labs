import React from 'react';
import type { Meta, StoryObj } from '@storybook/react';
import { FeaturedWidget } from './FeaturedWidget';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import featuredWidgetReducer, {
  setFeaturedWidget,
  setHighlight,
} from '../features/featuredWidget/featuredWidgetSlice';

// Create a custom store setup function
const createTestStore = (initialState?: any) => {
  return configureStore({
    reducer: {
      featuredWidget: featuredWidgetReducer,
    },
    preloadedState: initialState ? { featuredWidget: initialState } : undefined
  });
}

// Create a decorator to provide the Redux store with styling context
const withStore = (initialState?: any) => (Story: any) => {
  const store = createTestStore(initialState);
  return (
    <Provider store={store}>
      <div style={{ maxWidth: '600px', margin: '0 auto', padding: '20px' }}>
        <Story />
      </div>
    </Provider>
  );
};

const meta: Meta<typeof FeaturedWidget> = {
  title: 'Widgets/FeaturedWidget',
  component: FeaturedWidget,
  // We don't use any global decorators here because we'll provide our custom store per story
  decorators: [],
  parameters: {
    // Ensure MSW doesn't interfere with these stories
    msw: { handlers: [] },
  },
};
export default meta;

// Empty state (no featured widget)
export const Empty: StoryObj<typeof FeaturedWidget> = {
  decorators: [withStore()],
};

// With widget data
export const WithWidget: StoryObj<typeof FeaturedWidget> = {
  decorators: [
    withStore({
      widget: { id: 1, name: 'Test Widget' },
      isHighlighted: false
    })
  ],
};

// Highlighted widget
export const HighlightedWidget: StoryObj<typeof FeaturedWidget> = {
  decorators: [
    withStore({
      widget: { id: 2, name: 'Premium Widget' },
      isHighlighted: true
    })
  ],
}; 