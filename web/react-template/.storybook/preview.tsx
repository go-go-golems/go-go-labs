import React from 'react';
import { initialize, mswLoader } from 'msw-storybook-addon';
import { Provider } from 'react-redux';
import { store } from '../src/store';
import type { Preview } from '@storybook/react';

// Import CSS files - order matters!
import './storybook.css'; // Import Storybook-specific CSS first
import '../src/index.css'; // Then application CSS
import '../src/App.css';   // Component-specific CSS last

// Initialize MSW
initialize({
  serviceWorker: {
    url: './mockServiceWorker.js',
  },
});

const preview: Preview = {
  loaders: [mswLoader],
  parameters: {
    backgrounds: {
      default: 'dark',
      values: [
        { name: 'dark', value: '#242424' },
        { name: 'light', value: '#f8f8f8' },
      ],
    },
  },
  decorators: [
    (Story) => (
      <Provider store={store}>
        <div className="storybook-container">
          <Story />
        </div>
      </Provider>
    ),
  ],
};

export default preview; 