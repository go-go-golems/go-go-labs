# Setting Up a React Project with Bun, Vite, TypeScript, RTK, MSW, and Storybook

This guide will walk you through setting up a complete React development environment with modern tools and best practices. The stack includes:

- **Bun**: Fast JavaScript runtime and package manager
- **Vite**: Lightning-fast build tool and development server
- **React 19**: With TypeScript and SWC for type safety and fast compilation
- **Redux Toolkit**: For state management with RTK Query
- **Storybook**: For component development and testing
- **MSW (Mock Service Worker)**: For API mocking in both development and testing

## 1. Install Prerequisites

### Install Bun

```bash
# Install Bun on macOS/Linux
/bin/bash -c "$(curl -fsSL https://bun.sh/install)"

# Start a new shell to load Bun
exec /bin/zsh
```

## 2. Create the Project

### Scaffold a React + TypeScript Project

```bash
# Create a new Vite project with React, TypeScript, and SWC
bun create vite my-app --template react-swc-ts

# Navigate to project folder
cd my-app

# Install dependencies
bun install
```

### Add Redux Toolkit and React-Redux

```bash
bun add @reduxjs/toolkit react-redux
```

## 3. Add Storybook

```bash
# Install Storybook with Vite builder
bunx storybook@latest init --type react --builder vite --yes
```

## 4. Add MSW for API Mocking

```bash
# Install MSW and Storybook addon
bun add -D msw msw-storybook-addon

# Initialize MSW service worker in public directory
bunx msw init public/ --save
```

## 5. Configure Project Files

### Configure TypeScript

Create or update `tsconfig.json`:

```json
{
  "compilerOptions": {
    "target": "ES2020",
    "useDefineForClassFields": true,
    "lib": ["ES2020", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "skipLibCheck": true,

    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "resolveJsonModule": true,
    "isolatedModules": true,
    "noEmit": true,
    "jsx": "react-jsx",

    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noFallthroughCasesInSwitch": true
  },
  "include": ["src"],
  "references": [{ "path": "./tsconfig.node.json" }]
}
```

Create `tsconfig.node.json`:

```json
{
  "compilerOptions": {
    "composite": true,
    "skipLibCheck": true,
    "module": "ESNext",
    "moduleResolution": "bundler",
    "allowSyntheticDefaultImports": true
  },
  "include": ["vite.config.ts"]
}
```

### Add Vite Type Declarations

Create `src/vite-env.d.ts`:

```typescript
/// <reference types="vite/client" />

interface ImportMeta {
  readonly env: {
    readonly DEV: boolean;
    readonly PROD: boolean;
    readonly MODE: string;
    // Add other environment variables you might use
  };
}
```

### Configure Vite with API Proxy

Create or update `vite.config.ts`:

```typescript
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react-swc';

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': {         // anything starting with /api →
        target: 'http://localhost:8080',
        changeOrigin: true,
        secure: false,  // allow self-signed HTTPS while you iterate
      },
    },
  },
});
```

## 6. Set Up Redux Store and RTK Query

### Create API Service with RTK Query

Create `src/services/widgetsApi.ts`:

```typescript
import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';

export interface Widget { id: number; name: string }

export const widgetsApi = createApi({
  reducerPath: 'widgetsApi',
  baseQuery: fetchBaseQuery({ baseUrl: '/api' }),
  // Add tag-based invalidation for cache control
  tagTypes: ['Widgets'],
  endpoints: builder => ({
    getWidgets: builder.query<Widget[], void>({
      query: () => 'widgets',
      // Provide a tag to the cache result
      providesTags: ['Widgets'],
      // Customize how the response is cached
      keepUnusedDataFor: 30, // in seconds
    }),
  }),
});

export const { useGetWidgetsQuery } = widgetsApi;
```

### Create Redux Store with Feature Slice

Create `src/features/featuredWidget/featuredWidgetSlice.ts`:

