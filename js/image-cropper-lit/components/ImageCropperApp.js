import { LitElement, html, css } from 'https://unpkg.com/lit@2.6.1/index.js?module';
import './ImageList.js';
import './ImageCanvas.js';
import './Controls.js';
import './Previews.js';
import { applyPerspectiveTransform } from '../utils/perspectiveTransform.js';
import { downloadImage } from '../utils/download.js';

class ImageCropperApp extends LitElement {
    static styles = css`
        #container {
            display: flex;
            height: 100vh;
            box-sizing: border-box;
            padding: 10px;
        }
        #main-section {
            flex-grow: 1;
            display: flex;
            flex-direction: column;
            margin-left: 20px;
        }
        #fileInputContainer {
            padding: 10px;
            border-top: 1px solid #ccc;
        }
    `;

    static properties = {
        images: { type: Array },
        activeImageIndex: { type: Number },
        points: { type: Array },
        boxClosed: { type: Boolean },
        showPreviews: { type: Boolean },
        autoAdvance: { type: Boolean },
        autoDownload: { type: Boolean },
        zoomLevel: { type: Number },
    };

    constructor() {
        super();
        this.images = [];
        this.activeImageIndex = -1;
        this.points = [];
        this.boxClosed = false;
        this.showPreviews = false;
        this.autoAdvance = false;
        this.autoDownload = false;
        this.zoomLevel = 1;
        this.handleKeyDown = this.handleKeyDown.bind(this);
        this.loadPreferences();
        console.log('ImageCropperApp: Initialized');
    }

    connectedCallback() {
        super.connectedCallback();
        window.addEventListener('keydown', this.handleKeyDown);
    }

    disconnectedCallback() {
        super.disconnectedCallback();
        window.removeEventListener('keydown', this.handleKeyDown);
    }

    render() {
        console.log('ImageCropperApp: Rendering', {
            imagesCount: this.images.length,
            activeImageIndex: this.activeImageIndex,
            pointsCount: this.points.length,
            boxClosed: this.boxClosed,
            showPreviews: this.showPreviews,
            autoAdvance: this.autoAdvance,
            autoDownload: this.autoDownload
        });
        return html`
            <div id="container">
                <image-list 
                    .images="${this.images}" 
                    .activeIndex="${this.activeImageIndex}" 
                    @image-selected="${this.onImageSelected}">
                </image-list>
                <div id="main-section">
                    <image-canvas 
                        .activeImage="${this.images[this.activeImageIndex]?.img}" 
                        .points="${this.points}"
                        .zoomLevel="${this.zoomLevel}"
                        @points-updated="${this.onPointsUpdated}"
                        @box-closed="${this.onBoxClosed}"
                        @zoom-updated="${this.onZoomUpdated}">
                    </image-canvas>
                    <image-controls 
                        .points="${this.points}"
                        .boxClosed="${this.boxClosed}"
                        .showPreviews="${this.showPreviews}"
                        .autoAdvance="${this.autoAdvance}"
                        .autoDownload="${this.autoDownload}"
                        @undo="${this.undoPoint}" 
                        @clear="${this.clearPoints}" 
                        @extract="${this.extractImage}"
                        @toggle-previews="${this.togglePreviews}"
                        @download-all="${this.downloadAll}"
                        @toggle-auto-advance="${this.toggleAutoAdvance}"
                        @toggle-auto-download="${this.toggleAutoDownload}">
                    </image-controls>
                    ${this.showPreviews ? html`
                        <image-previews 
                            .activeImage="${this.images[this.activeImageIndex]?.img}" 
                            .points="${this.points}"
                            .boxClosed="${this.boxClosed}">
                        </image-previews>
                    ` : ''}
                    <div id="fileInputContainer">
                        <input type="file" id="fileInput" multiple @change="${this.handleFileInput}">
                    </div>
                </div>
            </div>
        `;
    }

    handleKeyDown(event) {
        if (event.key === 'ArrowUp') {
            this.navigateImage(-1);
        } else if (event.key === 'ArrowDown') {
            this.navigateImage(1);
        }
    }

    navigateImage(direction) {
        const newIndex = this.activeImageIndex + direction;
        if (newIndex >= 0 && newIndex < this.images.length) {
            this.activeImageIndex = newIndex;
            this.points = this.images[this.activeImageIndex].points || [];
            this.boxClosed = this.points.length === 4;
            this.requestUpdate();
        }
    }

