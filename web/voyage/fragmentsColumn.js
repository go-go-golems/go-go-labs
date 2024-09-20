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
    }

    render() {
        const fragmentsList = this.element.querySelector('#fragments-list');
        fragmentsList.innerHTML = '';
        const fragments = this.state.get('prompt_fragments') || [];
        const checkedFragments = this.state.get('checked_fragments') || [];
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
            label.addEventListener('click', () => this.addFragmentToPrompt(fragment));
            const deleteBtn = document.createElement('button');
            deleteBtn.textContent = 'Delete';
            deleteBtn.addEventListener('click', () => this.deleteFragment(index));
            div.appendChild(checkbox);
            div.appendChild(label);
            div.appendChild(deleteBtn);
            fragmentsList.appendChild(div);
        });
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

    addFragmentToPrompt(fragment) {
        let currentPrompt = this.state.get('current_prompt');
        const newPrompt = currentPrompt ? `${currentPrompt}, ${fragment}` : fragment;
        this.state.set('current_prompt', newPrompt.trim());
        this.updateUI();
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
}

export default FragmentsColumn;