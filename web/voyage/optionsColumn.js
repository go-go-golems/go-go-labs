class OptionsColumn {
    constructor(state, updateUI) {
        this.state = state;
        this.updateUI = updateUI;
        this.element = document.getElementById('options-column');
        this.init();
    }

    init() {
        const aspectOptions = this.element.querySelectorAll('input[name="aspect-ratio"]');
        aspectOptions.forEach(radio => {
            radio.addEventListener('change', (e) => this.handleAspectRatioChange(e));
        });

        const modelSelect = this.element.querySelector('#model-version-select');
        modelSelect.addEventListener('change', (e) => this.handleModelVersionChange(e));
    }

    render() {
        const options = this.state.get('options');
        
        // Update aspect ratio selection
        const aspectOptions = this.element.querySelectorAll('input[name="aspect-ratio"]');
        aspectOptions.forEach(radio => {
            radio.checked = (radio.value === options.aspect_ratio);
        });

        // If no matching aspect ratio is found, add a custom option
        if (!Array.from(aspectOptions).some(radio => radio.checked)) {
            this.addCustomAspectRatio(options.aspect_ratio);
        }

        // Update model version selection
        const modelSelect = this.element.querySelector('#model-version-select');
        if (modelSelect.querySelector(`option[value="${options.model_version}"]`)) {
            modelSelect.value = options.model_version;
        } else {
            this.addCustomModelVersion(options.model_version);
        }
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

    addCustomAspectRatio(ratio) {
        const container = this.element.querySelector('#aspect-ratio-options');
        const customRadio = document.createElement('label');
        customRadio.innerHTML = `<input type="radio" name="aspect-ratio" value="${ratio}" checked> ${ratio} (Custom)`;
        container.appendChild(customRadio);
    }

    addCustomModelVersion(version) {
        const modelSelect = this.element.querySelector('#model-version-select');
        const customOption = document.createElement('option');
        customOption.value = version;
        customOption.textContent = `${version} (Custom)`;
        customOption.selected = true;
        modelSelect.appendChild(customOption);
    }
}

export default OptionsColumn;