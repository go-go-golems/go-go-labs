import { LitElement } from 'https://cdn.jsdelivr.net/gh/lit/dist@2/core/lit-core.min.js';

// Create a function to include Bootstrap CSS in shadow DOM
function createBootstrapLink() {
  const link = document.createElement('link');
  link.rel = 'stylesheet';
  link.href = 'https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css';
  return link;
}

// Base class for all components that need Bootstrap
export class BootstrapLitElement extends LitElement {
  firstUpdated() {
    super.firstUpdated();
    // Add Bootstrap CSS to shadow DOM
    this.shadowRoot.appendChild(createBootstrapLink());
  }
} 