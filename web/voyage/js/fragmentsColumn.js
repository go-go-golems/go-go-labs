import { html, render } from 'https://cdn.jsdelivr.net/gh/lit/dist@3/all/lit-all.min.js';
import { showConfirmation } from './utils.js';

class FragmentsColumn {
    constructor(state, updateUI) {
        this.state = state;
        this.updateUI = updateUI;
        this.element = document.getElementById('fragments-column');
        this.init();
    }

    init() {
        this.element.addEventListener('click', (e) => {
            if (e.target.id === 'add-fragment-btn') this.addFragment();
            if (e.target.id === 'randomize-btn') this.randomizeAndAddFragments();
            if (e.target.id === 'unselect-all-btn') this.unselectAllFragments();
            if (e.target.id === 'save-selection-btn') this.openSaveSelectionModal();
        });
    }

    render() {
        const fragments = this.state.get('prompt_fragments') || [];
        const checkedFragments = this.state.get('checked_fragments') || [];
        const currentPrompt = this.state.get('current_prompt') || '';
        const savedSelections = this.state.get('saved_selections') || [];

        const template = html`
            <h2>Prompt Fragments</h2>
            <div class="checkbox-group" id="fragments-list">
                ${fragments.map((fragment, index) => this.renderFragment(fragment, index, checkedFragments, currentPrompt))}
            </div>
            <button id="add-fragment-btn">Add New Fragment</button>
            <div class="button-group">
                <button id="randomize-btn" class="randomize-btn">Randomize</button>
                <button id="unselect-all-btn">Unselect All</button>
            </div>
            <button id="save-selection-btn">Save Fragment Selection</button>
            <h3>Saved Selections</h3>
            <div id="saved-selections-list">
                ${savedSelections.map((savedSelection, index) => this.renderSavedSelection(savedSelection, index))}
            </div>
        `;

        render(template, this.element);
    }

    renderFragment(fragment, index, checkedFragments, currentPrompt) {
        return html`
            <div class="list-item">
                <input type="checkbox" id="fragment-${index}" 
                       ?checked=${checkedFragments.includes(index)}
                       @change=${(e) => this.updateCheckedFragments(index, e.target.checked)}>
                <label for="fragment-${index}" 
                       class=${this.isFragmentInPrompt(fragment, currentPrompt) ? 'active-fragment' : ''}
                       @click=${(e) => { e.preventDefault(); this.toggleFragment(fragment); }}>
                    ${fragment}
                </label>
                <button @click=${() => this.deleteFragment(index)}>Delete</button>
            </div>
        `;
    }

    renderSavedSelection(savedSelection, index) {
        return html`
            <div class="list-item">
                <span @click=${() => this.restoreSavedSelection(savedSelection.selection)}>${savedSelection.name}</span>
                <button @click=${() => this.deleteSavedSelection(index)}>Delete</button>
            </div>
        `;
    }

    isFragmentInPrompt(fragment, prompt) {
        return prompt.includes(fragment);
    }

    toggleFragment(fragment) {
        let currentPrompt = this.state.get('current_prompt') || "";
        if (this.isFragmentInPrompt(fragment, currentPrompt)) {
            currentPrompt = this.removeFragmentFromPrompt(fragment, currentPrompt);
        } else {
            currentPrompt = this.addFragmentToPrompt(fragment, currentPrompt);
        }
        this.state.set('current_prompt', currentPrompt);
        this.updateUI();
        showConfirmation(`Fragment "${fragment}" toggled`);
    }

    addFragmentToPrompt(fragment, prompt) {
        return prompt ? `${prompt}, ${fragment}` : fragment;
    }

    removeFragmentFromPrompt(fragment, prompt) {
        const regex = new RegExp(`(,\\s*)?${fragment}(,\\s*)?`);
        let newPrompt = prompt.replace(regex, ',');
        // Remove leading/trailing commas and whitespace
        newPrompt = newPrompt.replace(/^,\s*/, '').replace(/,\s*$/, '');
        return newPrompt;
    }

