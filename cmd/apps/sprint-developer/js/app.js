import { html } from 'https://cdn.jsdelivr.net/gh/lit/dist@2/core/lit-core.min.js';
import { BootstrapLitElement } from './base-component.js';
import { filmData, adjustChartLetterForPushPull, chartTimes } from './data.js';
import { FilmSelector, PushPullSelector, TemperatureSelector, OptionalStepsSelector } from './selectors.js';
import { DevelopmentTimer } from './timer.js';
import { Documentation } from './components/Documentation.js';

// Register custom elements
customElements.define('film-selector', FilmSelector);
customElements.define('push-pull-selector', PushPullSelector);
customElements.define('temperature-selector', TemperatureSelector);
customElements.define('optional-steps-selector', OptionalStepsSelector);
customElements.define('development-timer', DevelopmentTimer);
customElements.define('process-documentation', Documentation);

// Main application component
export class DevelopmentApp extends BootstrapLitElement {
  static properties = {
    filmData: { type: Object },
    selectedFilm: { type: String },
    pushPullValue: { type: Number },
    temperature: { type: Number },
    optionalSteps: { type: Array },
    processSteps: { type: Array },
    developmentTime: { type: Number },
    chartLetter: { type: String }
  };
  
  constructor() {
    super();
    this.filmData = filmData;
    this.selectedFilm = '';
    this.pushPullValue = 0;
    this.temperature = 20;
    this.optionalSteps = [
      { id: 'prewet', name: 'Water Pre-wet', enabled: false, duration: 60 },
      { id: 'prewash', name: 'Water Pre-wash', enabled: false, duration: 60 },
      { id: 'fixerRemover', name: 'Remove Fixer', enabled: false, duration: 180 },
      { id: 'stabilize', name: 'Stabilize', enabled: false, duration: 60 }
    ];
    this.processSteps = [];
    this.developmentTime = 0;
    this.chartLetter = '';
    
    // Request notification permission on startup
    if ('Notification' in window && Notification.permission !== 'granted' && Notification.permission !== 'denied') {
      Notification.requestPermission();
    }
  }
  
  render() {
    return html`
      <div class="container">
        <div class="row mb-4">
          <div class="col-12">
            <div class="documentation-container">
              <process-documentation></process-documentation>
            </div>
          </div>
        </div>
        
        <div class="row">
          <div class="col-md-5">
            <div class="card mb-4">
              <div class="card-header bg-primary text-white">
                <h5 class="mb-0">Film Settings</h5>
              </div>
              <div class="card-body">
                <film-selector 
                  .films=${Object.entries(this.filmData).map(([id, data]) => ({ id, name: data.name }))}
                  .selectedFilm=${this.selectedFilm}
                  @film-selected=${this._onFilmSelected}>
                </film-selector>
                
                <push-pull-selector
                  .value=${this.pushPullValue}
                  @push-pull-changed=${this._onPushPullChanged}>
                </push-pull-selector>
                
                <temperature-selector
                  .temperature=${this.temperature}
                  @temperature-changed=${this._onTemperatureChanged}>
                </temperature-selector>
                
                <optional-steps-selector
                  .steps=${this.optionalSteps}
                  @steps-changed=${this._onOptionalStepsChanged}>
                </optional-steps-selector>
              </div>
            </div>
            
            ${this.selectedFilm ? html`
              <div class="card mb-4">
                <div class="card-header bg-info text-white">
                  <h5 class="mb-0">Development Information</h5>
                </div>
                <div class="card-body">
                  <p><strong>Film:</strong> ${this.filmData[this.selectedFilm]?.name}</p>
                  <p><strong>Base Chart Letter:</strong> ${this.filmData[this.selectedFilm]?.chartLetter}</p>
                  <p><strong>Adjusted Chart Letter:</strong> ${this.chartLetter} 
                    ${this.pushPullValue !== 0 ? 
                      `(${this.pushPullValue > 0 ? '+' : ''}${this.pushPullValue} stops)` : 
                      ''}
                  </p>
                  <p><strong>Development Time:</strong> ${this._formatTime(this.developmentTime)}</p>
                  <p><strong>Temperature:</strong> ${this.temperature}°C / ${this._convertToFahrenheit(this.temperature)}°F</p>
                </div>
              </div>
            ` : ''}
          </div>
          
          <div class="col-md-7">
            <div class="card">
              <div class="card-header bg-success text-white">
                <h5 class="mb-0">Development Timer</h5>
              </div>
              <div class="card-body">
                <development-timer
                  id="timer"
                  .steps=${this.processSteps}>
                </development-timer>
              </div>
            </div>
          </div>
        </div>
      </div>
    `;
  }
  
