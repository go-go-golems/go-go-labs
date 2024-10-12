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
        .toggle-container {
            margin-top: 10px;
        }
        .toggle-container label {
            margin-right: 10px;
        }
    `;

    static properties = {
        points: { type: Array },
        boxClosed: { type: Boolean },
        showPreviews: { type: Boolean },
        autoAdvance: { type: Boolean },
        autoDownload: { type: Boolean },
    };

    render() {
        console.log('Controls: Rendering', {
            pointsCount: this.points.length,
            boxClosed: this.boxClosed,
            showPreviews: this.showPreviews,
            autoAdvance: this.autoAdvance,
            autoDownload: this.autoDownload
        });
        return html`
            <div class="controls">
                <button @click="${this.undo}" ?disabled="${this.points.length === 0}">Undo</button>
                <button @click="${this.clear}" ?disabled="${this.points.length === 0}">Clear</button>
                <button @click="${this.extract}" ?disabled="${!this.boxClosed}">Download</button>
                <button @click="${this.togglePreviews}">${this.showPreviews ? 'Hide' : 'Show'} Previews</button>
                <button @click="${this.downloadAll}">Download All</button>
            </div>
            <div class="toggle-container">
                <label>
                    <input type="checkbox" ?checked="${this.autoAdvance}" @change="${this.toggleAutoAdvance}">
                    Auto-advance
                </label>
                <label>
                    <input type="checkbox" ?checked="${this.autoDownload}" @change="${this.toggleAutoDownload}">
                    Auto-download
                </label>
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

    togglePreviews() {
        console.log('Controls: Toggle previews clicked');
        this.dispatchEvent(new CustomEvent('toggle-previews', { bubbles: true, composed: true }));
    }

    downloadAll() {
        console.log('Controls: Download all clicked');
        this.dispatchEvent(new CustomEvent('download-all', { bubbles: true, composed: true }));
    }

    toggleAutoAdvance() {
        console.log('Controls: Toggle auto-advance clicked');
        this.dispatchEvent(new CustomEvent('toggle-auto-advance', { bubbles: true, composed: true }));
    }

    toggleAutoDownload() {
        console.log('Controls: Toggle auto-download clicked');
        this.dispatchEvent(new CustomEvent('toggle-auto-download', { bubbles: true, composed: true }));
    }
}

customElements.define('image-controls', Controls);
