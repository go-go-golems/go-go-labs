
Let's start by creating multiple components:

1. `CameraStream`: The main component that orchestrates the entire functionality.
2. `CameraSelector`: A component for camera selection.
3. `ResolutionSelector`: A component for resolution selection.
4. `CaptureControls`: A component for frame count and countdown speed controls.
5. `ProgressBar`: A reusable component for both progress and countdown visualization.

Here's the implementation:

1. First, let's create the main `CameraStream` component:

```javascript
// camera-stream.js
import { LitElement, html, css } from 'lit';
import './camera-selector.js';
import './resolution-selector.js';
import './capture-controls.js';
import './progress-bar.js';

export class CameraStream extends LitElement {
  static properties = {
    cameras: { type: Array },
    selectedCamera: { type: String },
    resolutions: { type: Array },
    selectedResolution: { type: String },
    frameCount: { type: Number },
    countdownSpeed: { type: Number },
    isCapturing: { type: Boolean },
    progress: { type: Number },
    countdown: { type: Number },
    captureRunCount: { type: Number },
    stream: { type: Object },
  };

  static styles = css`
    :host {
      display: block;
      font-family: Arial, sans-serif;
      text-align: center;
    }
    #videoElement {
      max-width: 100%;
      background-color: #000;
    }
    #controls {
      margin: 20px;
      display: flex;
      flex-wrap: wrap;
      justify-content: center;
      gap: 10px;
    }
    #startCaptureButton {
      padding: 10px 20px;
      font-size: 16px;
      margin-top: 10px;
    }
  `;

  constructor() {
    super();
    this.cameras = [];
    this.selectedCamera = '';
    this.resolutions = [];
    this.selectedResolution = '';
    this.frameCount = 5;
    this.countdownSpeed = 1;
    this.isCapturing = false;
    this.progress = 0;
    this.countdown = 3;
    this.captureRunCount = 0;
    this.stream = null;
  }

  connectedCallback() {
    super.connectedCallback();
    this.getConnectedCameras();
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    this.stopStream();
  }

  async getConnectedCameras() {
    try {
      await navigator.mediaDevices.getUserMedia({ video: true });
      const devices = await navigator.mediaDevices.enumerateDevices();
      this.cameras = devices.filter(device => device.kind === 'videoinput');
      if (this.cameras.length > 0) {
        this.selectedCamera = this.cameras[0].deviceId;
        await this.handleCameraChange({ detail: this.selectedCamera });
      }
    } catch (error) {
      console.error('Error accessing cameras: ', error);
    }
  }

  async handleCameraChange(e) {
    this.selectedCamera = e.detail;
    await this.populateResolutions(this.selectedCamera);
    await this.startStream(this.selectedCamera, this.selectedResolution);
  }

  async populateResolutions(deviceId) {
    const commonResolutions = [
      { width: 1920, height: 1080 },
      { width: 1280, height: 720 },
      { width: 1024, height: 576 },
      { width: 800, height: 600 },
      { width: 640, height: 480 },
      { width: 320, height: 240 }
    ];
    const supportedResolutions = [];

    for (let res of commonResolutions) {
      const supported = await this.testResolution(deviceId, res.width, res.height);
      if (supported) {
        supportedResolutions.push(`${res.width}x${res.height}`);
      }
    }

    if (supportedResolutions.length === 0) {
      supportedResolutions.push('Default');
    }

    this.resolutions = supportedResolutions;
    this.selectedResolution = this.resolutions[0];
  }

  async testResolution(deviceId, width, height) {
    const constraints = {
      video: {
        deviceId: { exact: deviceId },
        width: { exact: width },
        height: { exact: height }
      }
    };

    try {
      const testStream = await navigator.mediaDevices.getUserMedia(constraints);
      testStream.getTracks().forEach(track => track.stop());
      return true;
    } catch (error) {
      return false;
    }
  }

  async startStream(deviceId, resolution) {
    this.stopStream();

    let constraints = {
      video: { deviceId: { exact: deviceId } }
    };

    if (resolution && resolution !== 'Default') {
      const [width, height] = resolution.split('x').map(Number);
      constraints.video.width = { exact: width };
      constraints.video.height = { exact: height };
    }

    try {
      this.stream = await navigator.mediaDevices.getUserMedia(constraints);
      this.videoElement.srcObject = this.stream;
    } catch (error) {
      console.error('Error starting video stream: ', error);
    }
  }

  stopStream() {
    if (this.stream) {
      this.stream.getTracks().forEach(track => track.stop());
      this.stream = null;
    }
  }

  handleResolutionChange(e) {
    this.selectedResolution = e.detail;
    this.startStream(this.selectedCamera, this.selectedResolution);
  }

  handleFrameCountChange(e) {
    this.frameCount = e.detail;
  }

  handleCountdownSpeedChange(e) {
    this.countdownSpeed = e.detail;
  }

  async startCapture() {
    this.isCapturing = true;
    this.captureRunCount += 1;
    this.progress = 0;
    this.countdown = 3;

    await this.runCountdown(this.countdown);

    for (let i = 1; i <= this.frameCount; i++) {
      await this.runCountdown(this.countdownSpeed);
      this.captureFrame(i, this.captureRunCount);
      this.progress = (i / this.frameCount) * 100;
    }

    this.isCapturing = false;
    this.progress = 0;
    this.countdown = 3;
  }

  runCountdown(seconds) {
    return new Promise(resolve => {
      let remaining = seconds;
      const interval = setInterval(() => {
        if (remaining > 0) {
          this.countdown = remaining;
          remaining -= 1;
        } else {
          clearInterval(interval);
          this.countdown = 0;
          resolve();
        }
      }, 1000);
    });
  }

  captureFrame(frameNumber, runNumber) {
    const canvas = document.createElement('canvas');
    canvas.width = this.videoElement.videoWidth;
    canvas.height = this.videoElement.videoHeight;
    const context = canvas.getContext('2d');
    context.drawImage(this.videoElement, 0, 0, canvas.width, canvas.height);
    canvas.toBlob(blob => {
      const link = document.createElement('a');
      link.href = URL.createObjectURL(blob);
      const paddedFrameNumber = String(frameNumber).padStart(3, '0');
      link.download = `frame_${runNumber}_${paddedFrameNumber}.png`;
      link.click();
    }, 'image/png');
  }

  render() {
    return html`
      <h1>Camera Stream with Resolutions</h1>
      <div id="controls">
        <camera-selector
          .cameras=${this.cameras}
          .selectedCamera=${this.selectedCamera}
          @camera-change=${this.handleCameraChange}
        ></camera-selector>
        <resolution-selector
          .resolutions=${this.resolutions}
          .selectedResolution=${this.selectedResolution}
          @resolution-change=${this.handleResolutionChange}
        ></resolution-selector>
        <capture-controls
          .frameCount=${this.frameCount}
          .countdownSpeed=${this.countdownSpeed}
          @frame-count-change=${this.handleFrameCountChange}
          @countdown-speed-change=${this.handleCountdownSpeedChange}
        ></capture-controls>
        <button id="startCaptureButton" @click=${this.startCapture} ?disabled=${this.isCapturing}>
          Start Capture
        </button>
      </div>
      ${this.isCapturing ? html`
        <progress-bar
          type="countdown"
          .progress=${(this.countdown / (this.countdown + this.countdownSpeed)) * 100}
          .label=${this.countdown > 0 ? this.countdown : 'Capturing...'}
        ></progress-bar>
        <progress-bar
          type="progress"
          .progress=${this.progress}
          .label=${Math.round(this.progress) + '%'}
        ></progress-bar>
      ` : ''}
      <video id="videoElement" autoplay playsinline></video>
    `;
  }

  get videoElement() {
    return this.renderRoot.querySelector('#videoElement');
  }
}

customElements.define('camera-stream', CameraStream);
```

2. Now, let's create the `CameraSelector` component:

```javascript
// camera-selector.js
import { LitElement, html, css } from 'lit';

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
```

3. Next, the `ResolutionSelector` component:

```javascript
// resolution-selector.js
import { LitElement, html, css } from 'lit';

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
```

4. The `CaptureControls` component:

```javascript
// capture-controls.js
import { LitElement, html, css } from 'lit';

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
```

5. Finally, the `ProgressBar` component:

```javascript
// progress-bar.js
import { LitElement, html, css } from 'lit';

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
```

Now, let's update the `index.html` file to use our new component:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Camera Stream with Lit.js</title>
    <script type="module" src="./camera-stream.js"></script>
</head>
<body>
    <camera-stream></camera-stream>
</body>
</html>
```

This implementation breaks down the functionality into smaller, more manageable components. Each component is responsible for a specific part