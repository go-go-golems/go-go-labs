import { LitElement, html, css } from 'https://unpkg.com/lit@2.6.1/index.js?module';

export class ProgressBar extends LitElement {
  static properties = {
    type: { type: String },
    progress: { type: Number },
    label: { type: String },
  };

  static styles = css`
    :host {
      display: block;
      width: 100%;
      margin-bottom: 10px;
    }
    .bar {
      width: 100%;
      height: 30px;
      background-color: #ddd;
    }
    .fill {
      height: 100%;
      text-align: center;
      line-height: 30px;
      color: white;
      transition: width 0.1s linear;
    }
    .countdown {
      background-color: #FFA500;
    }
    .progress {
      background-color: #4CAF50;
    }
  `;

  render() {
    return html`
      <div class="bar">
        <div class="fill ${this.type}"
             style="width: ${this.progress}%;">
          ${this.label}
        </div>
      </div>
    `;
  }
}

customElements.define('progress-bar', ProgressBar);
