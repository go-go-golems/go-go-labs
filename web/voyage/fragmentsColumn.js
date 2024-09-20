import { showConfirmation } from './utils.js';

class FragmentsColumn {
    constructor(state, updateUI) {
        this.state = state;
        this.updateUI = updateUI;
        this.element = document.getElementById('fragments-column');
        this.init();
    }

    init() {
        this.element.querySelector('#add-fragment-btn').addEventListener('click', () => this.addFragment());
        this.element.querySelector('#randomize-btn').addEventListener('click', () => this.randomizeAndAddFragments());
        this.element.querySelector('#unselect-all-btn').addEventListener('click', () => this.unselectAllFragments());
        this.element.querySelector('#save-selection-btn').addEventListener('click', () => this.openSaveSelectionModal());
        this.renderSavedSelections();
        this.renderFragments();
    }

    renderFragments() {
        const fragmentsList = this.element.querySelector('#fragments-list');
        fragmentsList.innerHTML = '';
        const fragments = this.state.get('prompt_fragments') || [];
        const checkedFragments = this.state.get('checked_fragments') || [];
        const currentPrompt = this.state.get('current_prompt') || '';

        console.log("currentPrompt", currentPrompt);

        fragments.forEach((fragment, index) => {
            const div = document.createElement('div');
            div.className = 'list-item';
            const checkbox = document.createElement('input');
            checkbox.type = 'checkbox';
            checkbox.id = `fragment-${index}`;
            checkbox.checked = checkedFragments.includes(index);
            checkbox.addEventListener('change', () => this.updateCheckedFragments(index, checkbox.checked));
            
            const label = document.createElement('label');
            label.htmlFor = `fragment-${index}`;
            label.textContent = fragment;
            label.addEventListener('click', (e) => {
                e.preventDefault();
                this.toggleFragment(fragment);
            });
            
            // Add visual feedback for active fragments
            if (this.isFragmentInPrompt(fragment, currentPrompt)) {
                label.classList.add('active-fragment');
            }

            const deleteBtn = document.createElement('button');
            deleteBtn.textContent = 'Delete';
            deleteBtn.addEventListener('click', () => this.deleteFragment(index));
            
            div.appendChild(checkbox);
            div.appendChild(label);
            div.appendChild(deleteBtn);
            fragmentsList.appendChild(div);
        });
    }

    isFragmentInPrompt(fragment, prompt) {
        return prompt.includes(fragment);
    }

    render() {
        this.renderFragments();
        this.renderSavedSelections();
    }

    toggleFragment(fragment) {
        let currentPrompt = this.state.get('current_prompt') || "";
        console.log("currentPrompt", currentPrompt);
        console.log("toggle fragment", fragment);
        if (this.isFragmentInPrompt(fragment, currentPrompt)) {
            console.log("removing fragment", fragment, currentPrompt);
            currentPrompt = this.removeFragmentFromPrompt(fragment, currentPrompt);
        } else {
            console.log("adding fragment", fragment, currentPrompt);
            currentPrompt = this.addFragmentToPrompt(fragment, currentPrompt);
            console.log("currentPrompt after adding", currentPrompt);
        }
        console.log("currentPrompt after", currentPrompt);
        this.state.set('current_prompt', currentPrompt);
        this.updateUI();
        showConfirmation(`Fragment "${fragment}" toggled`);
    }

    addFragmentToPrompt(fragment, prompt) {
        console.log("adding fragment to prompt", fragment, prompt);
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

    renderSavedSelections() {
        const savedSelectionsList = this.element.querySelector('#saved-selections-list');
        savedSelectionsList.innerHTML = '';
        const savedSelections = this.state.get('saved_selections') || [];
        savedSelections.forEach((savedSelection, index) => {
            const div = document.createElement('div');
            div.className = 'list-item';
            const span = document.createElement('span');
            span.textContent = savedSelection.name;
            span.style.cursor = 'pointer';
            span.addEventListener('click', () => this.restoreSavedSelection(savedSelection.selection));
            const deleteBtn = document.createElement('button');
            deleteBtn.textContent = 'Delete';
            deleteBtn.addEventListener('click', () => this.deleteSavedSelection(index));
            div.appendChild(span);
            div.appendChild(deleteBtn);
            savedSelectionsList.appendChild(div);
        });
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