const DEFAULT_STATE = {
    prompt_fragments: [
        "a majestic lion",
        "in a lush jungle",
        "with vibrant colors",
        "photorealistic style"
    ],
    images: [
        { url: "https://example.com/lion.jpg", thumbnail: "", alt: "Lion" },
        { url: "https://example.com/jungle.jpg", thumbnail: "", alt: "Jungle" }
    ],
    options: {
        aspect_ratio: "16:9",
        model_version: "v5"
    },
    current_prompt: "",
    prompt_history: [
        "a serene lake at sunset",
        "cyberpunk cityscape with neon lights",
        "abstract geometric patterns in pastel colors"
    ],
    search_query: "",
    checked_fragments: [],
    saved_selections: []
};

class State {
    constructor() {
        this.loadFromLocalStorage();
    }

    loadFromLocalStorage() {
        const savedState = localStorage.getItem('midjourneyPromptState');
        this.data = savedState ? JSON.parse(savedState) : {};
        
        // Ensure all default keys are present
        Object.keys(DEFAULT_STATE).forEach(key => {
            if (!(key in this.data)) {
                this.data[key] = DEFAULT_STATE[key];
            }
        });

        this.save();
    }

    get(key) {
        return this.data[key];
    }

    set(key, value) {
        this.data[key] = value;
        this.save();
    }

    save() {
        localStorage.setItem('midjourneyPromptState', JSON.stringify(this.data));
    }

    addToHistory(prompt) {
        if (prompt && prompt !== this.data.prompt_history[0]) {
            this.data.prompt_history.unshift(prompt);
            this.save();
        }
    }
}

export default State;