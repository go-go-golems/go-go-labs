import { initApp } from './app.js';
import { Documentation } from './components/Documentation.js';

// Initialize the application
initApp();

document.addEventListener('DOMContentLoaded', () => {
  const appContainer = document.getElementById('app-container');
  
  // Create documentation component
  const documentation = new Documentation();
  appContainer.appendChild(documentation.element);
}); 