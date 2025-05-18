/// <reference types="vite/client" />

interface ImportMeta {
  readonly env: {
    readonly DEV: boolean;
    readonly PROD: boolean;
    readonly MODE: string;
    // Add other environment variables you might use
  };
} 