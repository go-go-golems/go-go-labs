import { html } from 'https://cdn.jsdelivr.net/gh/lit/dist@2/core/lit-core.min.js';
import { BootstrapLitElement } from './base-component.js';

// Development Timer Component
export class DevelopmentTimer extends BootstrapLitElement {
  static properties = {
    steps: { type: Array },
    currentStep: { type: Number },
    timeRemaining: { type: Number },
    isRunning: { type: Boolean },
    originalDurations: { type: Array }
  };
  
  constructor() {
    super();
    this.steps = [];
    this.currentStep = -1;
    this.timeRemaining = 0;
    this.isRunning = false;
    this.timer = null;
    this.originalDurations = [];
  }
  
  render() {
    return html`
      <div class="timer-container">
        <div class="current-step">
          ${this.currentStep >= 0 ? html`
            <h3>${this.steps[this.currentStep].name}</h3>
            <div class="time-display">${this._formatTime(this.timeRemaining)}</div>
            <div class="progress mb-3">
              <div class="progress-bar" role="progressbar" 
                style="width: ${this._calculateProgress()}%" 
                aria-valuenow="${this._calculateProgress()}" 
                aria-valuemin="0" aria-valuemax="100"></div>
            </div>
          ` : html`
            <h3 class="text-center">Ready to Start</h3>
            <p class="text-center text-muted">Select a film and configure settings, then press Start</p>
          `}
        </div>
        
        <div class="controls text-center mb-4">
          ${!this.isRunning ? html`
            <button class="btn btn-primary" @click=${this._startTimer} ?disabled=${this.steps.length === 0}>
              ${this.currentStep < 0 ? 'Start' : 'Resume'}
            </button>
          ` : html`
            <button class="btn btn-secondary" @click=${this._pauseTimer}>Pause</button>
          `}
          <button class="btn btn-danger" @click=${this._resetTimer} ?disabled=${this.currentStep < 0}>Reset</button>
          ${this.currentStep >= 0 ? html`
            <button class="btn btn-warning" @click=${this._restartCurrentStep} ?disabled=${this.currentStep < 0}>
              Restart Step
            </button>
            <button class="btn btn-info" @click=${this._extendCurrentStep} ?disabled=${this.currentStep < 0}>
              +30s
            </button>
          ` : ''}
          ${this.isRunning ? html`
            <button class="btn btn-success" @click=${this._nextStep}>Next Step</button>
          ` : ''}
        </div>
        
        <div class="step-list">
          <h4>Process Steps</h4>
          <ol class="list-group">
            ${this.steps.map((step, index) => html`
              <li class="list-group-item ${index === this.currentStep ? 'active' : ''} 
                ${index < this.currentStep ? 'list-group-item-success' : ''}
                ${index > this.currentStep ? 'cursor-pointer' : ''}
                ${index <= this.currentStep ? 'cursor-pointer' : ''}"
                @click=${() => this._goToStep(index)}>
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
    if (this.steps.length === 0) return;
    
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
    this.timeRemaining = 0;
    this.requestUpdate();
  }
  
  _nextStep() {
    this._pauseTimer();
    if (this.currentStep < this.steps.length - 1) {
      this.currentStep++;
      this.timeRemaining = this.steps[this.currentStep].duration;
      this._startTimer();
    } else {
      // Process complete
      this._playCompleteAlert();
      this._resetTimer();
    }
  }
  
  _playAlert() {
    // Play sound for step change
    try {
      // Simple beep sound (base64 encoded short WAV)
      const beepSound = 'data:audio/wav;base64,UklGRl9vT19XQVZFZm10IBAAAAABAAEAQB8AAEAfAAABAAgAZGF0YU'+Array(20).join('A');
      const audio = new Audio(beepSound);
      audio.play();
    } catch (e) {
      console.log('Audio playback failed:', e);
    }
    
    // Show browser notification if supported and permitted
    if ('Notification' in window) {
      if (Notification.permission === 'granted') {
        new Notification('Next Step', {
          body: `Time to move to: ${this.steps[this.currentStep + 1]?.name || 'Finished!'}`
        });
      } else if (Notification.permission !== 'denied') {
        Notification.requestPermission();
      }
    }
  }
  
  _playCompleteAlert() {
    try {
      // Different sound for completion (base64 encoded short WAV)
      const completeSound = 'data:audio/wav;base64,UklGRl9vT19XQVZFZm10IBAAAAABAAEAQB8AAEAfAAABAAgAZGF0YU'+Array(30).join('A');
      const audio = new Audio(completeSound);
      audio.play();
    } catch (e) {
      console.log('Audio playback failed:', e);
    }
    
    // Show browser notification for completion
    if ('Notification' in window && Notification.permission === 'granted') {
      new Notification('Development Complete!', {
        body: 'All steps have been completed.'
      });
    }
  }
  
  _restartCurrentStep() {
    if (this.currentStep < 0) return;
    
    // Reset the time remaining to the original duration for this step
    this.timeRemaining = this.steps[this.currentStep].duration;
    
    // If the timer is running, it will continue from the new time
    // If it's paused, the user can resume when ready
    this.requestUpdate();
  }
  
  _extendCurrentStep() {
    if (this.currentStep < 0) return;
    
    // Add 30 seconds to the current time remaining
    this.timeRemaining += 30;
    
    // Also update the step's duration so the progress bar calculation is correct
    this.steps[this.currentStep].duration += 30;
    
    this.requestUpdate();
  }
  
  _goToStep(stepIndex) {
    // Only allow going to steps that are current or previous
    if (stepIndex < 0 || stepIndex > this.steps.length - 1) return;
    
    // Pause the current timer
    this._pauseTimer();
    
    // Set the current step to the selected one
    this.currentStep = stepIndex;
    
    // Set the time remaining to the duration of the selected step
    this.timeRemaining = this.steps[stepIndex].duration;
    
    // Update the UI
    this.requestUpdate();
  }
  
  updateSteps(steps) {
    this.steps = steps;
    // Store original durations for potential resets
    this.originalDurations = steps.map(step => step.duration);
    this._resetTimer();
  }
} 