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
    `;

    static properties = {
        activeImage: { type: Object },
        points: { type: Array },
    };

    constructor() {
        super();
        this.activeImage = null;
        this.points = [];
    }

    render() {
        return html`
            <div id="canvas-container">
                <canvas id="imageCanvas" width="500" height="400"></canvas>
            </div>
        `;
    }

    firstUpdated() {
        this.canvas = this.renderRoot.querySelector('#imageCanvas');
        this.ctx = this.canvas.getContext('2d');
        this.canvas.addEventListener('click', this.handleCanvasClick.bind(this));
        this.draw();
    }

    updated(changedProperties) {
        if (changedProperties.has('activeImage') || changedProperties.has('points')) {
            this.draw();
        }
    }

    handleCanvasClick(event) {
        if (!this.activeImage) return;
        const rect = this.canvas.getBoundingClientRect();
        const x = event.clientX - rect.left;
        const y = event.clientY - rect.top;
        if (this.points.length < 4) {
            this.points = [...this.points, { x, y }];
            this.dispatchPointsUpdated();
        }
    }

    dispatchPointsUpdated() {
        this.dispatchEvent(new CustomEvent('points-updated', {
            detail: this.points,
            bubbles: true,
            composed: true
        }));
    }

    draw() {
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
        if (!this.activeImage) return;

        const scale = Math.min(this.canvas.width / this.activeImage.width, this.canvas.height / this.activeImage.height);
        const scaledWidth = this.activeImage.width * scale;
        const scaledHeight = this.activeImage.height * scale;
        const offsetX = (this.canvas.width - scaledWidth) / 2;
        const offsetY = (this.canvas.height - scaledHeight) / 2;

        this.ctx.drawImage(this.activeImage, offsetX, offsetY, scaledWidth, scaledHeight);

        if (this.points.length === 4) {
            this.ctx.beginPath();
            this.ctx.moveTo(this.points[0].x, this.points[0].y);
            this.points.forEach(point => this.ctx.lineTo(point.x, point.y));
            this.ctx.closePath();
            this.ctx.strokeStyle = 'yellow';
            this.ctx.stroke();
        }

        this.points.forEach(point => {
            this.ctx.fillStyle = 'red';
            this.ctx.beginPath();
            this.ctx.arc(point.x, point.y, 5, 0, 2 * Math.PI);
            this.ctx.fill();
        });
    }
}

customElements.define('image-canvas', ImageCanvas);