  _formatTime(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  }
  
  _convertToFahrenheit(celsius) {
    return (celsius * 9/5 + 32).toFixed(1);
  }
  
  _calculateDevelopmentTime() {
    if (!this.selectedFilm) {
      this.developmentTime = 0;
      this.chartLetter = '';
      return 0;
    }
    
    const film = this.filmData[this.selectedFilm];
    let chartLetter = film.chartLetter;
    
    // Adjust chart letter based on push/pull value
    chartLetter = adjustChartLetterForPushPull(chartLetter, this.pushPullValue);
    this.chartLetter = chartLetter;
    
    // Get development time from chart
    const time = chartTimes[chartLetter][this.temperature.toString()] || 0;
    this.developmentTime = time;
    return time;
  }
  
  _calculateProcessSteps() {
    const developTime = this._calculateDevelopmentTime();
    
    // Create the process steps array
    this.processSteps = [];
    
    // Add optional pre-wet step
    const prewet = this.optionalSteps.find(s => s.id === 'prewet');
    if (prewet && prewet.enabled) {
      this.processSteps.push({
        name: 'Water Pre-wet',
        duration: prewet.duration
      });
    }
    
    // Add development step
    if (developTime > 0) {
      this.processSteps.push({
        name: 'Develop',
        duration: developTime
      });
    }
    
    // Add stop bath step
    this.processSteps.push({
      name: 'Stop Bath',
      duration: 60 // 1 minute
    });
    
    // Add fixer step
    this.processSteps.push({
      name: 'Fix',
      duration: 180 // 3 minutes
    });
    
    // Add optional pre-wash step
    const prewash = this.optionalSteps.find(s => s.id === 'prewash');
    if (prewash && prewash.enabled) {
      this.processSteps.push({
        name: 'Water Pre-wash',
        duration: prewash.duration
      });
    }
    
    // Add optional fixer remover step
    const fixerRemover = this.optionalSteps.find(s => s.id === 'fixerRemover');
    if (fixerRemover && fixerRemover.enabled) {
      this.processSteps.push({
        name: 'Remove Fixer',
        duration: fixerRemover.duration
      });
    }
    
    // Add water wash step
    this.processSteps.push({
      name: 'Water Wash',
      duration: 300 // 5 minutes
    });
    
    // Add optional stabilize step
    const stabilize = this.optionalSteps.find(s => s.id === 'stabilize');
    if (stabilize && stabilize.enabled) {
      this.processSteps.push({
        name: 'Stabilize',
        duration: stabilize.duration
      });
    }
    
    // Update the timer component
    const timer = this.shadowRoot.querySelector('#timer');
    if (timer) {
      timer.updateSteps(this.processSteps);
    }
  }
  
  firstUpdated() {
    super.firstUpdated();
    // Calculate initial process steps
    this._calculateProcessSteps();
  }
  
  _onFilmSelected(e) {
    this.selectedFilm = e.detail.film;
    this._calculateProcessSteps();
  }
  
  _onPushPullChanged(e) {
    this.pushPullValue = e.detail.value;
    this._calculateProcessSteps();
  }
  
  _onTemperatureChanged(e) {
    this.temperature = e.detail.temperature;
    this._calculateProcessSteps();
  }
  
  _onOptionalStepsChanged(e) {
    this.optionalSteps = e.detail.steps;
    this._calculateProcessSteps();
  }
}

// Register the main application component
customElements.define('development-app', DevelopmentApp);

// Initialize the app when the DOM is loaded
export function initApp() {
  window.addEventListener('DOMContentLoaded', () => {
    const appContainer = document.getElementById('app-container');
    appContainer.innerHTML = '';
    const app = document.createElement('development-app');
    appContainer.appendChild(app);
  });
} 