import { LitElement, html, css } from 'https://unpkg.com/lit@2.6.1/index.js?module';

class Controls extends LitElement {
    static styles = css`
        .controls {
            margin-top: 10px;
        }
        .controls button {
            margin-right: 5px;
            padding: 5px 10px;
            font-size: 14px;
        }
    `;

    static properties = {
        points: { type: Array },
        boxClosed: { type: Boolean },
    };

    render() {
        console.log('Controls: Rendering', {
            pointsCount: this.points.length,
            boxClosed: this.boxClosed
        });
        return html`
            <div class="controls">
                <button @click="${this.undo}" ?disabled="${this.points.length === 0}">Undo</button>
                <button @click="${this.clear}" ?disabled="${this.points.length === 0}">Clear</button>
                <button @click="${this.extract}" ?disabled="${!this.boxClosed}">Download</button>
            </div>
        `;
    }

    undo() {
        console.log('Controls: Undo clicked');
        this.dispatchEvent(new CustomEvent('undo', { bubbles: true, composed: true }));
    }

    clear() {
        console.log('Controls: Clear clicked');
        this.dispatchEvent(new CustomEvent('clear', { bubbles: true, composed: true }));
    }

    extract() {
        console.log('Controls: Extract clicked');
        this.dispatchEvent(new CustomEvent('extract', { bubbles: true, composed: true }));
    }
}

customElements.define('image-controls', Controls);