```typescript
import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { Widget } from '../../services/widgetsApi';

interface FeaturedWidgetState {
  widget: Widget | null;
  isHighlighted: boolean;
}

const initialState: FeaturedWidgetState = {
  widget: null,
  isHighlighted: false,
};

export const featuredWidgetSlice = createSlice({
  name: 'featuredWidget',
  initialState,
  reducers: {
    setFeaturedWidget: (state, action: PayloadAction<Widget>) => {
      state.widget = action.payload;
    },
    clearFeaturedWidget: (state) => {
      state.widget = null;
    },
    toggleHighlight: (state) => {
      state.isHighlighted = !state.isHighlighted;
    },
    setHighlight: (state, action: PayloadAction<boolean>) => {
      state.isHighlighted = action.payload;
    },
  },
});

export const {
  setFeaturedWidget,
  clearFeaturedWidget,
  toggleHighlight,
  setHighlight,
} = featuredWidgetSlice.actions;

export default featuredWidgetSlice.reducer;
```

### Configure the Redux Store

Create `src/store.ts`:

```typescript
import { configureStore, createSlice } from '@reduxjs/toolkit';
import { widgetsApi } from './services/widgetsApi';
import featuredWidgetReducer from './features/featuredWidget/featuredWidgetSlice';

export const store = configureStore({
  reducer: {
    featuredWidget: featuredWidgetReducer,
    [widgetsApi.reducerPath]: widgetsApi.reducer,
  },
  middleware: getDefaultMiddleware => 
    getDefaultMiddleware().concat(widgetsApi.middleware),
});
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
```

## 7. Set Up MSW for API Mocking

### Create MSW Handlers

Create `src/mocks/handlers.ts`:

```typescript
import { http, HttpResponse } from 'msw';
import type { Widget } from '../services/widgetsApi';

const widgets: Widget[] = [
  { id: 1, name: 'Foo' },
  { id: 2, name: 'Bar' },
];

export const handlers = [
  http.get('/api/widgets', () => {
    return HttpResponse.json(widgets);
  }),
];
```

### Set Up MSW Browser Integration

Create `src/mocks/browser.ts`:

```typescript
import { setupWorker } from 'msw/browser';
import { handlers } from './handlers';

export const worker = setupWorker(...handlers);
```

### Configure MSW in the Application Entry Point

Update `src/main.tsx`:

```typescript
import React from 'react';
import ReactDOM from 'react-dom/client';
import { Provider } from 'react-redux';
import { store } from './store';
import App from './App';
import './index.css';

async function start() {
  if (import.meta.env.DEV) {
    const { worker } = await import('./mocks/browser');
    worker.start();
  }

  ReactDOM.createRoot(document.getElementById('root')!).render(
    <React.StrictMode>
      <Provider store={store}>
        <App />
      </Provider>
    </React.StrictMode>,
  );
}

start();
```

## 8. Create React Components

### Create the WidgetList Component

Create `src/components/WidgetList.tsx`:

```tsx
import React, { useEffect } from 'react';
import { useDispatch } from 'react-redux';
import { useGetWidgetsQuery } from '../services/widgetsApi';
import { setFeaturedWidget } from '../features/featuredWidget/featuredWidgetSlice';
import type { AppDispatch } from '../store';

export const WidgetList = () => {
  const { 
    data, 
    isLoading, 
    error,
    refetch,
    isFetching
  } = useGetWidgetsQuery(undefined, {
    // Disable caching for this example to see changes immediately
    refetchOnMountOrArgChange: true,
    // Refetch on window focus
    refetchOnFocus: true
  });
  
  const dispatch = useDispatch<AppDispatch>();

  // Force refetch on component mount
  useEffect(() => {
    refetch();
  }, [refetch]);

  if (isLoading || isFetching) return <>Loading…</>;
  if (error) return <>Error</>;
  if (!data || data.length === 0) return <>No widgets found</>;

  return (
    <div>
      <ul className="widget-list">
        {data.map(widget => (
          <li key={widget.id} className="widget-list-item">
            {widget.name}
            <button 
              onClick={() => dispatch(setFeaturedWidget(widget))}
              className="feature-button"
            >
              Feature
            </button>
          </li>
        ))}
      </ul>
      <button 
        onClick={() => refetch()}
        className="refresh-button"
      >
        Refresh Widgets
      </button>
    </div>
  );
};
```

### Create the FeaturedWidget Component

Create `src/components/FeaturedWidget.tsx`:

```tsx
import React from 'react';
import { useSelector, useDispatch } from 'react-redux';
import type { RootState, AppDispatch } from '../store';
import { toggleHighlight, clearFeaturedWidget } from '../features/featuredWidget/featuredWidgetSlice';

export const FeaturedWidget: React.FC = () => {
  const dispatch = useDispatch<AppDispatch>();
  const { widget, isHighlighted } = useSelector((state: RootState) => state.featuredWidget);

  if (!widget) {
    return <div className="featured-widget-empty">No widget is currently featured</div>;
  }

  return (
    <div className={`featured-widget ${isHighlighted ? 'highlighted' : ''}`}>
      <h3>Featured Widget</h3>
      <div className="widget-detail">
        <strong>ID:</strong> {widget.id}
      </div>
      <div className="widget-detail">
        <strong>Name:</strong> {widget.name}
      </div>
      <div className="widget-actions">
        <button onClick={() => dispatch(toggleHighlight())}>
          {isHighlighted ? 'Remove Highlight' : 'Highlight'}
        </button>
        <button onClick={() => dispatch(clearFeaturedWidget())}>
          Clear
        </button>
      </div>
    </div>
  );
};
```

### Create a Combined WidgetManager Component

Create `src/components/WidgetManager.tsx`:

```tsx
import React from 'react';
import { WidgetList } from './WidgetList';
import { FeaturedWidget } from './FeaturedWidget';

export const WidgetManager: React.FC = () => {
  return (
    <div className="widget-manager">
      <div className="widget-manager-left">
        <h2>Available Widgets</h2>
        <WidgetList />
      </div>
      <div className="widget-manager-right">
        <FeaturedWidget />
      </div>
    </div>
  );
};
```

### Add App Component

Update `src/App.tsx`:

```tsx
import React from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { WidgetManager } from './components/WidgetManager';
import './App.css';

function App() {

  return (
    <div className="app">
      <h1>Vite + React + Redux</h1>
      
      <WidgetManager />
    </div>
  );
}

export default App;
```

## 9. Add Styling

Create or update `src/App.css`:

```css
.app {
  max-width: 800px;
  margin: 0 auto;
  padding: 2rem;
  text-align: center;
}

.card {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  margin: 1rem 0;
}

button {
  border-radius: 8px;
  border: 1px solid transparent;
  padding: 0.6em 1.2em;
  font-size: 1em;
  font-weight: 500;
  font-family: inherit;
  background-color: #1a1a1a;
  color: white;
  cursor: pointer;
  transition: border-color 0.25s;
}

button:hover {
  border-color: #646cff;
}

button:focus,
button:focus-visible {
  outline: 4px auto -webkit-focus-ring-color;
}

/* Widget Manager */
.widget-manager {
  display: flex;
  gap: 2rem;
  margin-top: 2rem;
}

.widget-manager-left {
  flex: 1;
}

.widget-manager-right {
  flex: 1;
}

/* Widget List Styles */
.widget-list {
  list-style: none;
  padding: 0;
  margin: 1rem 0;
  text-align: left;
}

.widget-list-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem;
  border-bottom: 1px solid #333;
}

.widget-list-item:last-child {
  border-bottom: none;
}

.feature-button {
  font-size: 0.8em;
  padding: 0.4em 0.8em;
  background-color: #3b3b3b;
}

.refresh-button {
  margin-top: 1rem;
  width: 100%;
  font-size: 0.9em;
  background-color: #2c2c3a;
  border: 1px solid #444;
  transition: all 0.3s ease;
}

.refresh-button:hover {
  background-color: #3c3c4a;
  border-color: #646cff;
}

/* Featured Widget Styles */
.featured-widget {
  margin: 2rem 0;
  padding: 1.5rem;
  border-radius: 8px;
  background-color: #242424;
  border: 1px solid #444;
  transition: all 0.3s ease;
}

.featured-widget.highlighted {
  background-color: #2a2a4a;
  border-color: #646cff;
  box-shadow: 0 0 15px rgba(100, 108, 255, 0.3);
}

.featured-widget-empty {
  margin: 2rem 0;
  padding: 1.5rem;
  border-radius: 8px;
  background-color: #242424;
  border: 1px dashed #444;
  color: #888;
}

.widget-detail {
  margin: 0.5rem 0;
  text-align: left;
}

.widget-actions {
  display: flex;
  gap: 1rem;
  margin-top: 1rem;
  justify-content: flex-end;
}
```

