# Film Development Timer App - Implementation Plan

## Overview
We'll create a self-contained web application for film development timing using the SPRINT developer system. The app will allow users to:
- Select film type from a predefined list
- Adjust for push/pull processing
- Toggle optional steps in the development process
- Select development temperature
- Follow a multi-step timer with alerts for each step

## Technology Stack
- **HTML/CSS**: Bootstrap 5 for responsive layout and styling
- **JavaScript**: Lit.js for component-based UI
- **Packaging**: Self-contained single HTML file with embedded JS and CSS

## Implementation Plan

### 1. Data Preparation
- [x] Extract film types and their chart letters from the documentation
- [x] Create a mapping of chart letters to development times at different temperatures
- [x] Define the development process steps with default timings
- [x] Create data structures for push/pull adjustments

### 2. UI Components
- [ ] Create a film selection dropdown component
  ```js
  class FilmSelector extends LitElement {
    static properties = {
      films: { type: Array },
      selectedFilm: { type: String }
    };
    
    render() {
      return html`
        <div class="form-group">
          <label for="film-select">Film Type</label>
          <select id="film-select" class="form-select" @change=${this._handleChange}>
            ${this.films.map(film => html`<option value=${film.id}>${film.name}</option>`)}
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
  ```

- [ ] Create a push/pull selector component
  ```js
  class PushPullSelector extends LitElement {
    static properties = {
      value: { type: Number }
    };
    
    render() {
      return html`
        <div class="form-group">
          <label for="push-pull">Push/Pull (stops)</label>
          <select id="push-pull" class="form-select" @change=${this._handleChange}>
            <option value="-2">Pull 2 stops</option>
            <option value="-1">Pull 1 stop</option>
            <option value="0" selected>Normal</option>
            <option value="1">Push 1 stop</option>
            <option value="2">Push 2 stops</option>
            <option value="3">Push 3 stops</option>
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
  ```

- [ ] Create a temperature selector component
  ```js
  class TemperatureSelector extends LitElement {
    static properties = {
      temperature: { type: Number }
    };
    
    render() {
      return html`
        <div class="form-group">
          <label for="temperature">Development Temperature</label>
          <select id="temperature" class="form-select" @change=${this._handleChange}>
            <option value="18">18°C / 64.5°F</option>
            <option value="20" selected>20°C / 68°F</option>
            <option value="22">22°C / 71.5°F</option>
            <option value="24">24°C / 75°F</option>
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
  ```

- [ ] Create optional steps toggle component
  ```js
  class OptionalStepsSelector extends LitElement {
    static properties = {
      steps: { type: Array }
    };
    
    constructor() {
      super();
      this.steps = [
        { id: 'prewet', name: 'Water Pre-wet', enabled: true },
        { id: 'prewash', name: 'Water Pre-wash', enabled: true },
        { id: 'fixerRemover', name: 'Remove Fixer', enabled: true },
        { id: 'stabilize', name: 'Stabilize', enabled: true }
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
  ```

- [ ] Create the timer component
  ```js
  class DevelopmentTimer extends LitElement {
    static properties = {
      steps: { type: Array },
      currentStep: { type: Number },
      timeRemaining: { type: Number },
      isRunning: { type: Boolean }
    };
    
    constructor() {
      super();
      this.steps = [];
      this.currentStep = -1;
      this.timeRemaining = 0;
      this.isRunning = false;
      this.timer = null;
    }
    
    render() {
      return html`
        <div class="timer-container">
          <div class="current-step">
            ${this.currentStep >= 0 ? html`
              <h3>${this.steps[this.currentStep].name}</h3>
              <div class="time-display">${this._formatTime(this.timeRemaining)}</div>
              <div class="progress">
                <div class="progress-bar" role="progressbar" 
                  style="width: ${this._calculateProgress()}%" 
                  aria-valuenow="${this._calculateProgress()}" 
                  aria-valuemin="0" aria-valuemax="100"></div>
              </div>
            ` : html`
              <h3>Ready to Start</h3>
            `}
          </div>
          
          <div class="controls">
            ${!this.isRunning ? html`
              <button class="btn btn-primary" @click=${this._startTimer}>
                ${this.currentStep < 0 ? 'Start' : 'Resume'}
              </button>
            ` : html`
              <button class="btn btn-secondary" @click=${this._pauseTimer}>Pause</button>
            `}
            <button class="btn btn-danger" @click=${this._resetTimer}>Reset</button>
            ${this.isRunning ? html`
              <button class="btn btn-success" @click=${this._nextStep}>Next Step</button>
            ` : ''}
          </div>
          
          <div class="step-list">
            <h4>Process Steps</h4>
            <ol class="list-group">
              ${this.steps.map((step, index) => html`
                <li class="list-group-item ${index === this.currentStep ? 'active' : ''} 
                  ${index < this.currentStep ? 'list-group-item-success' : ''}">
                  ${step.name} - ${this._formatTime(step.duration)}
                </li>
              `)}
            </ol>
          </div>
        </div>
      `;
    }
    
    _formatTime(seconds) {
      const mins = Math.floor(seconds / 60);
      const secs = seconds % 60;
      return `${mins}:${secs.toString().padStart(2, '0')}`;
    }
    
    _calculateProgress() {
      if (this.currentStep < 0) return 0;
      const totalTime = this.steps[this.currentStep].duration;
      const elapsed = totalTime - this.timeRemaining;
      return Math.floor((elapsed / totalTime) * 100);
    }
    
    _startTimer() {
      if (this.currentStep < 0) {
        this.currentStep = 0;
        this.timeRemaining = this.steps[0].duration;
      }
      
      this.isRunning = true;
      this.timer = setInterval(() => {
        if (this.timeRemaining > 0) {
          this.timeRemaining--;
        } else {
          this._playAlert();
          this._nextStep();
        }
      }, 1000);
    }
    
    _pauseTimer() {
      this.isRunning = false;
      clearInterval(this.timer);
    }
    
    _resetTimer() {
      this._pauseTimer();
      this.currentStep = -1;
    }
    
    _nextStep() {
      this._pauseTimer();
      if (this.currentStep < this.steps.length - 1) {
        this.currentStep++;
        this.timeRemaining = this.steps[this.currentStep].duration;
        this._startTimer();
      } else {
        this._resetTimer();
      }
    }
    
    _playAlert() {
      // Play sound and show notification
      const audio = new Audio('data:audio/wav;base64,...'); // Embed a short beep sound
      audio.play();
      
      // Show browser notification if supported
      if ('Notification' in window && Notification.permission === 'granted') {
        new Notification('Next Step', {
          body: `Time to move to: ${this.steps[this.currentStep + 1]?.name || 'Finished!'}`
        });
      }
    }
    
    updateSteps(steps) {
      this.steps = steps;
      this._resetTimer();
    }
  }
  ```

### 3. Main Application Logic
- [ ] Create the main application component to coordinate all other components
  ```js
  class DevelopmentApp extends LitElement {
    static properties = {
      filmData: { type: Object },
      selectedFilm: { type: String },
      pushPullValue: { type: Number },
      temperature: { type: Number },
      optionalSteps: { type: Array },
      processSteps: { type: Array }
    };
    
    constructor() {
      super();
      this.filmData = this._loadFilmData();
      this.selectedFilm = '';
      this.pushPullValue = 0;
      this.temperature = 20;
      this.optionalSteps = [
        { id: 'prewet', name: 'Water Pre-wet', enabled: true, duration: 60 },
        { id: 'prewash', name: 'Water Pre-wash', enabled: true, duration: 60 },
        { id: 'fixerRemover', name: 'Remove Fixer', enabled: true, duration: 180 },
        { id: 'stabilize', name: 'Stabilize', enabled: true, duration: 60 }
      ];
      this.processSteps = [];
      
      this._calculateProcessSteps();
    }
    
    render() {
      return html`
        <div class="container">
          <h1>Film Development Timer</h1>
          
          <div class="row">
            <div class="col-md-6">
              <div class="card mb-3">
                <div class="card-header">Film Settings</div>
                <div class="card-body">
                  <film-selector 
                    .films=${Object.entries(this.filmData).map(([id, data]) => ({ id, name: data.name }))}
                    @film-selected=${this._onFilmSelected}>
                  </film-selector>
                  
                  <push-pull-selector
                    @push-pull-changed=${this._onPushPullChanged}>
                  </push-pull-selector>
                  
                  <temperature-selector
                    @temperature-changed=${this._onTemperatureChanged}>
                  </temperature-selector>
                  
                  <optional-steps-selector
                    .steps=${this.optionalSteps}
                    @steps-changed=${this._onOptionalStepsChanged}>
                  </optional-steps-selector>
                </div>
              </div>
            </div>
            
            <div class="col-md-6">
              <div class="card">
                <div class="card-header">Development Timer</div>
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
    
    _loadFilmData() {
      // This would contain all the film data extracted from the documentation
      return {
        'apx100': { name: 'Agfa APX100', chartLetter: 'N' },
        'apx400': { name: 'Agfa APX400', chartLetter: 'R' },
        'panfplus': { name: 'Ilford PanF+', chartLetter: 'N' },
        'fp4plus': { name: 'Ilford FP4+', chartLetter: 'N' },
        'hp5plus': { name: 'Ilford HP5+', chartLetter: 'O' },
        // ... more films
      };
    }
    
    _getChartTimes() {
      // Development times in seconds for each chart letter at different temperatures
      return {
        'L': { '18': 480, '20': 390, '22': 315, '24': 255 },
        'M': { '18': 555, '20': 450, '22': 360, '24': 300 },
        'N': { '18': 630, '20': 510, '22': 420, '24': 330 },
        'O': { '18': 750, '20': 600, '22': 480, '24': 390 },
        'P': { '18': 840, '20': 690, '22': 555, '24': 450 },
        'Q': { '18': 960, '20': 780, '22': 630, '24': 510 },
        'R': { '18': 1080, '20': 900, '22': 750, '24': 600 },
        'S': { '18': 1260, '20': 1020, '22': 840, '24': 675 },
        'T': { '18': 1500, '20': 1200, '22': 960, '24': 780 }
      };
    }
    
    _calculateDevelopmentTime() {
      if (!this.selectedFilm) return 0;
      
      const film = this.filmData[this.selectedFilm];
      let chartLetter = film.chartLetter;
      
      // Adjust chart letter based on push/pull value
      // Each push/pull stop typically moves 1-2 chart letters
      const letterOrder = ['L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T'];
      const baseIndex = letterOrder.indexOf(chartLetter);
      const adjustedIndex = Math.max(0, Math.min(letterOrder.length - 1, baseIndex + this.pushPullValue));
      const adjustedLetter = letterOrder[adjustedIndex];
      
      // Get development time from chart
      const chartTimes = this._getChartTimes();
      return chartTimes[adjustedLetter][this.temperature.toString()];
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
      this.shadowRoot.querySelector('#timer').updateSteps(this.processSteps);
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
  ```

### 4. HTML Structure and Integration
- [ ] Create the main HTML file with all necessary dependencies
  ```html
  <!DOCTYPE html>
  <html lang="en">
  <head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Film Development Timer</title>
    
    <!-- Bootstrap CSS -->
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    
    <!-- Lit.js -->
    <script type="module">
      import { LitElement, html, css } from 'https://cdn.jsdelivr.net/gh/lit/dist@2/core/lit-core.min.js';
      
      // Component definitions will go here
      
      // Register custom elements
      customElements.define('film-selector', FilmSelector);
      customElements.define('push-pull-selector', PushPullSelector);
      customElements.define('temperature-selector', TemperatureSelector);
      customElements.define('optional-steps-selector', OptionalStepsSelector);
      customElements.define('development-timer', DevelopmentTimer);
      customElements.define('development-app', DevelopmentApp);
    </script>
    
    <style>
      /* Additional custom styles */
      .timer-container {
        margin-bottom: 20px;
      }
      
      .time-display {
        font-size: 3rem;
        font-weight: bold;
        text-align: center;
        margin: 20px 0;
      }
      
      .step-list {
        margin-top: 20px;
      }
    </style>
  </head>
  <body>
    <development-app></development-app>
    
    <!-- Bootstrap JS Bundle with Popper -->
    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
  </body>
  </html>
  ```

### 5. Data Extraction and Preparation
- [ ] Extract all film types and their chart letters from the documentation
- [ ] Create a comprehensive mapping of development times
- [ ] Define default process steps with their durations
- [ ] Create a push/pull adjustment logic

### 6. Testing and Refinement
- [ ] Test the application with various film types and settings
- [ ] Ensure timer functionality works correctly
- [ ] Test notifications and alerts
- [ ] Optimize for mobile devices
- [ ] Add error handling and validation

## Implementation Approach
1. First, create the basic HTML structure with Bootstrap and Lit.js imports
2. Extract and organize all the data from the documentation
3. Implement the UI components one by one
4. Integrate the components into the main application
5. Test and refine the application
6. Package everything into a single self-contained HTML file

## Additional Features (if time permits)
- Save settings to localStorage for persistence
- Add dark mode toggle
- Add a history of previous development sessions
- Add custom film type support
- Add custom process step support
- Add print development support 