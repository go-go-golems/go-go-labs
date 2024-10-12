import { LitElement, html, css } from 'https://unpkg.com/lit@2.6.1/index.js?module';

import { applyPerspectiveTransform } from '../utils/perspectiveTransform.js';

class Previews extends LitElement {
    static properties = {
        activeImage: { type: Object },
        points: { type: Array },
        boxClosed: { type: Boolean },
    };

    static styles = css`
        #previews {
            display: flex;
            justify-content: space-around;
        }
        .preview-container {
            text-align: center;
        }
        .info-container {
            margin-top: 10px;
            text-align: left;
            font-size: 12px;
        }
        canvas {
            max-width: 100%;
            height: auto;
        }
    `;

    render() {
        const transformInfo = this.getTransformInfo();

        return html`
            <div id="previews">
                <div class="preview-container">
                    <h3>Selected Area</h3>
                    <canvas id="selectedPreview" width="400" height="400"></canvas>
                </div>
                <div class="preview-container">
                    <h3>Transformed Image</h3>
                    <canvas id="transformedPreview" width="400" height="400"></canvas>
                </div>
            </div>
            ${transformInfo ? this.renderTransformInfo(transformInfo) : ''}
        `;
    }

    renderTransformInfo(info) {
        return html`
            <div class="info-container">
                <h4>Transform Information:</h4>
                <p>Points:</p>
                <ul>
                    ${info.points.map((point, index) => html`
                        <li>Point ${index + 1}: (${point.x.toFixed(2)}, ${point.y.toFixed(2)})</li>
                    `)}
                </ul>
                <p>Side lengths:</p>
                <ul>
                    ${info.sideLengths.map((length, index) => html`
                        <li>Side ${index + 1}: ${length.toFixed(2)}px</li>
                    `)}
                </ul>
                <p>Source points: ${JSON.stringify(info.sourcePoints)}</p>
                <p>Target points: ${JSON.stringify(info.targetPoints)}</p>
                <p>Target dimensions: ${info.targetWidth.toFixed(2)}x${info.targetHeight.toFixed(2)}</p>
            </div>
        `;
    }

    getTransformInfo() {
        if (!this.activeImage || this.points.length !== 4 || !this.boxClosed) {
            return null;
        }

        const sideLengths = this.calculateSideLengths();
        const targetWidth = Math.max(sideLengths[0], sideLengths[2]);
        const targetHeight = Math.max(sideLengths[1], sideLengths[3]);
        const sourcePoints = this.points.map(p => [p.x, p.y]);
        const targetPoints = [
            [0, 0],
            [targetWidth, 0],
            [targetWidth, targetHeight],
            [0, targetHeight]
        ];

        return {
            points: this.points,
            sideLengths,
            sourcePoints,
            targetPoints,
            targetWidth,
            targetHeight
        };
    }

    calculateSideLengths() {
        const lengths = [];
        for (let i = 0; i < 4; i++) {
            const p1 = this.points[i];
            const p2 = this.points[(i + 1) % 4];
            const dx = p2.x - p1.x;
            const dy = p2.y - p1.y;
            lengths.push(Math.sqrt(dx * dx + dy * dy));
        }
        return lengths;
    }

    firstUpdated() {
        console.log('Previews: First updated');
        this.selectedPreview = this.renderRoot.querySelector('#selectedPreview');
        this.transformedPreview = this.renderRoot.querySelector('#transformedPreview');
        this.updatePreviews();
    }

    updated(changedProperties) {
        console.log('Previews: Updated', changedProperties);
        if (changedProperties.has('activeImage') || changedProperties.has('points') || changedProperties.has('boxClosed')) {
            this.updatePreviews();
        }
    }

    updatePreviews() {
        console.log('Previews: Updating previews');
        console.log('Active Image:', this.activeImage ? 'Present' : 'Not present');
        console.log('Box Closed:', this.boxClosed);

        if (!this.activeImage || this.points.length !== 4 || !this.boxClosed) {
            console.log('Previews: Conditions not met, clearing canvases');
            this.clearCanvas(this.selectedPreview);
            this.clearCanvas(this.transformedPreview);
            return;
        }

        // Log detailed information about points and side lengths
        console.log('Points:');
        this.points.forEach((point, index) => {
            console.log(`  Point ${index + 1}: (${point.x}, ${point.y})`);
        });

        const sideLengths = this.calculateSideLengths();
        console.log('Side lengths:');
        sideLengths.forEach((length, index) => {
            console.log(`  Side ${index + 1}: ${length.toFixed(2)}px`);
        });

        console.log('Previews: Drawing selected area');
        const selectedCtx = this.selectedPreview.getContext('2d');
        this.clearCanvas(this.selectedPreview);

        // Draw the selected area
        const path = new Path2D();
        path.moveTo(this.points[0].x, this.points[0].y);
        for (let i = 1; i < this.points.length; i++) {
            path.lineTo(this.points[i].x, this.points[i].y);
        }
        path.closePath();

        selectedCtx.save();
        selectedCtx.clip(path);
        selectedCtx.drawImage(this.activeImage, 0, 0);
        selectedCtx.restore();

        console.log('Previews: Applying perspective transform');
        // Log information about the perspective transform
        const sourcePoints = this.points.map(p => [p.x, p.y]);
        const targetWidth = Math.max(sideLengths[0], sideLengths[2]);
        const targetHeight = Math.max(sideLengths[1], sideLengths[3]);
        const targetPoints = [
            [0, 0],
            [targetWidth, 0],
            [targetWidth, targetHeight],
            [0, targetHeight]
        ];
        console.log('Perspective transform:');
        console.log('  Source points:', sourcePoints);
        console.log('  Target points:', targetPoints);
        console.log(`  Target dimensions: ${targetWidth.toFixed(2)}x${targetHeight.toFixed(2)}`);

        // Draw the transformed image
        const transformedCanvas = applyPerspectiveTransform(this.activeImage, this.points);
        const transformedCtx = this.transformedPreview.getContext('2d');
        this.clearCanvas(this.transformedPreview);
        transformedCtx.drawImage(transformedCanvas, 0, 0, this.transformedPreview.width, this.transformedPreview.height);

        console.log('Previews: Update complete');
    }

    clearCanvas(canvas) {
        if (canvas) {
            console.log('Previews: Clearing canvas', canvas.id);
            const ctx = canvas.getContext('2d');
            ctx.clearRect(0, 0, canvas.width, canvas.height);
        }
    }
}

customElements.define('image-previews', Previews);
