import State from './state.js';
import FragmentsColumn from './components/fragmentsColumn.js';
import PromptColumn from './components/promptColumn.js';
import OptionsColumn from './components/optionsColumn.js';
import HistorySection from './components/historySection.js';
import { showConfirmation } from './utils.js';

class App {
    constructor() {
        this.state = new State();
        this.fragmentsColumn = new FragmentsColumn(this.state, () => this.updateUI());
        this.promptColumn = new PromptColumn(this.state, () => this.updateUI());
        this.optionsColumn = new OptionsColumn(this.state, () => this.updateUI());
        this.historySection = new HistorySection(this.state, () => this.updateUI());

        this.initModal();
        this.initImportExport();
        this.initSaveSelectionModal();
    }

    initModal() {
        document.getElementById('confirm-add-image-btn').addEventListener('click', () => this.addImageURL());
        document.getElementById('cancel-add-image-btn').addEventListener('click', () => this.closeModal());
    }

    initImportExport() {
        document.getElementById('export-btn').addEventListener('click', () => this.exportState());
        document.getElementById('import-btn').addEventListener('click', () => this.importState());
    }

    initSaveSelectionModal() {
        document.getElementById('confirm-save-selection-btn').addEventListener('click', () => this.fragmentsColumn.saveFragmentSelection());
        document.getElementById('cancel-save-selection-btn').addEventListener('click', () => this.fragmentsColumn.closeSaveSelectionModal());
    }

    updateUI() {
        this.fragmentsColumn.render();
        this.promptColumn.render();
        this.optionsColumn.render();
        this.historySection.render();
        this.state.save();
    }

    addImageURL() {
        const url = document.getElementById('new-image-url').value.trim();
        if (url) {
            const images = this.state.get('images');
            images.unshift({ url, thumbnail: "", alt: "New image" });
            this.state.set('images', images);
            this.updateUI();
            this.closeModal();
            showConfirmation("Image added successfully!");
        }
    }

    closeModal() {
        document.getElementById('image-modal').style.display = 'none';
    }

    exportState() {
        const dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(this.state.data));
        const downloadAnchorNode = document.createElement('a');
        downloadAnchorNode.setAttribute("href", dataStr);
        downloadAnchorNode.setAttribute("download", "midjourney_prompt_state.json");
        document.body.appendChild(downloadAnchorNode);
        downloadAnchorNode.click();
        downloadAnchorNode.remove();
        showConfirmation("State exported successfully!");
    }

    importState() {
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = 'application/json';
        input.onchange = e => {
            const file = e.target.files[0];
            const reader = new FileReader();
            reader.onload = event => {
                try {
                    const importedState = JSON.parse(event.target.result);
                    this.state.data = importedState;
                    this.state.save();
                    this.updateUI();
                    showConfirmation("State imported successfully!");
                } catch (error) {
                    alert('Error importing state: ' + error.message);
                }
            };
            reader.readAsText(file);
        };
        input.click();
    }
}

// Initialize the app
document.addEventListener('DOMContentLoaded', () => {
    const app = new App();
    app.updateUI();
});