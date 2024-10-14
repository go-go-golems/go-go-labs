import { LitElement, html, css } from 'https://unpkg.com/lit@2.6.1/index.js?module';

export class CaptureControls extends LitElement {
  static properties = {
    frameCount: { type: Number },
    countdownSpeed: { type: Number },
  };

  static styles = css`
    :host {
      display: inline-block;
    }
    select {
      padding: 5px;
      margin: 0 5px;
    }
  `;

  render() {
    return html`
      <label for="frameCountSelect">Number of Frames: </label>
      <select id="frameCountSelect" @change=${this._handleFrameCountChange}>
        ${[5, 10, 20].map(count => html`
          <option value="${count}" ?selected=${count === this.frameCount}>${count}</option>
        `)}
      </select>

      <label for="countdownSpeedSelect">Countdown Speed (s): </label>
      <select id="countdownSpeedSelect" @change=${this._handleCountdownSpeedChange}>
        ${[1, 2, 3].map(speed => html`
          <option value="${speed}" ?selected=${speed === this.countdownSpeed}>${speed}</option>
        `)}
      </select>
    `;
  }

  _handleFrameCountChange(e) {
    this.dispatchEvent(new CustomEvent('frame-count-change', {
      detail: parseInt(e.target.value, 10),
      bubbles: true,
      composed: true
    }));
  }

  _handleCountdownSpeedChange(e) {
    this.dispatchEvent(new CustomEvent('countdown-speed-change', {
      detail: parseInt(e.target.value, 10),
      bubbles: true,
      composed: true
    }));
  }
}

customElements.define('capture-controls', CaptureControls);