## 10. Configure Storybook

### Configure Storybook Main File

Update `.storybook/main.ts`:

```typescript
import { StorybookConfig } from '@storybook/react-vite';

const config: StorybookConfig = {
  stories: ['../src/**/*.stories.@(ts|tsx)'],
  addons: ['msw-storybook-addon'],
  framework: { name: '@storybook/react-vite', options: {} },
  staticDirs: ['../public'],
};
export default config;
```

### Create Storybook Preview File

Create `.storybook/preview.tsx`:

```tsx
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
```

### Add Storybook-Specific CSS

Create `.storybook/storybook.css`:

```css
/* Basic styling for Storybook */
#storybook-root {
  font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
  line-height: 1.5;
  color: rgba(255, 255, 255, 0.87);
  background-color: #242424;
  padding: 1rem;
}

.storybook-container {
  max-width: 800px;
  margin: 0 auto;
  padding: 1rem;
}

/* Widget styles duplicated from App.css for Storybook context */
/* ...add the same CSS as App.css... */
```

## 11. Create Stories for Components

### Create WidgetList Stories

Create `src/components/WidgetList.stories.tsx`:

```tsx
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
```

### Create FeaturedWidget Stories

Create `src/components/FeaturedWidget.stories.tsx`:

```tsx
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
```

### Create WidgetManager Stories

Create `src/components/WidgetManager.stories.tsx`:

```tsx
import React from 'react';
import type { Meta, StoryObj } from '@storybook/react';
import { WidgetManager } from './WidgetManager';
import { Provider } from 'react-redux';
import { configureStore } from '@reduxjs/toolkit';
import { widgetsApi } from '../services/widgetsApi';
import featuredWidgetReducer from '../features/featuredWidget/featuredWidgetSlice';
import { http, HttpResponse, delay } from 'msw';

// Create a test store that includes both the widgetsApi and featuredWidget
const createTestStore = (initialState?: any) => {
  return configureStore({
    reducer: {
      featuredWidget: featuredWidgetReducer,
      [widgetsApi.reducerPath]: widgetsApi.reducer,
    },
    middleware: (getDefaultMiddleware) => 
      getDefaultMiddleware().concat(widgetsApi.middleware) as any,
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
```

## 12. Run the Application

### Run Development Server

```bash
bun run dev
```

This starts the Vite development server, usually at http://localhost:5173.

### Run Storybook

```bash
bun run storybook
```

This starts Storybook, usually at http://localhost:6006.

## 13. Key Concepts and Best Practices

### RTK Query Caching

RTK Query implements sophisticated caching to minimize network requests:

- Use `refetchOnMountOrArgChange` to control when data refetches
- Use `providesTags` and `invalidatesTags` for cache invalidation
- Add a manual refresh button for user-controlled refetching
- Implement optimistic updates for a smoother UX

### MSW for API Mocking

MSW intercepts network requests at the service worker level:

- Define handlers for each API endpoint
- Create different handler variations for testing different states
- Use response transformers for specific test cases
- Conditionally start MSW only in development mode

### Storybook with Redux Integration

Storybook allows testing isolated components:

- Create custom store configurations for each story
- Use decorators to provide Redux context
- Override global styles for Storybook-specific styling
- Use MSW handlers per story to control API responses

### TypeScript Type Safety

TypeScript enhances code quality and developer experience:

- Define interfaces for API responses
- Type Redux state and actions
- Use utility types for better type inference
- Define type-safe hooks for accessing store state

## 14. Troubleshooting Common Issues

### Storybook CSS Loading Issues

If CSS isn't properly loading in Storybook:

1. Import CSS files in the Storybook preview file
2. Create Storybook-specific CSS if needed
3. Use style decorators for component-specific styling
4. Check import order - global styles should come before component styles

### MSW Integration Problems

If MSW isn't intercepting requests:

1. Ensure the service worker is properly initialized
2. Check that handlers match the exact API endpoints
3. Verify that MSW is initialized before component rendering
4. Use the browser DevTools Network tab to debug request issues

### Redux Store Integration

If components aren't updating with Redux changes:

1. Check that components are wrapped with `Provider`
2. Verify selectors are correctly accessing