    addFragment() {
        const fragment = prompt("Enter new prompt fragment:");
        if (fragment) {
            const fragments = this.state.get('prompt_fragments');
            fragments.push(fragment.trim());
            this.state.set('prompt_fragments', fragments);
            this.updateUI();
            showConfirmation("Fragment added successfully!");
        }
    }

    deleteFragment(index) {
        const fragments = this.state.get('prompt_fragments');
        fragments.splice(index, 1);
        this.state.set('prompt_fragments', fragments);
        
        // Also update checked_fragments
        const checkedFragments = this.state.get('checked_fragments') || [];
        const updatedCheckedFragments = checkedFragments.filter(i => i !== index).map(i => i > index ? i - 1 : i);
        this.state.set('checked_fragments', updatedCheckedFragments);
        
        this.updateUI();
        showConfirmation("Fragment deleted successfully!");
    }

    randomizeAndAddFragments() {
        const fragments = this.state.get('prompt_fragments');
        const checkedFragments = this.state.get('checked_fragments') || [];
        const selectedFragments = checkedFragments.map(index => fragments[index]);
        
        if (selectedFragments.length === 0) return;

        const currentPrompt = this.state.get('current_prompt') || '';
        const currentFragments = currentPrompt.split(',').map(f => f.trim());

        const availableFragments = selectedFragments.filter(f => !currentFragments.includes(f));

        if (availableFragments.length === 0) {
            showConfirmation("All selected fragments are already in the prompt!");
            return;
        }

        const numberToSelect = Math.min(
            Math.floor(Math.random() * availableFragments.length) + 1,
            availableFragments.length
        );
        const shuffled = availableFragments.sort(() => 0.5 - Math.random());
        const randomizedFragments = shuffled.slice(0, numberToSelect);
        const fragmentsText = randomizedFragments.join(', ');

        const newPrompt = currentPrompt ? `${currentPrompt}, ${fragmentsText}` : fragmentsText;
        this.state.set('current_prompt', newPrompt.trim());
        this.updateUI();
        showConfirmation("Random fragments added to prompt!");
    }

    updateCheckedFragments(index, isChecked) {
        const checkedFragments = this.state.get('checked_fragments') || [];
        if (isChecked) {
            if (!checkedFragments.includes(index)) {
                checkedFragments.push(index);
            }
        } else {
            const indexToRemove = checkedFragments.indexOf(index);
            if (indexToRemove !== -1) {
                checkedFragments.splice(indexToRemove, 1);
            }
        }
        this.state.set('checked_fragments', checkedFragments);
    }

    unselectAllFragments() {
        const checkedFragments = [];
        this.state.set('checked_fragments', checkedFragments);
        this.updateUI();
        showConfirmation("All fragments unselected!");
    }

    openSaveSelectionModal() {
        const modal = document.getElementById('save-selection-modal');
        modal.style.display = 'flex';
        document.getElementById('selection-name').value = '';
        document.getElementById('selection-name').focus();
    }

    saveFragmentSelection() {
        const name = document.getElementById('selection-name').value.trim();
        if (name) {
            const checkedFragments = this.state.get('checked_fragments');
            const savedSelections = this.state.get('saved_selections') || [];
            savedSelections.push({ name, selection: checkedFragments });
            this.state.set('saved_selections', savedSelections);
            this.updateUI();
            this.closeSaveSelectionModal();
            showConfirmation("Fragment selection saved!");
        }
    }

    closeSaveSelectionModal() {
        document.getElementById('save-selection-modal').style.display = 'none';
    }

    restoreSavedSelection(selection) {
        this.state.set('checked_fragments', selection);
        this.updateUI();
        showConfirmation("Saved selection restored!");
    }

    deleteSavedSelection(index) {
        const savedSelections = this.state.get('saved_selections') || [];
        savedSelections.splice(index, 1);
        this.state.set('saved_selections', savedSelections);
        this.updateUI();
        showConfirmation("Saved selection deleted!");
    }
}

export default FragmentsColumn;