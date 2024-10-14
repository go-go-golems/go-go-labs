import { LitElement, html, css } from 'https://unpkg.com/lit@2.6.1/index.js?module';
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
      { width: 2592, height: 1944 },
      { width: 1280, height: 960 },
      { width: 1920, height: 1080 },
      { width: 1280, height: 720 },
      { width: 1024, height: 576 },
      { width: 800, height: 600 },
      { width: 640, height: 480 },
      { width: 320, height: 240 }
    ];
    const supportedResolutions = [];

    console.log(`Populating resolutions for device: ${deviceId}`);

    // Test common resolutions
    for (let res of commonResolutions) {
      const supported = await this.testResolution(deviceId, res.width, res.height);
      if (supported) {
        supportedResolutions.push(`${res.width}x${res.height}`);
        console.log(`Supported common resolution: ${res.width}x${res.height}`);
      } else {
        console.log(`Unsupported common resolution: ${res.width}x${res.height}`);
      }
    }

    // Test capabilities if available
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ video: { deviceId: { exact: deviceId } } });
      const track = stream.getVideoTracks()[0];
      
      if (track.getCapabilities) {
        const capabilities = track.getCapabilities();
        console.log('Device capabilities:', capabilities);

        if (capabilities.width && capabilities.height) {
          const widths = capabilities.width.max ? [capabilities.width.max, ...Array.from({ length: 5 }, (_, i) => capabilities.width.max - (i + 1) * 320).filter(w => w >= capabilities.width.min)] : [];
          const heights = capabilities.height.max ? [capabilities.height.max, ...Array.from({ length: 5 }, (_, i) => capabilities.height.max - (i + 1) * 240).filter(h => h >= capabilities.height.min)] : [];

          console.log('Testing resolutions from capabilities:');
          for (let width of widths) {
            for (let height of heights) {
              const supported = await this.testResolution(deviceId, width, height);
              if (supported) {
                const resolution = `${width}x${height}`;
                if (!supportedResolutions.includes(resolution)) {
                  supportedResolutions.push(resolution);
                  console.log(`Supported resolution from capabilities: ${resolution}`);
                }
              } else {
                console.log(`Unsupported resolution from capabilities: ${width}x${height}`);
              }
            }
          }
        }
      } else {
        console.log('getCapabilities function not available for this device/browser');
      }

      stream.getTracks().forEach(track => track.stop());
    } catch (error) {
      console.error('Error getting device capabilities:', error);
    }

    if (supportedResolutions.length === 0) {
      supportedResolutions.push('Default');
      console.log('No supported resolutions found, using Default');
    }

    this.resolutions = supportedResolutions;
    this.selectedResolution = this.resolutions[0];
    console.log('Final supported resolutions:', this.resolutions);
    console.log('Selected resolution:', this.selectedResolution);
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
