# Refactoring Your Application to Use Redux Toolkit for State Management

In modern web development, managing application state efficiently and predictably is crucial, especially as your application grows in complexity. While hand-rolling state management solutions can work for smaller projects, leveraging established libraries like **Redux Toolkit (RTK Toolkit)** can significantly enhance maintainability, scalability, and developer experience.

In this tutorial, we'll guide you through updating your existing JavaScript application to use Redux Toolkit for state management instead of your custom `State` class. We'll maintain your current setup using CDN imports and Lit for rendering components.

## Table of Contents

1. [Why Redux Toolkit?](#why-redux-toolkit)
2. [Setting Up Redux Toolkit via CDN](#setting-up-redux-toolkit-via-cdn)
3. [Configuring the Redux Store](#configuring-the-redux-store)
4. [Defining Slices](#defining-slices)
5. [Integrating Redux Store with Your Application](#integrating-redux-store-with-your-application)
6. [Refactoring Components to Use Redux](#refactoring-components-to-use-redux)
7. [Handling State Persistence with LocalStorage](#handling-state-persistence-with-localstorage)
8. [Final Thoughts](#final-thoughts)

---

## Why Redux Toolkit?

**Redux Toolkit** is the official, opinionated, batteries-included toolset for efficient Redux development. It simplifies Redux setup and reduces boilerplate, making state management more straightforward and less error-prone.

**Benefits of Using Redux Toolkit:**

- **Simplified Configuration:** Quickly set up the store with sensible defaults.
- **Immutability Handling:** Automatically handle immutable updates.
- **Built-in Middleware:** Includes useful middleware like `redux-thunk` for async logic.
- **Developer Tools Integration:** Seamlessly integrates with Redux DevTools for debugging.
- **Scalability:** Easily manage complex state logic as your application grows.

By migrating to Redux Toolkit, you leverage a robust state management solution that enhances your application's scalability and maintainability.

---

## Setting Up Redux Toolkit via CDN

Since your application utilizes CDN imports, we'll integrate Redux Toolkit and its dependencies using CDN links.

### 1. Include Redux Toolkit and Dependencies

Add the following `<script>` tags to your HTML file before your application scripts to include Redux and Redux Toolkit via CDN:

```html
<!-- Redux -->
<script src="https://cdn.jsdelivr.net/npm/redux@4.2.1/dist/redux.min.js"></script>

<!-- Redux Toolkit -->
<script src="https://cdn.jsdelivr.net/npm/@reduxjs/toolkit@1.9.5/dist/redux-toolkit.umd.min.js"></script>

<!-- Optional: Redux DevTools Extension -->
<script src="https://cdn.jsdelivr.net/npm/redux-devtools-extension@2.13.9/dist/redux-devtools-extension.umd.min.js"></script>
```

**Note:** Ensure these scripts are loaded before your application scripts to make Redux and RTK available globally.

---

## Configuring the Redux Store

With Redux Toolkit included, the next step is to configure the Redux store, which holds the entire state of your application.

### 1. Create a `store.js` File

Create a new file named `store.js` in your `web/voyage/js/` directory.

```javascript
// === BEGIN: web/voyage/js/store.js ===

const { configureStore } = window.RTK;

// Import slices (to be defined later)
import { promptFragmentsSlice } from './slices/promptFragmentsSlice.js';
import { imagesSlice } from './slices/imagesSlice.js';
import { optionsSlice } from './slices/optionsSlice.js';
import { promptHistorySlice } from './slices/promptHistorySlice.js';
import { uiSlice } from './slices/uiSlice.js';

// Configure the Redux store with slices
const store = configureStore({
    reducer: {
        promptFragments: promptFragmentsSlice.reducer,
        images: imagesSlice.reducer,
        options: optionsSlice.reducer,
        promptHistory: promptHistorySlice.reducer,
        ui: uiSlice.reducer,
    },
    // Integrate Redux DevTools Extension if available
    devTools: window.__REDUX_DEVTOOLS_EXTENSION__ ? window.__REDUX_DEVTOOLS_EXTENSION__() : false,
});

export default store;

// === END: web/voyage/js/store.js ===
```

### 2. Define Slices

Slices in Redux Toolkit encapsulate the reducer logic and actions for a specific part of the state. We'll create separate slices for different state segments.

---

## Defining Slices

We'll define slices corresponding to different parts of your application's state: `prompt_fragments`, `images`, `options`, `prompt_history`, and `ui`.

### 1. Create a `slices` Directory

Organize your slices by creating a `slices` directory inside `web/voyage/js/`.

```
web/voyage/js/
├── app.js
├── components/
├── slices/
│   ├── promptFragmentsSlice.js
│   ├── imagesSlice.js
│   ├── optionsSlice.js
│   ├── promptHistorySlice.js
│   └── uiSlice.js
├── store.js
├── state.js
└── utils.js
```

### 2. Define Each Slice

#### a. `promptFragmentsSlice.js`

```javascript
// === BEGIN: web/voyage/js/slices/promptFragmentsSlice.js ===

const { createSlice } = window.RTK;

const initialState = {
    prompt_fragments: [
        "a majestic lion",
        "in a lush jungle",
        "with vibrant colors",
        "photorealistic style"
    ],
    checked_fragments: [],
    saved_selections: []
};

export const promptFragmentsSlice = createSlice({
    name: 'promptFragments',
    initialState,
    reducers: {
        addFragment: (state, action) => {
            state.prompt_fragments.push(action.payload);
        },
        deleteFragment: (state, action) => {
            state.prompt_fragments.splice(action.payload, 1);
            // Adjust checked_fragments indices
            state.checked_fragments = state.checked_fragments
                .filter(index => index !== action.payload)
                .map(index => (index > action.payload ? index - 1 : index));
        },
        toggleCheckedFragment: (state, action) => {
            const index = action.payload;
            if (state.checked_fragments.includes(index)) {
                state.checked_fragments = state.checked_fragments.filter(i => i !== index);
            } else {
                state.checked_fragments.push(index);
            }
        },
        unselectAllFragments: (state) => {
            state.checked_fragments = [];
        },
        saveSelection: (state, action) => {
            state.saved_selections.push(action.payload);
        },
        deleteSavedSelection: (state, action) => {
            state.saved_selections.splice(action.payload, 1);
        }
    }
});

export const {
    addFragment,
    deleteFragment,
    toggleCheckedFragment,
    unselectAllFragments,
    saveSelection,
    deleteSavedSelection
} = promptFragmentsSlice.actions;

// === END: web/voyage/js/slices/promptFragmentsSlice.js ===
```

#### b. `imagesSlice.js`

```javascript
// === BEGIN: web/voyage/js/slices/imagesSlice.js ===

const { createSlice } = window.RTK;

const initialState = {
    images: [
        { url: "https://example.com/lion.jpg", thumbnail: "", alt: "Lion" },
        { url: "https://example.com/jungle.jpg", thumbnail: "", alt: "Jungle" }
    ]
};

export const imagesSlice = createSlice({
    name: 'images',
    initialState,
    reducers: {
        addImage: (state, action) => {
            state.images.unshift(action.payload);
        },
        deleteImage: (state, action) => {
            state.images.splice(action.payload, 1);
        }
    }
});

export const { addImage, deleteImage } = imagesSlice.actions;

// === END: web/voyage/js/slices/imagesSlice.js ===
```

#### c. `optionsSlice.js`

```javascript
// === BEGIN: web/voyage/js/slices/optionsSlice.js ===

const { createSlice } = window.RTK;

const initialState = {
    aspect_ratio: "16:9",
    model_version: "v5"
};

export const optionsSlice = createSlice({
    name: 'options',
    initialState,
    reducers: {
        setAspectRatio: (state, action) => {
            state.aspect_ratio = action.payload;
        },
        setModelVersion: (state, action) => {
            state.model_version = action.payload;
        }
    }
});

export const { setAspectRatio, setModelVersion } = optionsSlice.actions;

// === END: web/voyage/js/slices/optionsSlice.js ===
```

#### d. `promptHistorySlice.js`

```javascript
// === BEGIN: web/voyage/js/slices/promptHistorySlice.js ===

const { createSlice } = window.RTK;

const initialState = {
    prompt_history: [
        "a serene lake at sunset",
        "cyberpunk cityscape with neon lights",
        "abstract geometric patterns in pastel colors"
    ],
    search_query: "",
    current_prompt: ""
};

export const promptHistorySlice = createSlice({
    name: 'promptHistory',
    initialState,
    reducers: {
        addToHistory: (state, action) => {
            const prompt = action.payload;
            if (prompt && prompt !== state.prompt_history[0]) {
                state.prompt_history.unshift(prompt);
            }
        },
        setSearchQuery: (state, action) => {
            state.search_query = action.payload;
        },
        setCurrentPrompt: (state, action) => {
            state.current_prompt = action.payload;
        }
    }
});

export const { addToHistory, setSearchQuery, setCurrentPrompt } = promptHistorySlice.actions;

// === END: web/voyage/js/slices/promptHistorySlice.js ===
```

#### e. `uiSlice.js`

This slice manages UI-related state, such as modal visibility.

```javascript
// === BEGIN: web/voyage/js/slices/uiSlice.js ===

const { createSlice } = window.RTK;

const initialState = {
    // Define UI-related state if needed
    // Example:
    // isImageModalOpen: false
};

export const uiSlice = createSlice({
    name: 'ui',
    initialState,
    reducers: {
        // Define UI-related reducers if needed
    }
});

export const {
    // Export UI-related actions if defined
} = uiSlice.actions;

// === END: web/voyage/js/slices/uiSlice.js ===
```

---

## Integrating Redux Store with Your Application

Now that the store and slices are defined, we'll integrate the Redux store into your application and remove the custom `State` class.

### 1. Update `index.html`

Ensure that the `store.js` is loaded before your main application script (`app.js`).

```html
<!-- ... other scripts ... -->

<!-- Redux -->
<script src="https://cdn.jsdelivr.net/npm/redux@4.2.1/dist/redux.min.js"></script>
<!-- Redux Toolkit -->
<script src="https://cdn.jsdelivr.net/npm/@reduxjs/toolkit@1.9.5/dist/redux-toolkit.umd.min.js"></script>
<!-- Redux DevTools Extension (optional) -->
<script src="https://cdn.jsdelivr.net/npm/redux-devtools-extension@2.13.9/dist/redux-devtools-extension.umd.min.js"></script>

<!-- Store -->
<script type="module" src="./js/store.js"></script>

<!-- App -->
<script type="module" src="./js/app.js"></script>
```

### 2. Remove the Custom `State` Class

Since Redux will handle state management, you can remove the `state.js` file and any imports related to it.

- **Delete `state.js`:** Remove `web/voyage/js/state.js`.
- **Update Imports:** Remove all imports of `State` from your components and `app.js`.

### 3. Update `app.js`

Refactor `app.js` to use the Redux store instead of the custom `State` class.

```javascript
// === BEGIN: web/voyage/js/app.js ===

import store from './store.js';
import FragmentsColumn from './components/fragmentsColumn.js';
import PromptColumn from './components/promptColumn.js';
import OptionsColumn from './components/optionsColumn.js';
import HistorySection from './components/historySection.js';
import { showConfirmation } from './utils.js';
import { addImage } from './slices/imagesSlice.js';
import { setCurrentPrompt, addToHistory } from './slices/promptHistorySlice.js';

class App {
    constructor() {
        this.fragmentsColumn = new FragmentsColumn(store, () => this.updateUI());
        this.promptColumn = new PromptColumn(store, () => this.updateUI());
        this.optionsColumn = new OptionsColumn(store, () => this.updateUI());
        this.historySection = new HistorySection(store, () => this.updateUI());

        this.initModal();
        this.initImportExport();

        // Subscribe to store updates
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
        // No need for this.state.save(); Redux handles state persistence
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
                    // Dispatch actions to replace the state
                    // Note: Redux Toolkit does not support replacing the entire state out-of-the-box.
                    // For simplicity, you might need to recreate the store or implement a custom action.
                    // Here, we'll just log it.
                    console.warn("Importing state is not implemented. Please reload the app and set state manually.");
                    showConfirmation("State import is not implemented.");
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

// === END: web/voyage/js/app.js ===
```

**Key Changes:**

- **Store Integration:** Import the Redux store and pass it to components.
- **Dispatching Actions:** Use `store.dispatch` to update the state.
- **State Subscription:** Subscribe to store updates to trigger UI re-renders.
- **Removed `State` Class:** All state interactions now go through Redux.

---

## Refactoring Components to Use Redux

Each component previously relied on the `State` class for accessing and modifying state. We'll refactor them to interact with the Redux store instead.

### 1. Update `fragmentsColumn.js`

```javascript
// === BEGIN: web/voyage/js/components/fragmentsColumn.js ===

import { html, render } from 'https://cdn.jsdelivr.net/gh/lit/dist@3/all/lit-all.min.js';
import { showConfirmation } from '../utils.js';
import { addFragment, deleteFragment, toggleCheckedFragment, unselectAllFragments, saveSelection, deleteSavedSelection } from '../slices/promptFragmentsSlice.js';

class FragmentsColumn {
    constructor(store, updateUI) {
        this.store = store;
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
        const state = this.store.getState();
        const fragments = state.promptFragments.prompt_fragments || [];
        const checkedFragments = state.promptFragments.checked_fragments || [];
        const currentPrompt = state.promptHistory.current_prompt || '';
        const savedSelections = state.promptFragments.saved_selections || [];

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
                <button @click=${() => this.deleteFragmentByIndex(index)}>Delete</button>
            </div>
        `;
    }

    renderSavedSelection(savedSelection, index) {
        return html`
            <div class="list-item">
                <span @click=${() => this.restoreSavedSelection(savedSelection.selection)}>${savedSelection.name}</span>
                <button @click=${() => this.deleteSavedSelectionByIndex(index)}>Delete</button>
            </div>
        `;
    }

    isFragmentInPrompt(fragment, prompt) {
        return prompt.includes(fragment);
    }

    toggleFragment(fragment) {
        let state = this.store.getState();
        let currentPrompt = state.promptHistory.current_prompt || "";

        if (this.isFragmentInPrompt(fragment, currentPrompt)) {
            currentPrompt = this.removeFragmentFromPrompt(fragment, currentPrompt);
        } else {
            currentPrompt = this.addFragmentToPrompt(fragment, currentPrompt);
        }

        this.store.dispatch(setCurrentPrompt(currentPrompt.trim()));
        this.store.dispatch(addToHistory(currentPrompt.trim()));
        this.updateUI();
        showConfirmation(`Fragment "${fragment}" toggled`);
    }

    addFragmentToPrompt(fragment, prompt) {
        return prompt ? `${prompt}, ${fragment}` : fragment;
    }

    removeFragmentFromPrompt(fragment, prompt) {
        const regex = new RegExp(`(,\\s*)?${this.escapeRegExp(fragment)}(,\\s*)?`, 'g');
        let newPrompt = prompt.replace(regex, ',');
        // Remove leading/trailing commas and whitespace
        newPrompt = newPrompt.replace(/^,\s*/, '').replace(/,\s*$/, '');
        return newPrompt;
    }

    escapeRegExp(string) {
        return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    }

    addFragment() {
        const fragment = prompt("Enter new prompt fragment:");
        if (fragment) {
            this.store.dispatch(addFragment(fragment.trim()));
            this.updateUI();
            showConfirmation("Fragment added successfully!");
        }
    }

    deleteFragmentByIndex(index) {
        this.store.dispatch(deleteFragment(index));
        this.updateUI();
        showConfirmation("Fragment deleted successfully!");
    }

    randomizeAndAddFragments() {
        const state = this.store.getState();
        const fragments = state.promptFragments.prompt_fragments;
        const checkedFragments = state.promptFragments.checked_fragments;
        const selectedFragments = checkedFragments.map(index => fragments[index]);

        if (selectedFragments.length === 0) return;

        const currentPrompt = state.promptHistory.current_prompt || '';
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
        this.store.dispatch(setCurrentPrompt(newPrompt.trim()));
        this.store.dispatch(addToHistory(newPrompt.trim()));
        this.updateUI();
        showConfirmation("Random fragments added to prompt!");
    }

    updateCheckedFragments(index, isChecked) {
        this.store.dispatch(toggleCheckedFragment(index));
    }

    unselectAllFragments() {
        this.store.dispatch(unselectAllFragments());
        this.updateUI();
        showConfirmation("All fragments unselected!");
    }

    openSaveSelectionModal() {
        const modal = document.getElementById('save-selection-modal');
        modal.style.display = 'flex';
        document.getElementById('selection-name').value = '';
        document.getElementById('selection-name').focus();

        // Attach event listeners if not already attached
        if (!this.saveSelectionListener) {
            this.saveSelectionListener = () => this.saveFragmentSelection();
            document.getElementById('confirm-save-selection-btn').addEventListener('click', this.saveSelectionListener);
            document.getElementById('cancel-save-selection-btn').addEventListener('click', () => this.closeSaveSelectionModal());
        }
    }

    saveFragmentSelection() {
        const name = document.getElementById('selection-name').value.trim();
        if (name) {
            const state = this.store.getState();
            const checkedFragments = state.promptFragments.checked_fragments;
            const selection = { name, selection: checkedFragments };
            this.store.dispatch(saveSelection(selection));
            this.updateUI();
            this.closeSaveSelectionModal();
            showConfirmation("Fragment selection saved!");
        }
    }

    closeSaveSelectionModal() {
        document.getElementById('save-selection-modal').style.display = 'none';
    }

    restoreSavedSelection(selection) {
        this.store.dispatch(setCheckedFragments(selection));
        this.updateUI();
        showConfirmation("Saved selection restored!");
    }

    deleteSavedSelectionByIndex(index) {
        this.store.dispatch(deleteSavedSelection(index));
        this.updateUI();
        showConfirmation("Saved selection deleted!");
    }
}

export default FragmentsColumn;

// === END: web/voyage/js/components/fragmentsColumn.js ===
```

**Key Changes:**

- **Store Access:** Components receive the Redux store as a constructor parameter.
- **Dispatching Actions:** Use `store.dispatch` to modify state.
- **Selecting State:** Use `store.getState()` to access current state.
- **Removed Direct State Manipulation:** All state changes go through Redux actions.

### 2. Update `promptColumn.js`

```javascript
// === BEGIN: web/voyage/js/components/promptColumn.js ===

import { html, render } from 'https://cdn.jsdelivr.net/gh/lit/dist@3/all/lit-all.min.js';
import { escapeRegExp, showConfirmation } from '../utils.js';
import { copyToClipboard, addImageURL } from '../app.js'; // Adjust based on your implementation
import { setCurrentPrompt, addToHistory } from '../slices/promptHistorySlice.js';
import { deleteImage } from '../slices/imagesSlice.js';

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
                            <button @click=${() => this.deleteImageByIndex(index)}>Delete</button>
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
            alert('Failed to copy: ' + err);
        });
    }

    openModal() {
        document.getElementById('image-modal').style.display = 'flex';
        document.getElementById('new-image-url').value = '';
        document.getElementById('new-image-url').focus();
    }

    addImageToPrompt(url) {
        let currentPrompt = this.store.getState().promptHistory.current_prompt;
        const newPrompt = currentPrompt ? `${url}, ${currentPrompt}` : url;
        this.store.dispatch(setCurrentPrompt(newPrompt.trim()));
        this.store.dispatch(addToHistory(newPrompt.trim()));
        this.updateUI();
    }

    deleteImageByIndex(index) {
        const state = this.store.getState();
        const image = state.images.images[index];
        if (image) {
            this.store.dispatch(deleteImage(index));
            let currentPrompt = state.promptHistory.current_prompt;
            if (image.url) {
                currentPrompt = this.removeImageFromPrompt(image.url, currentPrompt);
                this.store.dispatch(setCurrentPrompt(currentPrompt));
            }
            this.updateUI();
            showConfirmation("Image deleted successfully!");
        } else {
            console.error("Invalid image index");
        }
    }

    isImageInPrompt(url) {
        const currentPrompt = this.store.getState().promptHistory.current_prompt || '';
        return currentPrompt.includes(url);
    }

    toggleImage(url) {
        const state = this.store.getState();
        let currentPrompt = state.promptHistory.current_prompt || '';

        if (this.isImageInPrompt(url)) {
            currentPrompt = this.removeImageFromPrompt(url, currentPrompt);
        } else {
            currentPrompt = this.addImageToPrompt(url, currentPrompt);
        }

        this.store.dispatch(setCurrentPrompt(currentPrompt.trim()));
        this.store.dispatch(addToHistory(currentPrompt.trim()));
        this.updateUI();
        showConfirmation(`Image "${url}" toggled`);
    }

    addImageToPrompt(url, prompt) {
        return prompt ? `${prompt}, ${url}` : url;
    }

    removeImageFromPrompt(url, prompt) {
        const regex = new RegExp(`(,\\s*)?${this.escapeRegExp(url)}(,\\s*)?`, 'g');
        let newPrompt = prompt.replace(regex, ',');
        // Remove leading/trailing commas and whitespace
        newPrompt = newPrompt.replace(/^,\s*/, '').replace(/,\s*$/, '');
        return newPrompt;
    }

    escapeRegExp(string) {
        return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    }
}

export default PromptColumn;

// === END: web/voyage/js/components/promptColumn.js ===
```

**Key Changes:**

- **State Access:** Use `store.getState()` to access the current state.
- **Dispatching Actions:** Use `store.dispatch` to update the state.
- **Event Handlers:** Update event handlers to interact with Redux.

### 3. Update `optionsColumn.js`

```javascript
// === BEGIN: web/voyage/js/components/optionsColumn.js ===

import { html, render } from 'https://cdn.jsdelivr.net/gh/lit/dist@3/all/lit-all.min.js';
import { setAspectRatio, setModelVersion } from '../slices/optionsSlice.js';

class OptionsColumn {
    constructor(store, updateUI) {
        this.store = store;
        this.updateUI = updateUI;
        this.element = document.getElementById('options-column');
        this.init();
    }

    init() {
        this.element.addEventListener('change', (e) => {
            if (e.target.name === 'aspect-ratio') this.handleAspectRatioChange(e);
            if (e.target.id === 'model-version-select') this.handleModelVersionChange(e);
        });
    }

    render() {
        const state = this.store.getState();
        const options = state.options;

        const template = html`
            <h2>Options</h2>
            <div>
                <h3>Aspect Ratio</h3>
                <div id="aspect-ratio-options">
                    ${this.renderAspectRatioOptions(options.aspect_ratio)}
                </div>
            </div>
            <div>
                <h3>Model Version</h3>
                <select id="model-version-select" .value=${options.model_version}>
                    ${this.renderModelVersionOptions(options.model_version)}
                </select>
            </div>
        `;

        render(template, this.element);
    }

    renderAspectRatioOptions(selectedRatio) {
        const standardRatios = ['1:1', '16:9', '4:3'];
        return html`
            ${standardRatios.map(ratio => html`
                <label>
                    <input type="radio" name="aspect-ratio" value=${ratio} ?checked=${ratio === selectedRatio}>
                    ${ratio}
                </label><br>
            `)}
            ${!standardRatios.includes(selectedRatio) ? html`
                <label>
                    <input type="radio" name="aspect-ratio" value=${selectedRatio} checked>
                    ${selectedRatio} (Custom)
                </label>
            ` : ''}
        `;
    }

    renderModelVersionOptions(selectedVersion) {
        const standardVersions = ['7', '6', '5', '4', '3', '2'];
        return html`
            ${standardVersions.map(version => html`
                <option value=${version} ?selected=${version === selectedVersion}>v${version}</option>
            `)}
            ${!standardVersions.includes(selectedVersion) ? html`
                <option value=${selectedVersion} selected>${selectedVersion} (Custom)</option>
            ` : ''}
        `;
    }

    handleAspectRatioChange(event) {
        const ratio = event.target.value;
        this.store.dispatch(setAspectRatio(ratio));
        this.updateUI();
    }

    handleModelVersionChange(event) {
        const version = event.target.value;
        this.store.dispatch(setModelVersion(version));
        this.updateUI();
    }
}

export default OptionsColumn;

// === END: web/voyage/js/components/optionsColumn.js ===
```

**Key Changes:**

- **Dispatching Actions:** Use `setAspectRatio` and `setModelVersion` actions to update options.
- **State Access:** Access current options from the store.

### 4. Update `historySection.js`

```javascript
// === BEGIN: web/voyage/js/components/historySection.js ===

import { html, render } from 'https://cdn.jsdelivr.net/gh/lit/dist@3/all/lit-all.min.js';
import { highlightText } from '../utils.js';
import { setSearchQuery, setCurrentPrompt, addToHistory } from '../slices/promptHistorySlice.js';

class HistorySection {
    constructor(store, updateUI) {
        this.store = store;
        this.updateUI = updateUI;
        this.element = document.querySelector('.history');
        this.init();
    }

    init() {
        this.element.addEventListener('input', (e) => {
            if (e.target.id === 'search-history') this.searchHistory(e.target.value);
        });
    }

    render() {
        const state = this.store.getState();
        const query = state.promptHistory.search_query.toLowerCase();
        const filteredHistory = state.promptHistory.prompt_history.filter(prompt => prompt.toLowerCase().includes(query));

        const template = html`
            <h2>Prompt History</h2>
            <input type="text" id="search-history" class="search-input" placeholder="Search prompts" .value=${state.promptHistory.search_query}>
            <div id="history-list">
                ${filteredHistory.map(prompt => this.renderHistoryItem(prompt, query))}
            </div>
        `;

        render(template, this.element);
    }

    renderHistoryItem(prompt, query) {
        return html`
            <div class="list-item">
                <span @click=${() => this.loadPromptFromHistory(prompt)}
                      .innerHTML=${highlightText(prompt, query)}
                      style="cursor: pointer;">
                </span>
            </div>
        `;
    }

    searchHistory(query) {
        this.store.dispatch(setSearchQuery(query));
        this.render();
    }

    loadPromptFromHistory(prompt) {
        const { cleanPrompt, aspectRatio, modelVersion } = this.parsePromptOptions(prompt);
        
        this.store.dispatch(setCurrentPrompt(cleanPrompt));
        // Assuming you have actions to set aspect ratio and model version
        if (aspectRatio) {
            this.store.dispatch(setAspectRatio(aspectRatio));
        }
        if (modelVersion) {
            this.store.dispatch(setModelVersion(modelVersion));
        }
        this.store.dispatch(addToHistory(cleanPrompt));
        this.updateUI();
    }

    parsePromptOptions(prompt) {
        const options = {
            aspectRatio: null,
            modelVersion: null,
            cleanPrompt: prompt
        };

        const arMatch = prompt.match(/--ar\s+(\d+:\d+)/i);
        if (arMatch) {
            options.aspectRatio = arMatch[1];
            options.cleanPrompt = options.cleanPrompt.replace(arMatch[0], '').trim();
        }

        const vMatch = prompt.match(/--v\s+(\w+)/i);
        if (vMatch) {
            options.modelVersion = vMatch[1];
            options.cleanPrompt = options.cleanPrompt.replace(vMatch[0], '').trim();
        }

        return options;
    }
}

export default HistorySection;

// === END: web/voyage/js/components/historySection.js ===
```

**Key Changes:**

- **Dispatching Actions:** Use `setSearchQuery`, `setCurrentPrompt`, and `addToHistory` to update the state.
- **State Access:** Access prompt history and search query from the store.

---

## Handling State Persistence with LocalStorage

Previously, your custom `State` class handled persistence with `localStorage`. We'll replicate this behavior using Redux middleware.

### 1. Create `persistMiddleware.js`

Create a middleware to save the Redux state to `localStorage` whenever it changes.

```javascript
// === BEGIN: web/voyage/js/persistMiddleware.js ===

const persistMiddleware = store => next => action => {
    const result = next(action);
    const state = store.getState();
    localStorage.setItem('midjourneyPromptState', JSON.stringify(state));
    return result;
};

export default persistMiddleware;

// === END: web/voyage/js/persistMiddleware.js ===
```

### 2. Update `store.js` to Include Persistence

Modify your `store.js` to load the initial state from `localStorage` and apply the persistence middleware.

```javascript
// === BEGIN: web/voyage/js/store.js ===

const { configureStore } = window.RTK;

// Import slices
import { promptFragmentsSlice } from './slices/promptFragmentsSlice.js';
import { imagesSlice } from './slices/imagesSlice.js';
import { optionsSlice } from './slices/optionsSlice.js';
import { promptHistorySlice } from './slices/promptHistorySlice.js';
import { uiSlice } from './slices/uiSlice.js';

// Import persistence middleware
import persistMiddleware from './persistMiddleware.js';

// Function to load state from localStorage
function loadState() {
    try {
        const serializedState = localStorage.getItem('midjourneyPromptState');
        if (serializedState === null) {
            return undefined; // Let reducers initialize the state
        }
        return JSON.parse(serializedState);
    } catch (err) {
        console.error("Failed to load state from localStorage:", err);
        return undefined;
    }
}

// Load persisted state
const persistedState = loadState();

// Configure the Redux store with slices and middleware
const store = configureStore({
    reducer: {
        promptFragments: promptFragmentsSlice.reducer,
        images: imagesSlice.reducer,
        options: optionsSlice.reducer,
        promptHistory: promptHistorySlice.reducer,
        ui: uiSlice.reducer,
    },
    preloadedState: persistedState,
    middleware: (getDefaultMiddleware) => getDefaultMiddleware().concat(persistMiddleware),
    devTools: window.__REDUX_DEVTOOLS_EXTENSION__ ? window.__REDUX_DEVTOOLS_EXTENSION__() : false,
});

export default store;

// === END: web/voyage/js/store.js ===
```

**Key Changes:**

- **Load State:** Load the initial state from `localStorage`.
- **Persist Middleware:** Add a middleware that saves the state to `localStorage` on every action dispatch.

---

## Finalizing Import and Export Functionality

Previously, the `importState` function in `app.js` was not implemented due to Redux Toolkit's immutability. Here's how you can implement it:

### 1. Implement Import State Functionality

In `app.js`, modify the `importState` method to dispatch actions that replace each slice's state.

```javascript
// === BEGIN: web/voyage/js/app.js ===

// ... previous imports
import { replacePromptFragments } from './slices/promptFragmentsSlice.js';
import { replaceImages } from './slices/imagesSlice.js';
import { replaceOptions } from './slices/optionsSlice.js';
import { replacePromptHistory } from './slices/promptHistorySlice.js';

class App {
    // ... constructor and other methods

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
                    
                    // Dispatch actions to replace each slice's state
                    if (importedState.promptFragments) {
                        this.store.dispatch(replacePromptFragments(importedState.promptFragments));
                    }
                    if (importedState.images) {
                        this.store.dispatch(replaceImages(importedState.images));
                    }
                    if (importedState.options) {
                        this.store.dispatch(replaceOptions(importedState.options));
                    }
                    if (importedState.promptHistory) {
                        this.store.dispatch(replacePromptHistory(importedState.promptHistory));
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

export default App;

// === END: web/voyage/js/app.js ===
```

### 2. Define Replace Actions in Slices

For each slice, define a `replace` action to overwrite the current state.

#### a. `promptFragmentsSlice.js`

```javascript
// === BEGIN: web/voyage/js/slices/promptFragmentsSlice.js ===

const { createSlice } = window.RTK;

// ... existing code

export const promptFragmentsSlice = createSlice({
    name: 'promptFragments',
    initialState,
    reducers: {
        // ... existing reducers
        replacePromptFragments: (state, action) => {
            return action.payload;
        }
    }
});

export const {
    addFragment,
    deleteFragment,
    toggleCheckedFragment,
    unselectAllFragments,
    saveSelection,
    deleteSavedSelection,
    replacePromptFragments
} = promptFragmentsSlice.actions;

// === END: web/voyage/js/slices/promptFragmentsSlice.js ===
```

#### b. `imagesSlice.js`

```javascript
// === BEGIN: web/voyage/js/slices/imagesSlice.js ===

const { createSlice } = window.RTK;

// ... existing code

export const imagesSlice = createSlice({
    name: 'images',
    initialState,
    reducers: {
        // ... existing reducers
        replaceImages: (state, action) => {
            return action.payload;
        }
    }
});

export const { addImage, deleteImage, replaceImages } = imagesSlice.actions;

// === END: web/voyage/js/slices/imagesSlice.js ===
```

#### c. `optionsSlice.js`

```javascript
// === BEGIN: web/voyage/js/slices/optionsSlice.js ===

const { createSlice } = window.RTK;

// ... existing code

export const optionsSlice = createSlice({
    name: 'options',
    initialState,
    reducers: {
        // ... existing reducers
        replaceOptions: (state, action) => {
            return action.payload;
        }
    }
});

export const { setAspectRatio, setModelVersion, replaceOptions } = optionsSlice.actions;

// === END: web/voyage/js/slices/optionsSlice.js ===
```

#### d. `promptHistorySlice.js`

```javascript
// === BEGIN: web/voyage/js/slices/promptHistorySlice.js ===

const { createSlice } = window.RTK;

// ... existing code

export const promptHistorySlice = createSlice({
    name: 'promptHistory',
    initialState,
    reducers: {
        // ... existing reducers
        replacePromptHistory: (state, action) => {
            return action.payload;
        }
    }
});

export const { addToHistory, setSearchQuery, setCurrentPrompt, replacePromptHistory } = promptHistorySlice.actions;

// === END: web/voyage/js/slices/promptHistorySlice.js ===
```

**Note:** Ensure that all `replace` actions are properly exported and imported in `app.js`.

---
