import { LitElement, html, css } from 'https://unpkg.com/lit@2.6.1/index.js?module';

export class CameraSelector extends LitElement {
  static properties = {
    cameras: { type: Array },
    selectedCamera: { type: String },
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
      <label for="cameraSelect">Select Camera: </label>
      <select id="cameraSelect" @change=${this._handleChange}>
        ${this.cameras.map(
          (camera, index) => html`<option value="${camera.deviceId}" ?selected=${camera.deviceId === this.selectedCamera}>
            ${camera.label || `Camera ${index + 1}`}
          </option>`
        )}
      </select>
    `;
  }

  _handleChange(e) {
    this.dispatchEvent(new CustomEvent('camera-change', {
      detail: e.target.value,
      bubbles: true,
      composed: true
    }));
  }
}

customElements.define('camera-selector', CameraSelector);
