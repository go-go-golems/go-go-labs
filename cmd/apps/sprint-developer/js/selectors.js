import { html } from 'https://cdn.jsdelivr.net/gh/lit/dist@2/core/lit-core.min.js';
import { BootstrapLitElement } from './base-component.js';

// 1. Film Selection Component
export class FilmSelector extends BootstrapLitElement {
  static properties = {
    films: { type: Array },
    selectedFilm: { type: String }
  };
  
  constructor() {
    super();
    this.films = [];
    this.selectedFilm = '';
  }
  
  render() {
    return html`
      <div class="form-group">
        <label for="film-select">Film Type</label>
        <select id="film-select" class="form-select" @change=${this._handleChange}>
          <option value="" selected disabled>Select a film</option>
          ${this.films.map(film => html`
            <option value=${film.id} ?selected=${film.id === this.selectedFilm}>
              ${film.name}
            </option>
          `)}
        </select>
      </div>
    `;
  }
  
  _handleChange(e) {
    this.selectedFilm = e.target.value;
    this.dispatchEvent(new CustomEvent('film-selected', { 
      detail: { film: this.selectedFilm }
    }));
  }
}

// 2. Push/Pull Selector Component
export class PushPullSelector extends BootstrapLitElement {
  static properties = {
    value: { type: Number }
  };
  
  constructor() {
    super();
    this.value = 0;
  }
  
  render() {
    return html`
      <div class="form-group">
        <label for="push-pull">Push/Pull (stops)</label>
        <select id="push-pull" class="form-select" @change=${this._handleChange}>
          <option value="-2" ?selected=${this.value === -2}>Pull 2 stops</option>
          <option value="-1" ?selected=${this.value === -1}>Pull 1 stop</option>
          <option value="0" ?selected=${this.value === 0}>Normal</option>
          <option value="1" ?selected=${this.value === 1}>Push 1 stop</option>
          <option value="2" ?selected=${this.value === 2}>Push 2 stops</option>
          <option value="3" ?selected=${this.value === 3}>Push 3 stops</option>
        </select>
      </div>
    `;
  }
  
  _handleChange(e) {
    this.value = parseInt(e.target.value);
    this.dispatchEvent(new CustomEvent('push-pull-changed', { 
      detail: { value: this.value }
    }));
  }
}

// 3. Temperature Selector Component
export class TemperatureSelector extends BootstrapLitElement {
  static properties = {
    temperature: { type: Number }
  };
  
  constructor() {
    super();
    this.temperature = 20;
  }
  
  render() {
    return html`
      <div class="form-group">
        <label for="temperature">Development Temperature</label>
        <select id="temperature" class="form-select" @change=${this._handleChange}>
          <option value="18" ?selected=${this.temperature === 18}>18°C / 64.5°F</option>
          <option value="20" ?selected=${this.temperature === 20}>20°C / 68°F</option>
          <option value="22" ?selected=${this.temperature === 22}>22°C / 71.5°F</option>
          <option value="24" ?selected=${this.temperature === 24}>24°C / 75°F</option>
        </select>
      </div>
    `;
  }
  
  _handleChange(e) {
    this.temperature = parseInt(e.target.value);
    this.dispatchEvent(new CustomEvent('temperature-changed', { 
      detail: { temperature: this.temperature }
    }));
  }
}

// 4. Optional Steps Selector Component
export class OptionalStepsSelector extends BootstrapLitElement {
  static properties = {
    steps: { type: Array }
  };
  
  constructor() {
    super();
    this.steps = [
      { id: 'prewet', name: 'Water Pre-wet', enabled: false },
      { id: 'prewash', name: 'Water Pre-wash', enabled: false },
      { id: 'fixerRemover', name: 'Remove Fixer', enabled: false },
      { id: 'stabilize', name: 'Stabilize', enabled: false }
    ];
  }
  
  render() {
    return html`
      <div class="form-group">
        <label>Optional Steps</label>
        ${this.steps.map(step => html`
          <div class="form-check">
            <input class="form-check-input" type="checkbox" id="${step.id}" 
              ?checked=${step.enabled} @change=${e => this._toggleStep(step.id, e.target.checked)}>
            <label class="form-check-label" for="${step.id}">${step.name}</label>
          </div>
        `)}
      </div>
    `;
  }
  
  _toggleStep(id, enabled) {
    this.steps = this.steps.map(step => 
      step.id === id ? {...step, enabled} : step
    );
    this.dispatchEvent(new CustomEvent('steps-changed', { 
      detail: { steps: this.steps }
    }));
  }
} 