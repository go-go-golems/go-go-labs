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
        this.state.get('prompt_fragments').forEach((fragment, index) => {
            const div = document.createElement('div');
            div.className = 'list-item';
            const checkbox = document.createElement('input');
            checkbox.type = 'checkbox';
            checkbox.id = `fragment-${index}`;
            checkbox.checked = this.state.get('checked_fragments').includes(index);
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
        const checkedFragments = this.state.get('checked_fragments').map(index => fragments[index]);
        
        if (checkedFragments.length === 0) return;

        const numberToSelect = Math.floor(Math.random() * checkedFragments.length) + 1;
        const shuffled = [...checkedFragments].sort(() => 0.5 - Math.random());
        const selectedFragments = shuffled.slice(0, numberToSelect);
        const fragmentsText = selectedFragments.join(', ');

        let currentPrompt = this.state.get('current_prompt');
        const newPrompt = currentPrompt ? `${currentPrompt}, ${fragmentsText}` : fragmentsText;
        this.state.set('current_prompt', newPrompt.trim());
        this.updateUI();
        showConfirmation("Random fragments added to prompt!");
    }

    updateCheckedFragments(index, isChecked) {
        const checkedFragments = this.state.get('checked_fragments');
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