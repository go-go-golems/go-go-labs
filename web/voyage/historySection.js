import { highlightText } from './utils.js';

class HistorySection {
    constructor(state, updateUI) {
        this.state = state;
        this.updateUI = updateUI;
        this.element = document.querySelector('.history');
        this.init();
    }

    init() {
        this.element.querySelector('#search-history').addEventListener('input', (e) => this.searchHistory(e.target.value));
    }

    render() {
        this.updateHistoryList();
    }

    updateHistoryList() {
        const historyList = this.element.querySelector('#history-list');
        historyList.innerHTML = '';
        const query = this.state.get('search_query').toLowerCase();
        const filteredHistory = this.state.get('prompt_history').filter(prompt => prompt.toLowerCase().includes(query));
        filteredHistory.forEach((prompt) => {
            const div = document.createElement('div');
            div.className = 'list-item';
            const span = document.createElement('span');
            span.innerHTML = highlightText(prompt, query);
            span.style.cursor = 'pointer';
            span.addEventListener('click', () => this.loadPromptFromHistory(prompt));
            div.appendChild(span);
            historyList.appendChild(div);
        });
    }

    searchHistory(query) {
        this.state.set('search_query', query);
        this.updateHistoryList();
    }

    loadPromptFromHistory(prompt) {
        this.state.set('current_prompt', prompt);
        const arMatch = prompt.match(/--ar\s+(\d+:\d+)/i);
        const vMatch = prompt.match(/--v\s+(\w+)/i);
        const options = this.state.get('options');
        options.aspect_ratio = arMatch ? arMatch[1] : "16:9";
        options.model_version = vMatch ? vMatch[1] : "v5";
        this.state.set('options', options);
        this.updateUI();
    }
}

export default HistorySection;