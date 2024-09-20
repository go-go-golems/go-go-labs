import { html, render } from 'https://cdn.jsdelivr.net/gh/lit/dist@3/all/lit-all.min.js';
import { highlightText, parsePromptOptions } from '../utils.js';

class HistorySection {
    constructor(state, updateUI) {
        this.state = state;
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
        const query = this.state.get('search_query').toLowerCase();
        const filteredHistory = this.state.get('prompt_history').filter(prompt => prompt.toLowerCase().includes(query));

        const template = html`
            <h2>Prompt History</h2>
            <input type="text" id="search-history" class="search-input" placeholder="Search prompts" .value=${this.state.get('search_query')}>
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
        this.state.set('search_query', query);
        this.render();
    }

    loadPromptFromHistory(prompt) {
        const parsedOptions = parsePromptOptions(prompt);
        
        this.state.set('current_prompt', parsedOptions.cleanPrompt);
        
        const options = this.state.get('options');
        options.aspect_ratio = parsedOptions.aspectRatio || options.aspect_ratio;
        options.model_version = parsedOptions.modelVersion || options.model_version;
        this.state.set('options', options);
        
        this.updateUI();
    }
}

export default HistorySection;