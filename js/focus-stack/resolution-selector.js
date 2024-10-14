import { LitElement, html, css } from 'https://unpkg.com/lit@2.6.1/index.js?module';

export class ResolutionSelector extends LitElement {
  static properties = {
    resolutions: { type: Array },
    selectedResolution: { type: String },
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
      <label for="resolutionSelect">Select Resolution: </label>
      <select id="resolutionSelect" @change=${this._handleChange}>
        ${this.resolutions.map(
          res => html`<option value="${res}" ?selected=${res === this.selectedResolution}>${res}</option>`
        )}
      </select>
    `;
  }

  _handleChange(e) {
    this.dispatchEvent(new CustomEvent('resolution-change', {
      detail: e.target.value,
      bubbles: true,
      composed: true
    }));
  }
}

customElements.define('resolution-selector', ResolutionSelector);
