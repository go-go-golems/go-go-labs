import { html, render } from 'https://cdn.jsdelivr.net/gh/lit/dist@3/all/lit-all.min.js';
import { escapeRegExp, showConfirmation } from '../utils.js';

class PromptColumn {
    constructor(state, updateUI) {
        this.state = state;
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
        const template = html`
            <h2>Current Prompt</h2>
            <div class="prompt-area">
                <textarea id="current-prompt" rows="6" .value=${this.state.get('current_prompt')}></textarea>
                <div class="buttons">
                    <button id="copy-clipboard-btn">Copy to Clipboard</button>
                    <button id="add-image-btn">Add Image URL</button>
                </div>
                <h3>Images</h3>
                <div id="images-list">
                    ${this.state.get('images').map((image, index) => html`
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
        this.state.set('current_prompt', value);
    }

    copyToClipboard() {
        let promptToCopy = this.state.get('current_prompt');
        const options = this.state.get('options');
        promptToCopy += ` --ar ${options.aspect_ratio} --v ${options.model_version}`;
        navigator.clipboard.writeText(promptToCopy).then(() => {
            showConfirmation("Prompt copied to clipboard!");
            this.state.addToHistory(promptToCopy);
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

    addImageToPrompt(url) {
        let currentPrompt = this.state.get('current_prompt');
        const newPrompt = currentPrompt ? `${url}, ${currentPrompt}` : url;
        this.state.set('current_prompt', newPrompt.trim());
        this.updateUI();
    }

    deleteImage(index) {
        const images = this.state.get('images');
        if (index >= 0 && index < images.length) {
            const url = images[index].url;
            images.splice(index, 1);
            this.state.set('images', images);

            let currentPrompt = this.state.get('current_prompt');
            if (url) {
                currentPrompt = this.removeImageFromPrompt(url, currentPrompt);
            }
            this.state.set('current_prompt', currentPrompt);

            this.updateUI();
            showConfirmation("Image deleted successfully!");
        } else {
            console.error("Invalid image index");
        }
    }

    isImageInPrompt(url) {
        const currentPrompt = this.state.get('current_prompt') || '';
        return currentPrompt.includes(url);
    }

    toggleImage(url) {
        let currentPrompt = this.state.get('current_prompt') || '';
        if (this.isImageInPrompt(url)) {
            currentPrompt = this.removeImageFromPrompt(url, currentPrompt);
        } else {
            currentPrompt = this.addImageToPrompt(url, currentPrompt);
        }
        this.state.set('current_prompt', currentPrompt);
        this.updateUI();
        showConfirmation(`Image "${url}" toggled`);
    }

    addImageToPrompt(url, prompt) {
        return prompt ? `${prompt}, ${url}` : url;
    }

    removeImageFromPrompt(url, prompt) {
        const regex = new RegExp(`(,\\s*)?${escapeRegExp(url)}(,\\s*)?`);
        let newPrompt = prompt.replace(regex, ',');
        // Remove leading/trailing commas and whitespace
        newPrompt = newPrompt.replace(/^,\s*/, '').replace(/,\s*$/, '');
        return newPrompt;
    }
}

export default PromptColumn;