import { escapeRegExp, showConfirmation } from './utils.js';

class PromptColumn {
    constructor(state, updateUI) {
        this.state = state;
        this.updateUI = updateUI;
        this.element = document.getElementById('prompt-column');
        this.init();
    }

    init() {
        this.element.querySelector('#copy-clipboard-btn').addEventListener('click', () => this.copyToClipboard());
        this.element.querySelector('#add-image-btn').addEventListener('click', () => this.openModal());
        this.element.querySelector('#current-prompt').addEventListener('input', (e) => this.updateCurrentPrompt(e.target.value));
    }

    render() {
        const currentPrompt = this.element.querySelector('#current-prompt');
        currentPrompt.value = this.state.get('current_prompt');

        const imagesList = this.element.querySelector('#images-list');
        imagesList.innerHTML = '';
        this.state.get('images').forEach((image, index) => {
            const div = document.createElement('div');
            div.className = 'list-item';
            const img = document.createElement('img');
            img.src = image.thumbnail || image.url;
            img.alt = image.alt;
            img.style.cursor = 'pointer';
            img.addEventListener('click', () => this.addImageToPrompt(image.url));
            const addBtn = document.createElement('button');
            addBtn.textContent = 'Add to Prompt';
            addBtn.addEventListener('click', () => this.addImageToPrompt(image.url));
            const deleteBtn = document.createElement('button');
            deleteBtn.textContent = 'Delete';
            deleteBtn.addEventListener('click', () => this.deleteImage(index));
            div.appendChild(img);
            div.appendChild(addBtn);
            div.appendChild(deleteBtn);
            imagesList.appendChild(div);
        });
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
                const regex = new RegExp(`\\b${escapeRegExp(url)}\\b,?\\s*`, 'g');
                currentPrompt = currentPrompt.replace(regex, '').replace(/,\s*,/g, ',').replace(/,\s*$/, '').trim();
            }
            this.state.set('current_prompt', currentPrompt);

            this.updateUI();
            showConfirmation("Image deleted successfully!");
        } else {
            console.error("Invalid image index");
        }
    }
}

export default PromptColumn;