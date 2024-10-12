import { LitElement, html, css } from 'https://unpkg.com/lit@2.6.1/index.js?module';
import './ImageList.js';
import './ImageCanvas.js';
import './Controls.js';
import './Previews.js';

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
    };

    constructor() {
        super();
        this.images = [];
        this.activeImageIndex = -1;
        this.points = [];
    }

    render() {
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
                        @points-updated="${this.onPointsUpdated}">
                    </image-canvas>
                    <controls 
                        .points="${this.points}" 
                        @undo="${this.undoPoint}" 
                        @clear="${this.clearPoints}" 
                        @extract="${this.extractImage}">
                    </controls>
                    <previews 
                        .activeImage="${this.images[this.activeImageIndex]?.img}" 
                        .points="${this.points}">
                    </previews>
                    <div id="fileInputContainer">
                        <input type="file" id="fileInput" multiple @change="${this.handleFileInput}">
                    </div>
                </div>
            </div>
        `;
    }

    async handleFileInput(event) {
        const files = Array.from(event.target.files);
        for (const file of files) {
            const img = await this.loadImage(file);
            this.images = [...this.images, { file, img, points: [] }];
            if (this.activeImageIndex === -1) {
                this.activeImageIndex = 0;
            }
        }
        this.requestUpdate();
    }

    loadImage(file) {
        return new Promise((resolve, reject) => {
            const img = new Image();
            img.src = URL.createObjectURL(file);
            img.onload = () => resolve(img);
            img.onerror = reject;
        });
    }

    onImageSelected(e) {
        this.activeImageIndex = e.detail;
        this.points = this.images[this.activeImageIndex].points || [];
    }

    onPointsUpdated(e) {
        this.points = e.detail;
        if (this.activeImageIndex !== -1) {
            this.images[this.activeImageIndex].points = this.points;
            this.requestUpdate();
        }
    }

    undoPoint() {
        if (this.points.length > 0) {
            this.points = this.points.slice(0, -1);
            this.requestUpdate();
        }
    }

    clearPoints() {
        this.points = [];
        this.requestUpdate();
    }

    async extractImage() {
        if (this.points.length === 4 && this.activeImageIndex !== -1) {
            const { applyPerspectiveTransform } = await import('../utils/perspectiveTransform.js');
            const { downloadImage } = await import('../utils/download.js');
            const activeImage = this.images[this.activeImageIndex].img;
            const transformedCanvas = applyPerspectiveTransform(activeImage, this.points);
            downloadImage(transformedCanvas);
        }
    }
}

customElements.define('image-cropper-app', ImageCropperApp);