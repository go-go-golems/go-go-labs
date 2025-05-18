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