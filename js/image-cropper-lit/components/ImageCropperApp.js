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
        boxClosed: { type: Boolean },
    };

    constructor() {
        super();
        this.images = [];
        this.activeImageIndex = -1;
        this.points = [];
        this.boxClosed = false;
        console.log('ImageCropperApp: Initialized');
    }

    render() {
        console.log('ImageCropperApp: Rendering', {
            imagesCount: this.images.length,
            activeImageIndex: this.activeImageIndex,
            pointsCount: this.points.length,
            boxClosed: this.boxClosed
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
                        @points-updated="${this.onPointsUpdated}"
                        @box-closed="${this.onBoxClosed}">
                    </image-canvas>
                    <controls 
                        .points="${this.points}"
                        .boxClosed="${this.boxClosed}"
                        @undo="${this.undoPoint}" 
                        @clear="${this.clearPoints}" 
                        @extract="${this.extractImage}">
                    </controls>
                    <previews 
                        .activeImage="${this.images[this.activeImageIndex]?.img}" 
                        .points="${this.points}"
                        .boxClosed="${this.boxClosed}">
                    </previews>
                    <div id="fileInputContainer">
                        <input type="file" id="fileInput" multiple @change="${this.handleFileInput}">
                    </div>
                </div>
            </div>
        `;
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
            img.src = URL.createObjectURL(file);
            img.onload = () => resolve(img);
            img.onerror = reject;
        });
    }

    onBoxClosed() {
        console.log('ImageCropperApp: Box closed');
        this.boxClosed = true;
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
        if (this.activeImageIndex !== -1) {
            this.images[this.activeImageIndex].points = this.points;
            this.boxClosed = this.points.length === 4;
            this.requestUpdate();
        }
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
            this.requestUpdate();
        }
    }

    async extractImage() {
        console.log('ImageCropperApp: Extracting image');
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
