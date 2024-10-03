import { html, render } from 'https://cdn.jsdelivr.net/gh/lit/dist@3/all/lit-all.min.js';
import { escapeRegExp, showConfirmation } from '../utils.js';
import { setCurrentPrompt, addToHistory } from '../slices/promptHistorySlice.js';
import { addImage, deleteImage } from '../slices/imagesSlice.js';
import { setAspectRatio, setModelVersion } from '../slices/optionsSlice.js';

class PromptColumn {
    constructor(store, updateUI) {
        this.store = store;
        this.updateUI = updateUI;
        this.element = document.getElementById('prompt-column');
        this.init();
    }

    init() {
        this.element.addEventListener('click', (e) => {
            if (e.target.id === 'copy-clipboard-btn') this.copyToClipboard();
            if (e.target.id === 'add-image-btn') this.openModal();
        });
        this.element.addEventListener('input', (e) => {
            if (e.target.id === 'current-prompt') this.updateCurrentPrompt(e.target.value);
        });
    }

    render() {
        const state = this.store.getState();
        const currentPrompt = state.promptHistory.current_prompt || '';
        const images = state.images.images || [];

        const template = html`
            <h2>Current Prompt</h2>
            <div class="prompt-area">
                <textarea id="current-prompt" rows="6" .value=${currentPrompt}></textarea>
                <div class="buttons">
                    <button id="copy-clipboard-btn">Copy to Clipboard</button>
                    <button id="add-image-btn">Add Image URL</button>
                </div>
                <h3>Images</h3>
                <div id="images-list">
                    ${images.map((image, index) => html`
                        <div class="list-item">
                            <img src=${image.thumbnail || image.url} 
                                 alt=${image.alt} 
                                 style="cursor: pointer; max-height: 100px;"
                                 class=${this.isImageInPrompt(image.url) ? 'active-image' : ''}
                                 @click=${() => this.toggleImage(image.url)}>
                            <button @click=${() => this.deleteImage(index)}>Delete</button>
                        </div>
                    `)}
                </div>
            </div>
        `;

        render(template, this.element);
    }

    updateCurrentPrompt(value) {
        this.store.dispatch(setCurrentPrompt(value));
    }

    copyToClipboard() {
        const state = this.store.getState();
        let promptToCopy = state.promptHistory.current_prompt;
        const options = state.options;
        promptToCopy += ` --ar ${options.aspect_ratio} --v ${options.model_version}`;
        navigator.clipboard.writeText(promptToCopy).then(() => {
            showConfirmation("Prompt copied to clipboard!");
            this.store.dispatch(addToHistory(promptToCopy));
            this.updateUI();
        }).catch(err => {
            alert('Failed to copy: ', err);
        });
    }

    openModal() {
        document.getElementById('image-modal').style.display = 'flex';
        document.getElementById('new-image-url').value = '';
        document.getElementById('new-image-url').focus();
    }

    toggleImage(url) {
        const state = this.store.getState();
        let currentPrompt = state.promptHistory.current_prompt || '';
        if (this.isImageInPrompt(url)) {
            currentPrompt = this.removeImageFromPrompt(url, currentPrompt);
        } else {
            currentPrompt = this.addImageToPrompt(url, currentPrompt);
        }
        this.store.dispatch(setCurrentPrompt(currentPrompt));
        this.updateUI();
        showConfirmation(`Image "${url}" toggled`);
    }

    addImageToPrompt(url, prompt) {
        if (this.isImageInPrompt(url)) {
            return prompt;
        }
        return prompt ? `${url} ${prompt}` : url;
    }

    removeImageFromPrompt(url, prompt) {
        const regex = new RegExp(`(,\\s*)?${escapeRegExp(url)}(,\\s*)?`, 'g');
        let newPrompt = prompt.replace(regex, ',');
        newPrompt = newPrompt.replace(/^,\s*/, '').replace(/,\s*$/, '');
        return newPrompt;
    }

    deleteImage(index) {
        this.store.dispatch(deleteImage(index));
        this.updateUI();
        showConfirmation("Image deleted successfully!");
    }

    isImageInPrompt(url) {
        const currentPrompt = this.store.getState().promptHistory.current_prompt || '';
        return currentPrompt.includes(url);
    }
}

export default PromptColumn;