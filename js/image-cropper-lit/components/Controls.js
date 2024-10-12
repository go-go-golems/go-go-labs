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
    };

    render() {
        return html`
            <div class="controls">
                <button @click="${this.undo}" ?disabled="${this.points.length === 0}">Undo</button>
                <button @click="${this.clear}" ?disabled="${this.points.length === 0}">Clear</button>
                <button @click="${this.extract}" ?disabled="${this.points.length !== 4}">Extract and Download</button>
            </div>
        `;
    }

    undo() {
        this.dispatchEvent(new CustomEvent('undo', { bubbles: true, composed: true }));
    }

    clear() {
        this.dispatchEvent(new CustomEvent('clear', { bubbles: true, composed: true }));
    }

    extract() {
        this.dispatchEvent(new CustomEvent('extract', { bubbles: true, composed: true }));
    }
}

customElements.define('image-controls', Controls);
