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
        const aspectOptions = this.element.querySelectorAll('input[name="aspect-ratio"]');
        aspectOptions.forEach(radio => {
            radio.checked = (radio.value === options.aspect_ratio);
        });

        const modelSelect = this.element.querySelector('#model-version-select');
        modelSelect.value = options.model_version;
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