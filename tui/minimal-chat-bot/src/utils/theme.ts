import { Theme } from '../types/index.js';

export function getTheme(): Theme {
  return {
    primary: '#9D8CFF',    // Main accent color (purple)
    secondary: '#C2BBF0',  // Secondary accent (lighter purple)
    accent: '#7D6FFF',     // Highlight color (bright purple)
    error: '#FF6B6B',      // Error messages (red)
    text: '#FFFFFF',       // Primary text (white)
    dimText: '#AAAAAA',    // Secondary text (light gray)
    border: '#444444'      // Border color (dark gray)
  };
} 