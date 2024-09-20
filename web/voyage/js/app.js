import store from './store.js';
import FragmentsColumn from './components/fragmentsColumn.js';
import PromptColumn from './components/promptColumn.js';
import OptionsColumn from './components/optionsColumn.js';
import HistorySection from './components/historySection.js';
import { showConfirmation } from './utils.js';
import { addImage } from './slices/imagesSlice.js';
import { setCurrentPrompt, addToHistory } from './slices/promptHistorySlice.js';
import { replacePromptFragments } from './slices/promptFragmentsSlice.js';
import { replaceImages } from './slices/imagesSlice.js';
import { replaceOptions } from './slices/optionsSlice.js';
import { replacePromptHistory } from './slices/promptHistorySlice.js';

log.setLevel("debug");

class App {
    constructor() {
        this.fragmentsColumn = new FragmentsColumn(store, () => this.updateUI());
        this.promptColumn = new PromptColumn(store, () => this.updateUI());
        this.optionsColumn = new OptionsColumn(store, () => this.updateUI());
        this.historySection = new HistorySection(store, () => this.updateUI());

        this.initModal();
        this.initImportExport();

        store.subscribe(() => this.updateUI());
    }

    initModal() {
        document.getElementById('confirm-add-image-btn').addEventListener('click', () => this.addImageURL());
        document.getElementById('cancel-add-image-btn').addEventListener('click', () => this.closeModal());
    }

    initImportExport() {
        document.getElementById('export-btn').addEventListener('click', () => this.exportState());
        document.getElementById('import-btn').addEventListener('click', () => this.importState());
    }

    updateUI() {
        this.fragmentsColumn.render();
        this.promptColumn.render();
        this.optionsColumn.render();
        this.historySection.render();
    }

    addImageURL() {
        const url = document.getElementById('new-image-url').value.trim();
        if (url) {
            const newImage = { url, thumbnail: "", alt: "New image" };
            store.dispatch(addImage(newImage));
            this.updateUI();
            this.closeModal();
            showConfirmation("Image added successfully!");
        }
    }

    closeModal() {
        document.getElementById('image-modal').style.display = 'none';
    }

    exportState() {
        const state = store.getState();
        const dataStr = "data:text/json;charset=utf-8," + encodeURIComponent(JSON.stringify(state));
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
                    
                    if (importedState.promptFragments) {
                        store.dispatch(replacePromptFragments(importedState.promptFragments));
                    }
                    if (importedState.images) {
                        store.dispatch(replaceImages(importedState.images));
                    }
                    if (importedState.options) {
                        store.dispatch(replaceOptions(importedState.options));
                    }
                    if (importedState.promptHistory) {
                        store.dispatch(replacePromptHistory(importedState.promptHistory));
                    }

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