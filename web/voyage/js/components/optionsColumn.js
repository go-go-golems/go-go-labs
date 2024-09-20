import { html, render } from 'https://cdn.jsdelivr.net/gh/lit/dist@3/all/lit-all.min.js';

class OptionsColumn {
    constructor(state, updateUI) {
        this.state = state;
        this.updateUI = updateUI;
        this.element = document.getElementById('options-column');
        this.init();
    }

    init() {
        this.element.addEventListener('change', (e) => {
            if (e.target.name === 'aspect-ratio') this.handleAspectRatioChange(e);
            if (e.target.id === 'model-version-select') this.handleModelVersionChange(e);
        });
    }

    render() {
        const options = this.state.get('options');
        const template = html`
            <h2>Options</h2>
            <div>
                <h3>Aspect Ratio</h3>
                <div id="aspect-ratio-options">
                    ${this.renderAspectRatioOptions(options.aspect_ratio)}
                </div>
            </div>
            <div>
                <h3>Model Version</h3>
                <select id="model-version-select" .value=${options.model_version}>
                    ${this.renderModelVersionOptions(options.model_version)}
                </select>
            </div>
        `;

        render(template, this.element);
    }

    renderAspectRatioOptions(selectedRatio) {
        const standardRatios = ['1:1', '16:9', '4:3'];
        return html`
            ${standardRatios.map(ratio => html`
                <label>
                    <input type="radio" name="aspect-ratio" value=${ratio} ?checked=${ratio === selectedRatio}>
                    ${ratio}
                </label><br>
            `)}
            ${!standardRatios.includes(selectedRatio) ? html`
                <label>
                    <input type="radio" name="aspect-ratio" value=${selectedRatio} checked>
                    ${selectedRatio} (Custom)
                </label>
            ` : ''}
        `;
    }

    renderModelVersionOptions(selectedVersion) {
        const standardVersions = ['7', '6', '5', '4', '3', '2'];
        return html`
            ${standardVersions.map(version => html`
                <option value=${version} ?selected=${version === selectedVersion}>v${version}</option>
            `)}
            ${!standardVersions.includes(selectedVersion) ? html`
                <option value=${selectedVersion} selected>${selectedVersion} (Custom)</option>
            ` : ''}
        `;
    }

    handleAspectRatioChange(event) {
        const options = this.state.get('options');
        options.aspect_ratio = event.target.value;
        this.state.set('options', options);
        this.updateUI();
    }

    handleModelVersionChange(event) {
        const options = this.state.get('options');
        options.model_version = event.target.value;
        this.state.set('options', options);
        this.updateUI();
    }
}

export default OptionsColumn;