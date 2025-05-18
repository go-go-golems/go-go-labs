import React from 'react';
import type { Meta, StoryObj } from '@storybook/react';
import { WidgetList } from './WidgetList';
import { handlers } from '../mocks/handlers';
import { http, HttpResponse, delay } from 'msw';
import type { Widget } from '../services/widgetsApi';

// Add a decorator for styling context
const withStyleWrapper = (Story: any) => (
  <div style={{ maxWidth: '500px', margin: '0 auto', padding: '20px' }}>
    <Story />
  </div>
);

const meta: Meta<typeof WidgetList> = {
  title: 'Widgets/WidgetList',
  component: WidgetList,
  parameters: {
    msw: {
      handlers: handlers,
    },
  },
  // Apply the style wrapper to all stories
  decorators: [withStyleWrapper],
};
export default meta;

export const Default: StoryObj<typeof WidgetList> = {};

// Many widgets story with custom data
export const ManyWidgets: StoryObj<typeof WidgetList> = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/widgets', () => {
          return HttpResponse.json([
            { id: 1, name: 'Widget A' },
            { id: 2, name: 'Widget B' },
            { id: 3, name: 'Widget C' },
            { id: 4, name: 'Widget D' },
            { id: 5, name: 'Widget E' },
          ]);
        }),
      ],
    },
  },
};

// Loading state with delayed response
export const Loading: StoryObj<typeof WidgetList> = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/widgets', async () => {
          // Simulate slow network
          await delay(2000);
          return HttpResponse.json([
            { id: 1, name: 'Widget A' },
            { id: 2, name: 'Widget B' },
          ]);
        }),
      ],
    },
  },
};

// Empty list story
export const EmptyList: StoryObj<typeof WidgetList> = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/widgets', () => {
          return HttpResponse.json([]);
        }),
      ],
    },
  },
};

// Error state story
export const ErrorState: StoryObj<typeof WidgetList> = {
  parameters: {
    msw: {
      handlers: [
        http.get('/api/widgets', () => {
          return HttpResponse.error();
        }),
      ],
    },
  },
}; 