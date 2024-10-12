import { LitElement, html, css } from 'https://unpkg.com/lit@2.6.1/index.js?module';
import { applyPerspectiveTransform } from '../utils/perspectiveTransform.js';

class Previews extends LitElement {
    static styles = css`
        #previews {
            display: flex;
            margin-top: 20px;
            gap: 20px;
            justify-content: center;
        }
        #previews canvas {
            border: 1px solid #ccc;
        }
    `;

    static properties = {
        activeImage: { type: Object },
        points: { type: Array },
        boxClosed: { type: Boolean },
    };

    render() {
        console.log('Previews: Rendering', {
            activeImage: this.activeImage ? 'present' : 'null',
            pointsCount: this.points.length,
            boxClosed: this.boxClosed
        });
        if (!this.boxClosed) {
            return html``;
        }

        return html`
            <div id="previews">
                <canvas id="selectedPreview" width="200" height="200"></canvas>
                <canvas id="transformedPreview" width="200" height="200"></canvas>
            </div>
        `;
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
        if (!this.activeImage || this.points.length !== 4) {
          this.clearCanvas(this.selectedPreview);
          this.clearCanvas(this.transformedPreview);
          return;
        }
      
        const selectedCtx = this.selectedPreview.getContext('2d');
        this.clearCanvas(this.selectedPreview);
        
        const minX = Math.min(...this.points.map(p => p.x));
        const minY = Math.min(...this.points.map(p => p.y));
        const maxX = Math.max(...this.points.map(p => p.x));
        const maxY = Math.max(...this.points.map(p => p.y));
        const width = maxX - minX;
        const height = maxY - minY;
        
        selectedCtx.drawImage(this.activeImage, minX, minY, width, height, 0, 0, this.selectedPreview.width, this.selectedPreview.height);
      
        const transformedCanvas = applyPerspectiveTransform(this.activeImage, this.points);
        const transformedCtx = this.transformedPreview.getContext('2d');
        this.clearCanvas(this.transformedPreview);
        transformedCtx.drawImage(transformedCanvas, 0, 0, this.transformedPreview.width, this.transformedPreview.height);
      }
    clearCanvas(canvas) {
        console.log('Previews: Clearing canvas', canvas.id);
        const ctx = canvas.getContext('2d');
        ctx.clearRect(0, 0, canvas.width, canvas.height);
    }
}

customElements.define('image-previews', Previews);
