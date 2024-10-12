import { LitElement, html, css } from 'https://unpkg.com/lit@2.6.1/index.js?module';

class ImageCanvas extends LitElement {
    static styles = css`
        #canvas-container {
            position: relative;
            margin-left: 20px;
            flex-grow: 1;
            display: flex;
            flex-direction: column;
            align-items: center;
        }
        canvas {
            border: 1px solid black;
            cursor: crosshair;
        }
        #info {
            margin-top: 10px;
            font-family: Arial, sans-serif;
        }
    `;

    static properties = {
        activeImage: { type: Object },
        points: { type: Array },
    };

    constructor() {
        super();
        this.activeImage = null;
        this.points = [];
        this.scale = 1;
        this.offsetX = 0;
        this.offsetY = 0;
        console.log('ImageCanvas: Initialized');
    }

    render() {
        console.log('ImageCanvas: Rendering', {
            activeImage: this.activeImage ? 'present' : 'null',
            pointsCount: this.points.length
        });
        return html`
            <div id="canvas-container">
                <canvas id="imageCanvas" width="500" height="400"></canvas>
                <div id="info">
                    ${this.activeImage ? `Image dimensions: ${this.activeImage.width} x ${this.activeImage.height}` : ''}
                    ${this.points.map((point, index) => html`
                        <div>Point ${index + 1}: (${Math.round(point.x)}, ${Math.round(point.y)})</div>
                    `)}
                </div>
            </div>
        `;
    }

    firstUpdated() {
        console.log('ImageCanvas: First updated');
        this.canvas = this.renderRoot.querySelector('#imageCanvas');
        this.ctx = this.canvas.getContext('2d');
        this.canvas.addEventListener('click', this.handleCanvasClick.bind(this));
        this.draw();
    }

    updated(changedProperties) {
        console.log('ImageCanvas: Updated', changedProperties);
        if (changedProperties.has('activeImage') || changedProperties.has('points')) {
            this.draw();
        }
    }

    handleCanvasClick(event) {
        console.log('ImageCanvas: Canvas clicked', event);
        if (!this.activeImage) return;
        const rect = this.canvas.getBoundingClientRect();
        const canvasX = event.clientX - rect.left;
        const canvasY = event.clientY - rect.top;
        
        const imagePoint = this.canvasToImageCoordinates(canvasX, canvasY);
        
        if (this.points.length < 4) {
            this.points = [...this.points, imagePoint];
            this.dispatchPointsUpdated();
            
            if (this.points.length === 4) {
                this.dispatchEvent(new CustomEvent('box-closed', {
                    bubbles: true,
                    composed: true
                }));
            }
        }
    }

    canvasToImageCoordinates(canvasX, canvasY) {
        const imageX = (canvasX - this.offsetX) / this.scale;
        const imageY = (canvasY - this.offsetY) / this.scale;
        console.log('ImageCanvas: Canvas to Image coordinates', { canvasX, canvasY, imageX, imageY });
        return { x: imageX, y: imageY };
    }

    imageToCanvasCoordinates(imageX, imageY) {
        const canvasX = imageX * this.scale + this.offsetX;
        const canvasY = imageY * this.scale + this.offsetY;
        console.log('ImageCanvas: Image to Canvas coordinates', { imageX, imageY, canvasX, canvasY });
        return { x: canvasX, y: canvasY };
    }

    dispatchPointsUpdated() {
        console.log('ImageCanvas: Dispatching points updated', this.points);
        this.dispatchEvent(new CustomEvent('points-updated', {
            detail: this.points,
            bubbles: true,
            composed: true
        }));
    }

    draw() {
        console.log('ImageCanvas: Drawing', {
            activeImage: this.activeImage ? 'present' : 'null',
            pointsCount: this.points.length
        });
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
        if (!this.activeImage) return;

        this.scale = Math.min(this.canvas.width / this.activeImage.width, this.canvas.height / this.activeImage.height);
        const scaledWidth = this.activeImage.width * this.scale;
        const scaledHeight = this.activeImage.height * this.scale;
        this.offsetX = (this.canvas.width - scaledWidth) / 2;
        this.offsetY = (this.canvas.height - scaledHeight) / 2;

        this.ctx.drawImage(this.activeImage, this.offsetX, this.offsetY, scaledWidth, scaledHeight);

        if (this.points.length === 4) {
            this.ctx.beginPath();
            const startPoint = this.imageToCanvasCoordinates(this.points[0].x, this.points[0].y);
            this.ctx.moveTo(startPoint.x, startPoint.y);
            this.points.forEach(point => {
                const canvasPoint = this.imageToCanvasCoordinates(point.x, point.y);
                this.ctx.lineTo(canvasPoint.x, canvasPoint.y);
            });
            this.ctx.closePath();
            this.ctx.strokeStyle = 'yellow';
            this.ctx.lineWidth = 2;
            this.ctx.stroke();
        }

        this.points.forEach((point, index) => {
            const canvasPoint = this.imageToCanvasCoordinates(point.x, point.y);
            this.ctx.beginPath();
            this.ctx.arc(canvasPoint.x, canvasPoint.y, 5, 0, 2 * Math.PI);
            this.ctx.fillStyle = 'red';
            this.ctx.fill();
            this.ctx.fillStyle = 'white';
            this.ctx.fillText((index + 1).toString(), canvasPoint.x - 3, canvasPoint.y + 3);
        });
    }
}

customElements.define('image-canvas', ImageCanvas);
