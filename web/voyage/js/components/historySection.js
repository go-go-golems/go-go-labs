import { html, render } from 'https://cdn.jsdelivr.net/gh/lit/dist@3/all/lit-all.min.js';
import { highlightText, parsePromptOptions } from '../utils.js';
import { setSearchQuery, setCurrentPrompt, addToHistory } from '../slices/promptHistorySlice.js';
import { setAspectRatio, setModelVersion } from '../slices/optionsSlice.js';

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
        const parsedOptions = parsePromptOptions(prompt);
        
        this.store.dispatch(setCurrentPrompt(parsedOptions.cleanPrompt));
        
        if (parsedOptions.aspectRatio) {
            this.store.dispatch(setAspectRatio(parsedOptions.aspectRatio));
        }
        if (parsedOptions.modelVersion) {
            this.store.dispatch(setModelVersion(parsedOptions.modelVersion));
        }
        
        this.store.dispatch(addToHistory(prompt));
        this.updateUI();
    }
}

export default HistorySection;