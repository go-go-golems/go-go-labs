import React from 'react';
import ReactDOM from 'react-dom/client';
import { Provider } from 'react-redux';
import { store } from './store';
import App from './App';
import './index.css';

async function start() {
  // Only enable MSW in dev mode when explicitly requested
  // This allows the vite proxy to work for backend calls by default
  if (import.meta.env.DEV && import.meta.env.VITE_USE_MSW === 'true') {
    console.log('Starting MSW in development mode');
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