    togglePreviews() {
        this.showPreviews = !this.showPreviews;
        this.savePreferences();
    }

    async downloadAll() {
        for (let i = 0; i < this.images.length; i++) {
            const image = this.images[i];
            if (image.points && image.points.length === 4) {
                const transformedCanvas = applyPerspectiveTransform(image.img, image.points);
                await downloadImage(transformedCanvas, `cropped-image-${i + 1}.png`);
            }
        }
    }

    async handleFileInput(event) {
        console.log('ImageCropperApp: Handling file input', event.target.files);
        const files = Array.from(event.target.files);
        for (const file of files) {
            const img = await this.loadImage(file);
            this.images = [...this.images, { file, img, points: [] }];
            if (this.activeImageIndex === -1) {
                this.activeImageIndex = 0;
            }
        }
        console.log('ImageCropperApp: Files processed', this.images);
        this.requestUpdate();
    }

    loadImage(file) {
        console.log('ImageCropperApp: Loading image', file.name);
        return new Promise((resolve, reject) => {
            const img = new Image();
            img.onload = () => {
                console.log('ImageCropperApp: Image loaded', file.name);
                resolve(img);
            };
            img.onerror = reject;
            img.src = URL.createObjectURL(file);
        });
    }

    onBoxClosed() {
        console.log('ImageCropperApp: Box closed');
        this.boxClosed = true;
        
        // Save the current points to the active image
        if (this.activeImageIndex !== -1) {
            this.images[this.activeImageIndex].points = this.points;
        }
        
        if (this.autoDownload) {
            this.extractImage();
        }
        
        if (this.autoAdvance) {
            this.navigateImage(1);
        }
    }

    onImageSelected(e) {
        console.log('ImageCropperApp: Image selected', e.detail);
        this.activeImageIndex = e.detail;
        this.points = this.images[this.activeImageIndex].points || [];
        this.boxClosed = this.points.length === 4;
    }

    onPointsUpdated(e) {
        console.log('ImageCropperApp: Points updated', e.detail);
        this.points = e.detail;
        
        // Only update the image points if the box is not closed yet
        if (this.activeImageIndex !== -1 && !this.boxClosed) {
            this.images[this.activeImageIndex].points = this.points;
        }
        this.requestUpdate();
    }

    clearPoints() {
        console.log('ImageCropperApp: Clearing points');
        this.points = [];
        this.boxClosed = false;
        this.requestUpdate();
    }

    undoPoint() {
        console.log('ImageCropperApp: Undoing point');
        if (this.points.length > 0) {
            this.points = this.points.slice(0, -1);
            this.boxClosed = this.points.length === 4;
            this.requestUpdate();
        }
    }

    extractImage() {
        console.log('ImageCropperApp: Extracting image');
        if (this.points.length === 4 && this.activeImageIndex !== -1) {
            const activeImage = this.images[this.activeImageIndex].img;
            const transformedCanvas = applyPerspectiveTransform(activeImage, this.points);
            downloadImage(transformedCanvas, `cropped-image-${this.activeImageIndex + 1}.png`);
        }
    }

    toggleAutoAdvance() {
        this.autoAdvance = !this.autoAdvance;
        this.savePreferences();
    }

    toggleAutoDownload() {
        this.autoDownload = !this.autoDownload;
        this.savePreferences();
    }

    onZoomUpdated(e) {
        this.zoomLevel = e.detail;
        this.savePreferences();
    }

    loadPreferences() {
        const preferences = JSON.parse(localStorage.getItem('imageCropperPreferences')) || {};
        this.showPreviews = preferences.showPreviews || false;
        this.autoAdvance = preferences.autoAdvance || false;
        this.autoDownload = preferences.autoDownload || false;
        this.zoomLevel = preferences.zoomLevel || 1;
    }

    savePreferences() {
        const preferences = {
            showPreviews: this.showPreviews,
            autoAdvance: this.autoAdvance,
            autoDownload: this.autoDownload,
            zoomLevel: this.zoomLevel,
        };
        localStorage.setItem('imageCropperPreferences', JSON.stringify(preferences));
    }
}

customElements.define('image-cropper-app', ImageCropperApp);
